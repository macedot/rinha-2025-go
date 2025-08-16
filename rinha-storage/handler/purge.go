package handler

import (
	"rinha-storage/store"

	"github.com/valyala/fasthttp"
)

func PurgePaymentsHandler(memoryDB *store.PaymentDB) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		// if err := fileDB.EraseAll(); err != nil {
		// 	ctx.Error("Failed to purge payments:"+err.Error(), fasthttp.StatusInternalServerError)
		// 	return
		// }
		memoryDB.Clean()
		ctx.SetStatusCode(fasthttp.StatusOK)
	}
}
