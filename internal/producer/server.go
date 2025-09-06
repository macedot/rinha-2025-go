package producer

import (
	"log"
	"net/http"
	"os"
	"rinha-2025-go/internal/config"
	"rinha-2025-go/internal/models"
	"rinha-2025-go/pkg/utils"

	"github.com/ohler55/ojg/oj"
	"github.com/valyala/fasthttp"
)

func PostPayment(producer *Producer) func(c *fasthttp.RequestCtx) {
	return func(c *fasthttp.RequestCtx) {
		var payment models.Payment
		if err := oj.Unmarshal(c.PostBody(), &payment); err != nil {
			c.Error(err.Error(), fasthttp.StatusBadRequest)
			return
		}
		go producer.EnqueuePayment(&payment)
		c.SetStatusCode(fasthttp.StatusAccepted)
	}
}

func GetSummary(producer *Producer) func(c *fasthttp.RequestCtx) {
	return func(c *fasthttp.RequestCtx) {
		from := utils.UnsafeString(c.QueryArgs().Peek("from"))
		to := utils.UnsafeString(c.QueryArgs().Peek("to"))
		summary, err := producer.GetSummary(from, to)
		if err != nil {
			c.Error(err.Error(), fasthttp.StatusInternalServerError)
			return
		}
		body, err := oj.Marshal(summary)
		if err != nil {
			c.Error(err.Error(), fasthttp.StatusInternalServerError)
			return
		}
		c.SetBody(body)
		c.SetStatusCode(fasthttp.StatusOK)
	}
}

func PostPurgePayments(producer *Producer) func(c *fasthttp.RequestCtx) {
	return func(c *fasthttp.RequestCtx) {
		if err := producer.PurgePayments(); err != nil {
			c.Error(err.Error(), fasthttp.StatusInternalServerError)
			return
		}
		c.SetStatusCode(http.StatusOK)
	}
}

func RunServer(cfg *config.Config, producer *Producer) error {
	if cfg.ServerSocket == "" {
		log.Fatalln("ServerSocket is empty")
	}
	handlers := fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/payments":
			PostPayment(producer)(ctx)
		case "/payments-summary":
			GetSummary(producer)(ctx)
		case "/purge-payments":
			PostPurgePayments(producer)(ctx)
		default:
			ctx.Error("Not Found", fasthttp.StatusNotFound)
		}
	})
	defer os.Remove(cfg.ServerSocket)
	return fasthttp.ListenAndServeUNIX(cfg.ServerSocket, 0666, handlers)
}
