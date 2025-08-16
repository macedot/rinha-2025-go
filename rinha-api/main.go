package main

import (
	"context"
	"log"
	"os"
	"rinha-api/client"
	"rinha-api/handler"
	"rinha-api/health"
	"strings"

	"github.com/valyala/fasthttp"
)

func main() {
	serverSocket := os.Getenv("SERVER_SOCKET")
	if serverSocket == "" {
		log.Fatalln("SERVER_SOCKET environment variable not set")
	}
	client.NewSocket(serverSocket)

	storageSocket := os.Getenv("STORAGE_SOCKET")
	if storageSocket == "" {
		log.Fatalln("STORAGE_SOCKET environment variable not set")
	}

	storageClient := client.NewSocketClient().Init(storageSocket)

	ctx := context.Background()

	defaultChecker := &health.HealthManager{
		Storage:   client.NewSocketClient().Init(storageSocket),
		Client:    client.NewHttpClient().Init(),
		Processor: "default",
		Endpoint:  "http://payment-processor-default:8080/payments/service-health",
		Ctx:       ctx,
	}
	go defaultChecker.CheckAndUpdateHealth()

	fallbackChecker := &health.HealthManager{
		Storage:   client.NewSocketClient().Init(storageSocket),
		Client:    client.NewHttpClient().Init(),
		Processor: "fallback",
		Endpoint:  "http://payment-processor-fallback:8080/payments/service-health",
		Ctx:       ctx,
	}
	go fallbackChecker.CheckAndUpdateHealth()

	go handler.StartRetryWorker(ctx, defaultChecker, fallbackChecker)

	requestHandler := func(ctx *fasthttp.RequestCtx) {
		path := string(ctx.Path())
		switch {
		case ctx.IsPost() && strings.HasPrefix(path, "/payments"):
			handler.PaymentHandler(defaultChecker, fallbackChecker)(ctx)
		case ctx.IsGet() && strings.HasPrefix(path, "/payments-summary"):
			handleProxy(ctx, storageClient)
		case ctx.IsPost() && strings.HasPrefix(path, "/purge-payments"):
			handleProxy(ctx, storageClient)
		default:
			ctx.Error(path, fasthttp.StatusNotFound)
		}
	}

	log.Printf("Listening on %s", serverSocket)
	log.Fatal(fasthttp.ListenAndServeUNIX(serverSocket, 0666, requestHandler))
}

func handleProxy(ctx *fasthttp.RequestCtx, socket *client.SocketClient) {
	req := &ctx.Request
	resp := &ctx.Response
	if err := socket.Do(req, resp); err != nil {
		log.Println("Error forwarding request:", resp.StatusCode(), err.Error())
		ctx.Error(err.Error(), fasthttp.StatusBadGateway)
	}
}
