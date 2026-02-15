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

// BufferPool for JSON marshaling to reduce GC pressure
var BufferPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 0, 1024)
		return &b
	},
}

type PaymentWorker struct {
	ctx         context.Context
	config      *config.Config
	queue       *PaymentQueue
	client      *HttpClient
	redis       *database.Redis
	health      *Health
	paymentChan chan *models.Payment
}

func NewPaymentWorker(
	cfg *config.Config,
	redis *database.Redis,
	client *HttpClient,
	health *Health,
) *PaymentWorker {
	ctx := context.Background()
	return &PaymentWorker{
		ctx:         ctx,
		config:      cfg,
		queue:       NewPaymentQueue(ctx, redis),
		client:      client,
		redis:       redis,
		health:      health,
		paymentChan: make(chan *models.Payment, 1000),
	}
}

func (w *PaymentWorker) Close() {
	w.queue.Close()
}

func (w *PaymentWorker) EnqueuePayment(payment *models.Payment) {
	select {
	case w.paymentChan <- payment:
	default:
		w.queue.Enqueue(payment)
	}
}

func (w *PaymentWorker) ProcessQueue() {
	for payment := range w.paymentChan {
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
	activeInstance := w.getCurrentInstance()
	payment.Timestamp = time.Now().UTC()

	// Get buffer from pool for JSON marshaling
	bufPtr := BufferPool.Get().(*[]byte)
	defer BufferPool.Put(bufPtr)

	payload, err := oj.Marshal(payment, *bufPtr)
	if err != nil {
		return fmt.Errorf("failed to marshal payment: %w", err)
	}

	return w.forwardPayment(activeInstance, payment, payload)
}

func (w *PaymentWorker) forwardPayment(instance *config.Service, payment *models.Payment, payload []byte) error {
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
	}()
	wg.Wait()
	return nil
}

func (w *PaymentWorker) purgePaymentProcessor(instance *config.Service) error {
	if _, _, err := fasthttp.Post(nil, instance.URL+"/admin/purge-payments", nil); err != nil {
		log.Print("purgePaymentProcessor:ERROR:", instance.Table, "|", err)
		return err
	}
	return nil
}
