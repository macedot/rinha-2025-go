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

const minWaitTime = 100 * time.Millisecond

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
	return &PaymentWorker{
		ctx:    context.Background(),
		config: cfg,
		queue:  NewPaymentQueue(20 * 1024),
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
		select {
		case <-w.ctx.Done():
			return
		default:
			payment := w.queue.Dequeue()
			if err := w.ProcessPayment(payment); err != nil {
				w.queue.Enqueue(payment)
			}
		}
	}
}

func (w *PaymentWorker) getCurrentInstance() *config.ServiceStatus {
	var activeInstance *config.ServiceStatus
	for {
		activeInstance = w.health.GetActiveInstance()
		if activeInstance != nil && activeInstance.Mode != config.None {
			return activeInstance
		}
		waitTime := max(5*time.Second-time.Since(activeInstance.LastUpdate), minWaitTime)
		time.Sleep(waitTime)
	}
}

func (w *PaymentWorker) ProcessPayment(payment *models.Payment) error {
	payload, err := oj.Marshal(payment)
	if err != nil {
		return err
	}
	activeInstance := w.getCurrentInstance()
	instance := &w.config.GetServices()[activeInstance.Mode]
	return w.forwardPayment(instance, payment, payload)
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
		res.Default, err = w.redis.GetSummary(&services[0], param)
		if err != nil {
			log.Println("GetSummary:0:", err.Error())
		}
	}()
	if len(services) > 1 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res.Fallback, err = w.redis.GetSummary(&services[1], param)
			if err != nil {
				log.Println("GetSummary:1:", err.Error())
			}
		}()
	}
	wg.Wait()
	log.Print("GetSummary:", time.Since(start))
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
		ts := float64(timeValue.UTC().UnixMilli())
		return strconv.FormatFloat(ts, 'f', -1, 64), nil
	}
	return param, fmt.Errorf("invalid end time format")
}

func (w *PaymentWorker) PurgePayments() error {
	var wg sync.WaitGroup
	services := w.config.GetServices()
	for _, service := range services {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := w.purgePaymentProcessor(&service)
			if err != nil {
				log.Print("purgePaymentProcessor:", err)
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := w.redis.FlushAll()
		if err != nil {
			log.Print("FlushAll:", err)
		}
	}()
	wg.Wait()
	return nil
}

func (w *PaymentWorker) purgePaymentProcessor(instance *config.Service) error {
	if _, _, err := fasthttp.Post(nil, instance.URL+"/admin/purge-payments", nil); err != nil {
		return err
	}
	return nil
}
