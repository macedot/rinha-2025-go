package consumer

import (
	"context"
	"fmt"
	"rinha-2025-go/internal/config"
	"rinha-2025-go/internal/database"
	"rinha-2025-go/internal/models"
	"rinha-2025-go/internal/services"
	"sync"
	"time"

	"github.com/ohler55/ojg/oj"
	"github.com/valyala/fasthttp"
)

type Consumer struct {
	queue  *services.PaymentQueue
	client *services.HttpClient
	redis  *database.Redis
	health *Health
}

func NewConsumer(
	cfg *config.Config,
	redis *database.Redis,
	client *services.HttpClient,
	health *Health,
) *Consumer {
	return &Consumer{
		queue:  services.NewPaymentQueue(context.Background(), cfg.RedisSocket),
		client: client,
		redis:  redis,
		health: health,
	}
}

func (w *Consumer) Close() {
	w.queue.Close()
}

func (w *Consumer) ProcessQueue() {
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

func (w *Consumer) getCurrentInstance() *config.Service {
	for {
		if instance := w.health.GetActiveInstance(); instance != nil {
			return instance
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (w *Consumer) ProcessPayment(payment *models.Payment) error {
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

func (w *Consumer) forwardPayment(instance *config.Service, payment *models.Payment, payload []byte) error {
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
