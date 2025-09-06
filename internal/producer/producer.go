package producer

import (
	"context"
	"fmt"
	"log"
	"rinha-2025-go/internal/config"
	"rinha-2025-go/internal/database"
	"rinha-2025-go/internal/models"
	"rinha-2025-go/internal/services"
	"strconv"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

type Producer struct {
	config *config.Config
	queue  *services.PaymentQueue
	redis  *database.Redis
}

func NewProducer(cfg *config.Config, redis *database.Redis) *Producer {
	return &Producer{
		config: cfg,
		queue:  services.NewPaymentQueue(context.Background(), cfg.RedisSocket),
		redis:  redis,
	}
}

func (w *Producer) Close() {
	w.queue.Close()
}

func (w *Producer) EnqueuePayment(payment *models.Payment) {
	w.queue.Enqueue(payment)
}

func (w *Producer) GetSummary(from, to string) (*models.SummaryResponse, error) {
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

func (w *Producer) PurgePayments() error {
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
	}()
	wg.Wait()
	log.Print("PurgePayments:", time.Since(start))
	return nil
}

func (w *Producer) purgePaymentProcessor(instance *config.Service) error {
	if _, _, err := fasthttp.Post(nil, instance.URL+"/admin/purge-payments", nil); err != nil {
		log.Print("purgePaymentProcessor:ERROR:", instance.Table, "|", err)
		return err
	}
	return nil
}
