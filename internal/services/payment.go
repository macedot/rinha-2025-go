package services

import (
	"context"
	"fmt"
	"log"
	"rinha-2025-go/internal/config"
	"rinha-2025-go/internal/database"
	"rinha-2025-go/internal/models"
	"strconv"
	"sync"
	"time"

	"github.com/ohler55/ojg/oj"
	"github.com/valyala/fasthttp"
)

type PaymentWorker struct {
	ctx    context.Context
	config *config.Config
	queue  *PaymentQueue
	client *HttpClient
	redis  *database.Redis
	health *Health
}

func NewPaymentWorker(
	cfg *config.Config,
	redis *database.Redis,
	client *HttpClient,
	health *Health,
) *PaymentWorker {
	ctx := context.Background()
	return &PaymentWorker{
		ctx:    ctx,
		config: cfg,
		queue:  NewPaymentQueue(ctx, cfg.RedisSocket),
		client: client,
		redis:  redis,
		health: health,
	}
}

func (w *PaymentWorker) Close() {
	w.queue.Close()
}

func (w *PaymentWorker) EnqueuePayment(payment *models.Payment) {
	// w.redis.AddStat("EnqueuePayment")
	go w.queue.Enqueue(payment)
}

func (w *PaymentWorker) ProcessQueue() {
	for {
		payment := w.queue.Dequeue()
		if payment == nil {
			time.Sleep(time.Second)
			continue
		}
		if err := w.ProcessPayment(payment); err != nil {
			w.queue.Enqueue(payment)
		}
	}
}

func (w *PaymentWorker) getCurrentInstance() *config.Service {
	for {
		if instance := w.health.GetActiveInstance(); instance != nil {
			return instance
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (w *PaymentWorker) ProcessPayment(payment *models.Payment) error {
	// w.redis.AddStat("ProcessPayment")
	var wg sync.WaitGroup
	var activeInstance *config.Service
	wg.Add(1)
	go func() {
		defer wg.Done()
		activeInstance = w.getCurrentInstance()
	}()
	var payload []byte
	wg.Add(1)
	go func() {
		defer wg.Done()
		payment.Timestamp = time.Now().UTC()
		payload, _ = oj.Marshal(payment)
	}()
	wg.Wait()
	return w.forwardPayment(activeInstance, payment, payload)
}

func (w *PaymentWorker) forwardPayment(instance *config.Service, payment *models.Payment, payload []byte) error {
	// w.redis.AddStat("forwardPayment")
	status, err := w.client.Post(instance.URL+"/payments", payload, instance)
	if err != nil || status < fasthttp.StatusOK || status >= fasthttp.StatusMultipleChoices {
		if status == fasthttp.StatusUnprocessableEntity {
			return nil
		}
		if status == 0 || status == fasthttp.StatusInternalServerError {
			time.Sleep(time.Second)
		}
		return fmt.Errorf("invalid status code: %d", status)
	}
	if err := w.redis.SavePayment(instance, payment); err != nil {
		return fmt.Errorf("failed to save payment: %w", err)
	}
	// w.redis.AddStat("SavePayment")
	return nil
}

func (w *PaymentWorker) GetSummary(from, to string) (*models.SummaryResponse, error) {
	param, err := processSummaryParam(from, to)
	if err != nil {
		return nil, err
	}
	services := w.config.GetServices()
	var res models.SummaryResponse
	var wg sync.WaitGroup
	wg.Add(1)
	start := time.Now()
	go func() {
		defer wg.Done()
		res.Default = w.redis.GetSummary(&services.Default, param)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		res.Fallback = w.redis.GetSummary(&services.Fallback, param)
	}()
	wg.Wait()
	log.Print("GetSummary:", time.Since(start))
	log.Println("QueueLength:", w.queue.Length())
	// log.Println("PaymentsRequested:", w.redis.GetStat("EnqueuePayment"))
	// log.Println("PaymentsProcessed:", w.redis.GetStat("ProcessPayment"))
	// log.Println("PaymentsForwarded:", w.redis.GetStat("forwardPayment"))
	// log.Println("PaymentsSaved:", w.redis.GetStat("SavePayment"))
	return &res, nil
}

func processSummaryParam(from, to string) (*models.SummaryParam, error) {
	var res models.SummaryParam
	var err error
	if res.StartTime, err = processTime(from, "-inf"); err != nil {
		return nil, fmt.Errorf("invalid start time format")
	}
	if res.EndTime, err = processTime(to, "+inf"); err != nil {
		return nil, fmt.Errorf("invalid end time format")
	}
	return &res, nil
}

func processTime(param string, value string) (string, error) {
	if param == "" {
		return value, nil
	}
	if timeValue, err := time.Parse(time.RFC3339, param); err == nil {
		ts := float64(timeValue.UTC().UnixNano()) / 1e9
		return strconv.FormatFloat(ts, 'f', -1, 64), nil
	}
	return param, fmt.Errorf("invalid end time format")
}

func (w *PaymentWorker) PurgePayments() error {
	var wg sync.WaitGroup
	start := time.Now()
	services := w.config.GetServices()
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.purgePaymentProcessor(&services.Default)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.purgePaymentProcessor(&services.Fallback)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.redis.FlushAll()
		// w.redis.ResetStat("EnqueuePayment")
		// w.redis.ResetStat("ProcessPayment")
		// w.redis.ResetStat("forwardPayment")
		// w.redis.ResetStat("SavePayment")
	}()
	wg.Wait()
	log.Print("PurgePayments:", time.Since(start))
	return nil
}

func (w *PaymentWorker) purgePaymentProcessor(instance *config.Service) error {
	if _, _, err := fasthttp.Post(nil, instance.URL+"/admin/purge-payments", nil); err != nil {
		log.Print("purgePaymentProcessor:ERROR:", instance.Table, "|", err)
		return err
	}
	return nil
}

// func (w *PaymentWorker) UpdateServicesFee() error {
// 	var wg sync.WaitGroup
// 	start := time.Now()
// 	services := w.config.GetServices()
// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		summary := w.updateServiceFee(&services.Default)
// 		services.Default.Fee = summary.FeePerTransaction
// 	}()
// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		summary := w.updateServiceFee(&services.Fallback)
// 		services.Fallback.Fee = summary.FeePerTransaction
// 	}()
// 	wg.Wait()
// 	log.Print("UpdateServicesFee:", time.Since(start))
// 	return nil
// }

// func (w *PaymentWorker) updateServiceFee(instance *config.Service) *models.PaymentSummary {
// 	var summary models.PaymentSummary
// 	statusCode, body, err := w.client.Get(instance.URL+"/admin/payments-summary", instance)
// 	if err != nil || statusCode != fasthttp.StatusOK {
// 		log.Print("updateServiceFee:Get:", instance.Table, ":", statusCode, "|", err)
// 		return &summary
// 	}
// 	if err := oj.Unmarshal(body, &summary); err != nil {
// 		log.Print("updateServiceFee:Unmarshal:", instance.Table, ":", err)
// 		return &summary
// 	}
// 	return &summary
// }
