package server

import (
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"rinha-2025-go/internal/config"
	"rinha-2025-go/internal/models"
	"rinha-2025-go/internal/services"
	"rinha-2025-go/pkg/utils"

	"github.com/ohler55/ojg/oj"
	"github.com/valyala/fasthttp"
)

func PostPayment(worker *services.PaymentWorker) func(c *fasthttp.RequestCtx) {
	return func(c *fasthttp.RequestCtx) {
		var payment models.Payment
		if err := oj.Unmarshal(c.PostBody(), &payment); err != nil {
			c.Error(err.Error(), fasthttp.StatusBadRequest)
			return
		}
		go worker.EnqueuePayment(&payment)
		c.SetStatusCode(fasthttp.StatusAccepted)
	}
}

func GetSummary(worker *services.PaymentWorker) func(c *fasthttp.RequestCtx) {
	return func(c *fasthttp.RequestCtx) {
		from := utils.UnsafeString(c.QueryArgs().Peek("from"))
		to := utils.UnsafeString(c.QueryArgs().Peek("to"))
		summary, err := worker.GetSummary(from, to)
		if err != nil {
			c.Error(err.Error(), fasthttp.StatusInternalServerError)
			return
		}
		body, err := oj.Marshal(summary)
		if err != nil {
			c.Error(err.Error(), fasthttp.StatusInternalServerError)
			return
		}
		c.SetStatusCode(fasthttp.StatusOK)
		c.SetBody(body)
	}
}

func PostPurgePayments(worker *services.PaymentWorker) func(c *fasthttp.RequestCtx) {
	return func(c *fasthttp.RequestCtx) {
		if err := worker.PurgePayments(); err != nil {
			c.Error(err.Error(), fasthttp.StatusInternalServerError)
			return
		}
		c.SetStatusCode(http.StatusOK)
	}
}

func NewListenSocket(socketPath string) net.Listener {
	if err := os.MkdirAll(filepath.Dir(socketPath), 0777); err != nil {
		log.Fatalf("Failed to create socket directory: %v", err)
	}
	if err := os.RemoveAll(socketPath); err != nil {
		log.Fatalf("Failed to remove existing socket: %v", err)
	}
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("Failed to listen on Unix socket: %v", err)
	}
	if err := os.Chmod(socketPath, 0666); err != nil {
		log.Fatalf("Failed to set socket permissions: %v", err)
	}
	return listener
}

func RunServer(cfg *config.Config, worker *services.PaymentWorker) error {
	handlers := fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/payments":
			PostPayment(worker)(ctx)
		case "/payments-summary":
			GetSummary(worker)(ctx)
		case "/purge-payments":
			PostPurgePayments(worker)(ctx)
		default:
			ctx.Error("Not Found", fasthttp.StatusNotFound)
		}
	})

	if cfg.ServerSocket == "" {
		return fasthttp.ListenAndServe(":9999", handlers)
	}

	defer os.Remove(cfg.ServerSocket)
	return fasthttp.ListenAndServeUNIX(cfg.ServerSocket, 0666, handlers)
}
