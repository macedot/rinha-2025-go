package producer

import (
	"log"
	"sync"

	"github.com/macedot/rinha-2025-go/internal/producer/handler"
	"github.com/macedot/rinha-2025-go/pkg/client"
	"github.com/macedot/rinha-2025-go/pkg/server"
	"github.com/macedot/rinha-2025-go/pkg/util"
	"github.com/valyala/fasthttp"
)

func Run() error {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		socketIn := util.NewSocketFromEnv("SOCKET_PAYMENT_IN")
		socketOut := util.NewSocketFromEnv("SOCKET_PAYMENT_OUT")
		socketClient := client.NewSocketClient().Init(socketOut)
		log.Printf("Listen for payments on %s", socketIn)
		server.RunSocketServer(socketIn,
			func(ctx *fasthttp.RequestCtx) {
				handler.PaymentHandler(ctx, socketClient)
			})
	}()
	wg.Add(1)
	go func() {
		socketIn := util.NewSocketFromEnv("SOCKET_SUMMARY_IN")
		socketOut := util.NewSocketFromEnv("SOCKET_SUMMARY_OUT")
		socketClient := client.NewSocketClient().Init(socketOut)
		log.Printf("Listen for summary on %s", socketIn)
		server.RunSocketServer(socketIn,
			func(ctx *fasthttp.RequestCtx) {
				handler.SummaryHandler(ctx, socketClient)
			})
	}()
	wg.Wait()
	return nil
}
