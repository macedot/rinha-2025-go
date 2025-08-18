package handler

import (
	"sync"

	"github.com/macedot/rinha-2025-go/internal/storage/store"
	"github.com/valyala/fasthttp"
)

var mu sync.Mutex

func PurgePaymentsHandler(ctx *fasthttp.RequestCtx, memoryDB *store.PaymentDB) {
	mu.Lock()
	defer mu.Unlock()
	memoryDB.Clean()
	ctx.SetStatusCode(fasthttp.StatusAccepted)
}
