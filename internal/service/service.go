package service

import (
	"context"
	"log"

	"github.com/macedot/rinha-2025-go/internal/service/handler"
	"github.com/macedot/rinha-2025-go/internal/service/health"
	"github.com/macedot/rinha-2025-go/pkg/client"
	"github.com/macedot/rinha-2025-go/pkg/server"
	"github.com/macedot/rinha-2025-go/pkg/util"
	"github.com/valyala/fasthttp"
)

func Run() error {
	serverSocket := util.NewSocketFromEnv("SERVICE_SOCKET")
	storageSocket := util.GetEnv("STORAGE_SOCKET")
	ctx := context.Background()
	defaultChecker := &health.HealthManager{
		Client:       client.NewHttpClient().Init(),
		Storage:      client.NewSocketClient().Init(storageSocket),
		HealthClient: "http://payment-processor-default:8080/payments/service-health",
		HealthStore:  "http://storage/health/default",
		Processor:    "default",
	}
	go defaultChecker.CheckAndUpdateHealth()
	fallbackChecker := &health.HealthManager{
		Client:       client.NewHttpClient().Init(),
		Storage:      client.NewSocketClient().Init(storageSocket),
		HealthClient: "http://payment-processor-fallback:8080/payments/service-health",
		HealthStore:  "http://storage/health/fallback",
		Processor:    "fallback",
	}
	go fallbackChecker.CheckAndUpdateHealth()
	go handler.StartRetryWorker(ctx, defaultChecker, fallbackChecker)
	storageClient := client.NewSocketClient().Init(storageSocket)
	requestHandler := func(ctx *fasthttp.RequestCtx) {
		path := string(ctx.Path())
		switch path {
		case "/payments":
			handler.PaymentHandler(ctx, defaultChecker, fallbackChecker)
		case "/payments-summary":
			handler.SummaryHandler(ctx, storageClient)
		case "/purge-payments":
			handler.PurgePaymentsHandler(ctx, storageClient)
		case "/health":
			handler.HealthHandler(ctx, defaultChecker, fallbackChecker)
		default:
			ctx.Error("Not Found", fasthttp.StatusNotFound)
		}
	}
	log.Printf("Listening on %s", serverSocket)
	return server.RunSocketServer(serverSocket, requestHandler)
}
