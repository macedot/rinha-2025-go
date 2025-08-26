package handler

import (
	"time"

	"github.com/macedot/rinha-2025-go/internal/types"
	"github.com/macedot/rinha-2025-go/pkg/client"
	"github.com/valyala/fasthttp"
)

func PaymentHandler(ctx *fasthttp.RequestCtx, client *client.SocketClient) {
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
	paymentJSON, _ := payment.MarshalJSON()
	_, err := client.Post("http://unix/payments", paymentJSON)
	if err != nil {
		ctx.Error("Failed to queue payment", fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetStatusCode(fasthttp.StatusAccepted)
}
