package worker

import (
	"context"
	"log"
	"time"

	"github.com/macedot/rinha-2025-go/internal/types"
	"github.com/macedot/rinha-2025-go/internal/worker/health"
	"github.com/macedot/rinha-2025-go/pkg/client"
	"github.com/macedot/rinha-2025-go/pkg/storage"
	"github.com/macedot/rinha-2025-go/pkg/util"
	"github.com/redis/go-redis/v9"
	"github.com/valyala/fasthttp"
)

type Consumer struct {
	Redis           *redis.Client
	Client          *client.HttpClient
	defaultChecker  *health.HealthManager
	fallbackChecker *health.HealthManager
}

func (c *Consumer) Init() {
	go c.defaultChecker.CheckAndUpdateHealth()
	go c.fallbackChecker.CheckAndUpdateHealth()
}

func (c *Consumer) ProcessQueue() error {
	ctx := context.Background()
	for {
		result, err := c.Redis.BLPop(ctx, 0*time.Second, "payment_queue").Result()
		if err != nil {
			log.Printf("Queue pop error: %v", err)
			time.Sleep(time.Second)
			continue
		}
		payload := []byte(result[1])
		c.processPayment(payload)
	}
}

func (c *Consumer) processPayment(payload []byte) {
	defaultHealth, _ := c.defaultChecker.GetHealth()
	fallbackHealth, _ := c.fallbackChecker.GetHealth()
	processor := selectProcessor(defaultHealth, fallbackHealth)
	if processor == "" {
		c.addToRetryQueue(payload)
		return
	}

	endpoint := "http://payment-processor-default:8080/payments"
	key := "default_payments"
	if processor != "default" {
		endpoint = "http://payment-processor-fallback:8080/payments"
		key = "fallback_payments"
	}

	var payment types.PaymentRequest
	_ = payment.UnmarshalJSON(payload)

	score := float64(payment.RequestedAt.UnixNano())
	amount := int64(payment.Amount * 100)

	status, _, err := c.Client.Post(endpoint, payload)
	if err != nil || status != fasthttp.StatusOK {
		log.Println("Failed to process payment:", status, err)
		c.markProcessorAsFailing(processor)
		c.addToRetryQueue(payload)
		return
	}
	err = c.Redis.ZAdd(context.Background(), key, redis.Z{
		Score:  score,
		Member: amount,
	}).Err()
	if err != nil {
		log.Fatalln("Failed to store payment:", err.Error())
	}
}

func (c *Consumer) markProcessorAsFailing(proc string) {
	if proc == "default" {
		c.defaultChecker.SaveHealthToRedis(true, 9999)
	} else {
		c.fallbackChecker.SaveHealthToRedis(true, 9999)
	}
}

func (c *Consumer) addToRetryQueue(payload []byte) {
	if err := c.Redis.RPush(context.Background(), "payment_queue", payload).Err(); err != nil {
		log.Fatalln("Failed to requeue payment:", err.Error())
	}
}

func Run() error {
	rdb := storage.NewRedisClient(util.GetEnv("REDIS_ADDR"))
	defer rdb.Close()
	client := client.NewHttpClient().Init()
	consumer := &Consumer{
		Redis:  rdb,
		Client: client,
		defaultChecker: &health.HealthManager{
			Redis:     rdb,
			Client:    client,
			Processor: "default",
			Endpoint:  "http://payment-processor-default:8080/payments/service-health",
		},
		fallbackChecker: &health.HealthManager{
			Redis:     rdb,
			Client:    client,
			Processor: "fallback",
			Endpoint:  "http://payment-processor-fallback:8080/payments/service-health",
		},
	}
	consumer.Init()
	return consumer.ProcessQueue()
}

const gragefulLag = 100

func selectProcessor(defaultHealth, fallbackHealth *types.ProcessorHealth) string {
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
