package handler

import (
	"github.com/macedot/rinha-2025-go/internal/types"
	"github.com/valyala/fasthttp"
)

func HealthDefaultHandler(ctx *fasthttp.RequestCtx, h *types.HealthDB) {
	h.MuDefault.Lock()
	defer h.MuDefault.Unlock()
	method := string(ctx.Method())
	if method == "GET" {
		body, _ := h.Default.MarshalJSON()
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetBody(body)
		return
	}
	if method == "POST" {
		var health types.ProcessorHealth
		if err := health.UnmarshalJSON(ctx.PostBody()); err != nil {
			ctx.Error(err.Error(), fasthttp.StatusBadRequest)
			return
		}
		h.Default = health
		ctx.SetStatusCode(fasthttp.StatusOK)
		return
	}
	ctx.Error("Method Not Allowed", fasthttp.StatusMethodNotAllowed)
}

func HealthFallbackHandler(ctx *fasthttp.RequestCtx, h *types.HealthDB) {
	h.MuFallback.Lock()
	defer h.MuFallback.Unlock()
	method := string(ctx.Method())
	if method == "GET" {
		body, _ := h.Fallback.MarshalJSON()
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetBody(body)
		return
	}
	if method == "POST" {
		var health types.ProcessorHealth
		if err := health.UnmarshalJSON(ctx.PostBody()); err != nil {
			ctx.Error(err.Error(), fasthttp.StatusBadRequest)
			return
		}
		h.Fallback = health
		ctx.SetStatusCode(fasthttp.StatusOK)
		return
	}
	ctx.Error("Method Not Allowed", fasthttp.StatusMethodNotAllowed)
}
