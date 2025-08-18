package handler

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	"github.com/macedot/rinha-2025-go/internal/service/health"
	"github.com/macedot/rinha-2025-go/internal/types"
	"github.com/valyala/fasthttp"
)

type PaymentJob struct {
	Payment types.PaymentRequest
	Attempt int
}

var retryChannel = make(chan *PaymentJob, 10000)

var queueSize atomic.Int64

const MaxAttempts = 10

func AddToRetryQueue(job *PaymentJob) {
	retryChannel <- job
}

func PaymentHandler(ctx *fasthttp.RequestCtx, defaultChecker *health.HealthManager, fallbackChecker *health.HealthManager) {
	go ProcessPayment(ctx.PostBody(), defaultChecker, fallbackChecker)
	queueSize.Add(1)
	ctx.SetStatusCode(fasthttp.StatusAccepted)
}

func ProcessPayment(payload []byte, defaultChecker *health.HealthManager, fallbackChecker *health.HealthManager) {
	var job PaymentJob
	job.Payment.UnmarshalJSON(payload)
	job.Payment.RequestedAt = time.Now().UTC()
	go ProcessPaymentJob(&job, defaultChecker, fallbackChecker)
}

func markProcessorAsFailing(proc string, d *health.HealthManager, f *health.HealthManager) {
	if proc == "default" {
		d.SaveHealth(true, 9999)
	} else {
		f.SaveHealth(true, 9999)
	}
}

const gragefulLag = 150

// Select processor
func SelectProcessor(defaultHealth, fallbackHealth *types.ProcessorHealth) string {
	useDefault := (defaultHealth != nil && !defaultHealth.Failing)
	useFallback := (fallbackHealth != nil && !fallbackHealth.Failing)
	if !useDefault {
		if useFallback {
			return "fallback"
		}
		return ""
	}
	if !useFallback {
		return "default"
	}
	if defaultHealth.MinResponseTime <= (fallbackHealth.MinResponseTime + gragefulLag) {
		return "default"
	}
	if 3*defaultHealth.MinResponseTime <= fallbackHealth.MinResponseTime {
		return "default"
	}
	return "fallback"
}

func ProcessPaymentJob(job *PaymentJob, defaultChecker *health.HealthManager, fallbackChecker *health.HealthManager) bool {
	defaultHealth := defaultChecker.GetHealthStorage()
	fallbackHealth := fallbackChecker.GetHealthStorage()
	processor := SelectProcessor(defaultHealth, fallbackHealth)
	if processor == "" {
		AddToRetryQueue(job)
		return false
	}

	clientEndpoint := "http://payment-processor-default:8080/payments"
	storageEndpoint := "http://storage/payments/default"
	client := defaultChecker.Client
	storage := defaultChecker.Storage
	if processor != "default" {
		clientEndpoint = "http://payment-processor-fallback:8080/payments"
		storageEndpoint = "http://storage/payments/fallback"
		client = fallbackChecker.Client
		storage = fallbackChecker.Storage
	}

	body, _ := job.Payment.MarshalJSON()
	status, _, err := client.PostTimeout(clientEndpoint, body, 6*time.Second)
	if err != nil || status >= 500 {
		markProcessorAsFailing(processor, defaultChecker, fallbackChecker)
		job.Attempt++
		if job.Attempt < MaxAttempts {
			AddToRetryQueue(job)
		}
		return false
	}

	// SaveToDB
	_, err = storage.Post(storageEndpoint, body)
	if err != nil {
		log.Println("SaveToStorage:", storageEndpoint, err.Error())
	} else {
		queueSize.Add(-1)
		log.Println("SaveToStorage:", storageEndpoint, queueSize.Load())
	}

	return true
}

func StartRetryWorker(ctx context.Context, defaultChecker *health.HealthManager, fallbackChecker *health.HealthManager) {
	const workerCount = 8
	for range workerCount {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case job := <-retryChannel:
					ProcessPaymentJob(job, defaultChecker, fallbackChecker)
				}
			}
		}()
	}
}
