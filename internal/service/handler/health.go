package handler

import (
	"sync"

	"github.com/macedot/rinha-2025-go/internal/service/health"
	"github.com/macedot/rinha-2025-go/internal/types"
	"github.com/valyala/fasthttp"
)

func HealthHandler(ctx *fasthttp.RequestCtx, defaultChecker *health.HealthManager, fallbackChecker *health.HealthManager) {
	var health types.HealthSummary
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		health.Default = defaultChecker.GetHealthStorage()
	}()
	go func() {
		defer wg.Done()
		health.Fallback = fallbackChecker.GetHealthStorage()
	}()
	wg.Wait()
	ctx.SetStatusCode(fasthttp.StatusAccepted)
	body, _ := health.MarshalJSON()
	ctx.SetBody(body)
}
