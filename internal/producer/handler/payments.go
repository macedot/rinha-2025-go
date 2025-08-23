package handler

import (
	"time"

	"github.com/macedot/rinha-2025-go/internal/types"
	"github.com/redis/go-redis/v9"
	"github.com/valyala/fasthttp"
)

func PaymentHandler(ctx *fasthttp.RequestCtx, rdb *redis.Client) {
	if !ctx.IsPost() {
		ctx.Error("Method not allowed", fasthttp.StatusMethodNotAllowed)
		return
	}
	var payment types.PaymentRequest
	if err := payment.UnmarshalJSON(ctx.PostBody()); err != nil {
		ctx.Error("Invalid JSON", fasthttp.StatusBadRequest)
		return
	}
	if payment.CorrelationID == "" || payment.Amount <= 0 {
		ctx.Error("Invalid payment data", fasthttp.StatusBadRequest)
		return
	}
	payment.RequestedAt = time.Now().UTC()
	paymentJSON, err := payment.MarshalJSON()
	if err != nil {
		ctx.Error("Serialization error", fasthttp.StatusInternalServerError)
		return
	}
	err = rdb.RPush(ctx, "payment_queue", paymentJSON).Err()
	if err != nil {
		ctx.Error("Failed to queue payment", fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetStatusCode(fasthttp.StatusAccepted)
}
