package handler

import (
	"context"
	"encoding/json"
	"log"
	"rinha-api/client"
	"rinha-api/health"
	"time"

	"github.com/valyala/fasthttp"
)

type PaymentJob struct {
	CorrelationID string
	Amount        float64
	RequestedAt   time.Time
	Attempt       int
}

var retryChannel = make(chan PaymentJob, 100000)

// var (
// 	retryQueue []PaymentJob
// 	retryMu    sync.Mutex
// )

const MaxAttempts = 10

func AddToRetryQueue(job PaymentJob) {
	// retryMu.Lock()
	// defer retryMu.Unlock()
	// retryQueue = append(retryQueue, job)
	retryChannel <- job
}

func PaymentHandler(defaultChecker, fallbackChecker *health.HealthManager) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		var req struct {
			CorrelationID string  `json:"correlationId"`
			Amount        float64 `json:"amount"`
		}
		json.Unmarshal(ctx.PostBody(), &req)
		job := PaymentJob{
			CorrelationID: req.CorrelationID,
			Amount:        req.Amount,
			RequestedAt:   time.Now().UTC(),
			Attempt:       0,
		}
		go ProcessPayment(job, defaultChecker, fallbackChecker)
		ctx.SetStatusCode(fasthttp.StatusAccepted)
	}
}

// const (
// 	gracefulLagMs = 100 // Allow default to be up to 100ms slower than fallback
// )

func SelectProcessor(defaultHealth, _ *health.ProcessorHealth) string {
	return "default"
}

// Select processor by time with graceful lag
// func SelectProcessor(defaultHealth, fallbackHealth *health.ProcessorHealth) string {
// 	if defaultHealth == nil && fallbackHealth == nil {
// 		return ""
// 	}
// 	if defaultHealth.Failing && fallbackHealth.Failing {
// 		return ""
// 	}
// 	if !defaultHealth.Failing && fallbackHealth.Failing {
// 		return "default"
// 	}
// 	if defaultHealth.Failing && !fallbackHealth.Failing {
// 		return "fallback"
// 	}
// 	if defaultHealth.MinResponseTime <= fallbackHealth.MinResponseTime+gracefulLagMs {
// 		return "default"
// 	}
// 	return "fallback"
// }

// Select processor by time
// func SelectProcessor(defaultHealth, fallbackHealth *health.ProcessorHealth) string {
// 	if defaultHealth == nil && fallbackHealth == nil {
// 		return ""
// 	}
// 	if defaultHealth.Failing && fallbackHealth.Failing {
// 		return ""
// 	}
// 	if !defaultHealth.Failing && fallbackHealth.Failing {
// 		return "default"
// 	}
// 	if defaultHealth.Failing && !fallbackHealth.Failing {
// 		return "fallback"
// 	}
// 	if defaultHealth.MinResponseTime <= fallbackHealth.MinResponseTime {
// 		return "default"
// 	}
// 	return "fallback"
// }

// Select default if on
// func SelectProcessor(defaultHealth, fallbackHealth *health.ProcessorHealth) string {
// 	if defaultHealth == nil && fallbackHealth == nil {
// 		return ""
// 	}
// 	if defaultHealth.Failing && fallbackHealth.Failing {
// 		return ""
// 	}
// 	if !defaultHealth.Failing {
// 		return "default"
// 	}
// 	return "fallback"
// }

func markProcessorAsFailing(proc string, d *health.HealthManager, f *health.HealthManager) {
	// if proc == "default" {
	// 	d.SaveHealthToRedis(true, 9999)
	// } else {
	// 	f.SaveHealthToRedis(true, 9999)
	// }
}

func ProcessPayment(job PaymentJob, defaultChecker, fallbackChecker *health.HealthManager) bool {
	// defaultHealth, _ := defaultChecker.GetHealth()
	// fallbackHealth, _ := fallbackChecker.GetHealth()

	defaultHealth := health.ProcessorHealth{
		Failing:         false,
		MinResponseTime: 0,
		LastChecked:     time.Now().UTC(),
	}
	fallbackHealth := health.ProcessorHealth{
		Failing:         true,
		MinResponseTime: 9999,
		LastChecked:     time.Now().UTC(),
	}

	processor := SelectProcessor(&defaultHealth, &fallbackHealth)
	if processor == "" {
		AddToRetryQueue(job)
		return false
	}

	socketClient := defaultChecker.Storage
	httpClient := defaultChecker.Client
	endpoint := "http://payment-processor-default:8080/payments"
	if processor != "default" {
		socketClient = fallbackChecker.Storage
		httpClient = fallbackChecker.Client
		endpoint = "http://payment-processor-fallback:8080/payments"
	}

	payload := map[string]any{
		"correlationId": job.CorrelationID,
		"amount":        job.Amount,
		"requestedAt":   job.RequestedAt,
	}

	body, _ := json.Marshal(payload)
	status, _, err := httpClient.PostTimeout(endpoint, body, 6*time.Second)
	if err != nil || status >= fasthttp.StatusInternalServerError {
		markProcessorAsFailing(processor, defaultChecker, fallbackChecker)
		job.Attempt++
		if job.Attempt < MaxAttempts {
			AddToRetryQueue(job)
		}
		return false
	}

	SaveToDB(job, processor, socketClient)
	return true
}

func SaveToDB(job PaymentJob, processor string, storageClient *client.SocketClient) {
	payload := map[string]any{
		"amount":      job.Amount,
		"serverType":  processor,
		"requestedAt": job.RequestedAt,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[DB] Error marshaling payload: %v", err)
		return
	}
	_, err = storageClient.Post("http://storage/payments", body)
	if err != nil {
		log.Printf("[DB] Error sending request: %v", err)
		return
	}
}

func StartRetryWorker(ctx context.Context, defaultChecker, fallbackChecker *health.HealthManager) {
	const workerCount = 8
	// for range workerCount {
	// 	go func() {
	// 		for {
	// 			retryMu.Lock()
	// 			if len(retryQueue) == 0 {
	// 				retryMu.Unlock()
	// 				time.Sleep(1 * time.Millisecond)
	// 				continue
	// 			}
	// 			job := retryQueue[0]
	// 			retryQueue = retryQueue[1:]
	// 			retryMu.Unlock()
	// 			ProcessPayment(job, defaultChecker, fallbackChecker)
	// 		}
	// 	}()
	// }
	for range workerCount {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case job := <-retryChannel:
					ProcessPayment(job, defaultChecker, fallbackChecker)
				}
			}
		}()
	}
}
