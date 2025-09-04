package handler

import (
	"math"

	"github.com/macedot/rinha-2025-go/internal/storage/store"
	"github.com/macedot/rinha-2025-go/internal/types"
	"github.com/valyala/fasthttp"
)

func PaymentDefaultHandler(ctx *fasthttp.RequestCtx, memoryDB *store.PaymentDB) {
	body := ctx.PostBody()
	record, err := parsePaymentRecord(body)
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusBadRequest)
		return
	}
	memoryDB.AddRecordDefault(record)
	ctx.SetStatusCode(fasthttp.StatusAccepted)
}

func PaymentFallbackHandler(ctx *fasthttp.RequestCtx, memoryDB *store.PaymentDB) {
	body := ctx.PostBody()
	record, err := parsePaymentRecord(body)
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusBadRequest)
		return
	}
	memoryDB.AddRecordFallback(record)
	ctx.SetStatusCode(fasthttp.StatusAccepted)
}

func parsePaymentRecord(body []byte) (*types.PaymentRecord, error) {
	var request types.PaymentRequest
	if err := request.UnmarshalJSON(body); err != nil {
		return nil, err
	}
	return &types.PaymentRecord{
		Amount:      int64(math.Round(request.Amount * 100)),
		RequestedAt: request.RequestedAt,
	}, nil
}
