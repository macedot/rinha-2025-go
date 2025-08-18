package storage

import (
	"log"

	"github.com/macedot/rinha-2025-go/internal/storage/handler"
	"github.com/macedot/rinha-2025-go/internal/storage/store"
	"github.com/macedot/rinha-2025-go/pkg/server"
	"github.com/macedot/rinha-2025-go/pkg/util"
	"github.com/valyala/fasthttp"
)

func Run() error {
	storageSocket := util.NewSocketFromEnv("STORAGE_SOCKET")
	paymentDB := store.NewPaymentDB()
	healthDB := store.NewHealthDB()
	requestHandler := func(ctx *fasthttp.RequestCtx) {
		path := string(ctx.Path())
		switch path {
		case "/payments/default":
			handler.PaymentDefaultHandler(ctx, paymentDB)
		case "/payments/fallback":
			handler.PaymentFallbackHandler(ctx, paymentDB)
		case "/payments-summary":
			handler.SummaryHandler(ctx, paymentDB)
		case "/purge-payments":
			handler.PurgePaymentsHandler(ctx, paymentDB)
		case "/health/default":
			handler.HealthDefaultHandler(ctx, healthDB)
		case "/health/fallback":
			handler.HealthFallbackHandler(ctx, healthDB)
		default:
			ctx.Error("Not Found", fasthttp.StatusNotFound)
		}
	}
	log.Printf("Listening on %s", storageSocket)
	return server.RunSocketServer(storageSocket, requestHandler)
}
