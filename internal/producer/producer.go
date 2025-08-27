package producer

import (
	"log"
	"os"

	"github.com/macedot/rinha-2025-go/internal/producer/handler"
	"github.com/macedot/rinha-2025-go/pkg/client"
	"github.com/macedot/rinha-2025-go/pkg/server"
	"github.com/macedot/rinha-2025-go/pkg/util"
	"github.com/valyala/fasthttp"
)

func Run() error {
	requestSocket := util.NewSocketFromEnv("SOCKET_API")
	defer os.Remove(requestSocket)

	paymentSocket := util.NewSocketFromEnv("SOCKET_PAYMENT")
	paymentClient := client.NewSocketClient().Init(paymentSocket)

	summarySocket := util.NewSocketFromEnv("SOCKET_SUMMARY")
	summaryClient := client.NewSocketClient().Init(summarySocket)

	log.Printf("Listen on %s", requestSocket)
	return server.RunSocketServer(requestSocket,
		func(ctx *fasthttp.RequestCtx) {
			switch string(ctx.Path()) {
			case "/payments":
				handler.PaymentHandler(ctx, paymentClient)
			case "/payments-summary":
				handler.SummaryHandler(ctx, summaryClient)
			default:
				ctx.Error("Not Found", fasthttp.StatusNotFound)
			}
		})
}
