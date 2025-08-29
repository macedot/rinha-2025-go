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
	requestSocket := util.GetEnv("SOCKET_API")
	defer os.Remove(requestSocket)

	paymentSocket := util.GetEnv("SOCKET_PAYMENT")
	paymentClient := client.NewHttpClient().InitSocket(paymentSocket)

	summarySocket := util.GetEnv("SOCKET_SUMMARY")
	summaryClient := client.NewHttpClient().InitSocket(summarySocket)

	log.Printf("Listen on %s", requestSocket)
	return server.RunSocketServer(requestSocket,
		func(ctx *fasthttp.RequestCtx) {
			switch string(ctx.Path()) {
			case "/payments":
				handler.PaymentHandler(ctx, paymentClient)
			case "/payments-summary":
				handler.SummaryHandler(ctx, summaryClient)
			default:
				ctx.Error("Not Found - producer", fasthttp.StatusNotFound)
			}
		})
}
