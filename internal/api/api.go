package api

import (
	"log"

	"github.com/macedot/rinha-2025-go/internal/api/handler"
	"github.com/macedot/rinha-2025-go/pkg/server"
	"github.com/macedot/rinha-2025-go/pkg/storage"
	"github.com/macedot/rinha-2025-go/pkg/util"
	"github.com/valyala/fasthttp"
)

func Run() error {
	rdb := storage.NewRedisClient(util.GetEnv("REDIS_ADDR"))
	defer rdb.Close()
	serverSocket := util.NewSocketFromEnv("SERVICE_SOCKET")
	requestHandler := func(ctx *fasthttp.RequestCtx) {
		path := string(ctx.Path())
		switch path {
		case "/payments":
			handler.PaymentHandler(ctx, rdb)
		case "/payments-summary":
			handler.SummaryHandler(ctx, rdb)
		case "/purge-payments":
			handler.PurgePaymentsHandler(ctx, rdb)
		default:
			ctx.Error("Not Found", fasthttp.StatusNotFound)
		}
	}
	log.Printf("Listening on %s", serverSocket)
	return server.RunSocketServer(serverSocket, requestHandler)
}
