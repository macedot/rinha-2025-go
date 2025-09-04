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
	payment.Timestamp = time.Now().UTC()
	w.queue.Enqueue(payment)
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
	ticker := time.NewTicker(500 + time.Millisecond)
	defer ticker.Stop()
	for range ticker.C {
		activeInstance := w.health.GetActiveInstance()
		if activeInstance != nil {
			return activeInstance
		}
	}
	return nil
}

func (w *PaymentWorker) ProcessPayment(payment *models.Payment) error {
	payload, err := oj.Marshal(payment)
	if err != nil {
		return err
	}
	activeInstance := w.getCurrentInstance()
	return w.forwardPayment(activeInstance, payment, payload)
}

func (w *PaymentWorker) forwardPayment(instance *config.Service, payment *models.Payment, payload []byte) error {
	status, err := w.client.Post(instance.URL+"/payments", payload, instance)
	if err != nil || status < fasthttp.StatusOK || status >= fasthttp.StatusMultipleChoices {
		if status == fasthttp.StatusUnprocessableEntity {
			return nil
		}
		return fmt.Errorf("invalid status code: %d", status)
	}
	w.redis.SavePayment(instance, payment)
	return nil
}

func (w *PaymentWorker) GetSummary(from, to string) (*models.SummaryResponse, error) {
	param, err := processSummary(from, to)
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
		res.Default, err = w.redis.GetSummary(&services.Default, param)
		if err != nil {
			log.Println("GetSummary:D:", err.Error())
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		res.Fallback, err = w.redis.GetSummary(&services.Fallback, param)
		if err != nil {
			log.Println("GetSummary:F:", err.Error())
		}
	}()
	wg.Wait()
	log.Print("GetSummary:", time.Since(start))
	log.Println("QueueLength:", w.queue.Length())
	return &res, nil
}

func processSummary(from, to string) (*models.SummaryParam, error) {
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
		err := w.purgePaymentProcessor(&services.Default)
		if err != nil {
			log.Print("purgePaymentProcessor:", err)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := w.purgePaymentProcessor(&services.Fallback)
		if err != nil {
			log.Print("purgePaymentProcessor:", err)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := w.redis.FlushAll()
		if err != nil {
			log.Print("FlushAll:", err)
		}
	}()
	wg.Wait()
	log.Print("PurgePayments:", time.Since(start))
	return nil
}

func (w *PaymentWorker) purgePaymentProcessor(instance *config.Service) error {
	if _, _, err := fasthttp.Post(nil, instance.URL+"/admin/purge-payments", nil); err != nil {
		return err
	}
	return nil
}
