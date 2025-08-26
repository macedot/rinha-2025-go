package consumer

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/macedot/rinha-2025-go/internal/consumer/cache"
	"github.com/macedot/rinha-2025-go/internal/consumer/health"
	"github.com/macedot/rinha-2025-go/internal/types"
	"github.com/macedot/rinha-2025-go/pkg/client"
	"github.com/macedot/rinha-2025-go/pkg/server"
	"github.com/macedot/rinha-2025-go/pkg/util"
	"github.com/valyala/fasthttp"
)

var paymentChannel = make(chan []byte, 10000)

type Consumer struct {
	Client          *client.HttpClient
	defaultChecker  *health.HealthManager
	fallbackChecker *health.HealthManager
	Payments        *cache.MemoryDB
}

func NewConsumer() *Consumer {
	client := client.NewHttpClient().Init()
	return &Consumer{
		Client: client,
		defaultChecker: health.NewHealthManager(
			client,
			"default",
			"http://payment-processor-default:8080/payments/service-health",
		),
		fallbackChecker: health.NewHealthManager(
			client,
			"fallback",
			"http://payment-processor-fallback:8080/payments/service-health",
		),
		Payments: cache.NewMemoryDB(),
	}
}

func (c *Consumer) Init() *Consumer {
	go c.defaultChecker.CheckAndUpdateHealth()
	go c.fallbackChecker.CheckAndUpdateHealth()
	return c
}

func (c *Consumer) ProcessQueue(ctx context.Context) *Consumer {
	for {
		select {
		case <-ctx.Done():
			return c
		case job := <-paymentChannel:
			c.processPayment(job)
		}
	}
}

func (c *Consumer) processPayment(payload []byte) {
	defaultHealth := c.defaultChecker.GetHealth()
	fallbackHealth := c.fallbackChecker.GetHealth()
	processor := selectProcessor(defaultHealth, fallbackHealth)
	if processor == "" {
		c.addToRetryQueue(payload)
		return
	}

	endpoint := "http://payment-processor-default:8080/payments"
	//key := "default_payments"
	if processor != "default" {
		endpoint = "http://payment-processor-fallback:8080/payments"
		//	key = "fallback_payments"
	}

	var payment types.PaymentRequest
	_ = payment.UnmarshalJSON(payload)

	status, _, err := c.Client.Post(endpoint, payload)
	if err != nil || status != fasthttp.StatusOK {
		log.Println("Failed to process payment:", status, err)
		c.markProcessorAsFailing(processor)
		c.addToRetryQueue(payload)
		return
	}
	// err = c.Redis.ZAdd(context.Background(), key, redis.Z{
	// 	Score:  float64(payment.RequestedAt.UnixNano()),
	// 	Member: payload,
	// }).Err()
}

func (c *Consumer) markProcessorAsFailing(proc string) {
	if proc == "default" {
		c.defaultChecker.SaveHealth(true, 9999)
	} else {
		c.fallbackChecker.SaveHealth(true, 9999)
	}
}

func (c *Consumer) addToRetryQueue(payload []byte) {
	paymentChannel <- payload
}

func (c *Consumer) GetSummary(from, to time.Time) []byte {
	response := c.Payments.QuerySummary(from, to)
	body, _ := response.MarshalJSON()
	return body
}

func Run() error {
	var wg sync.WaitGroup

	consumer := NewConsumer().Init()

	wg.Add(1)
	go func() {
		defer wg.Done()
		consumer.ProcessQueue(context.Background())
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ServerFromSocket("SOCKET_PAYMENT_IN",
			func(ctx *fasthttp.RequestCtx) {
				paymentChannel <- ctx.PostBody()
				ctx.SetStatusCode(fasthttp.StatusAccepted)
			})
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ServerFromSocket("SOCKET_SUMMARY_IN",
			func(ctx *fasthttp.RequestCtx) {
				fromStr := string(ctx.QueryArgs().Peek("from"))
				toStr := string(ctx.QueryArgs().Peek("to"))
				from := time.Unix(0, 0).UTC()
				to := time.Now().UTC()
				var err error
				if fromStr != "" {
					from, err = time.Parse(time.RFC3339, fromStr)
					if err != nil {
						ctx.Error("Invalid 'from' timestamp", fasthttp.StatusBadRequest)
						return
					}
				}
				if toStr != "" {
					to, err = time.Parse(time.RFC3339, toStr)
					if err != nil {
						ctx.Error("Invalid 'to' timestamp", fasthttp.StatusBadRequest)
						return
					}
				}
				if from.After(to) {
					ctx.Error("Invalid 'from' and 'to' timestamps", fasthttp.StatusBadRequest)
					return
				}
				body := consumer.GetSummary(from, to)
				ctx.SetStatusCode(fasthttp.StatusOK)
				ctx.SetContentType("application/json")
				ctx.SetBody(body)
				ctx.SetStatusCode(fasthttp.StatusOK)
			})
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ServerFromSocket("SOCKET_PURGE_IN",
			func(ctx *fasthttp.RequestCtx) {
				ctx.SetStatusCode(fasthttp.StatusOK)
			})
	}()

	wg.Wait()
	return nil
}

func ServerFromSocket(envVar string, handler func(ctx *fasthttp.RequestCtx)) error {
	socket := util.NewSocketFromEnv(envVar)
	log.Printf("Listen on %s", socket)
	return server.RunSocketServer(socket, handler)
}

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
	if 3*defaultHealth.MinResponseTime <= fallbackHealth.MinResponseTime {
		return "default"
	}
	return "fallback"
}
