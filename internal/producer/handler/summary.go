package handler

import (
	"fmt"

	"github.com/macedot/rinha-2025-go/pkg/client"
	"github.com/valyala/fasthttp"
)

func SummaryHandler(ctx *fasthttp.RequestCtx, client *client.SocketClient) {
	if !ctx.IsGet() {
		ctx.Error("Method not allowed", fasthttp.StatusMethodNotAllowed)
		return
	}
	from := string(ctx.QueryArgs().Peek("from"))
	to := string(ctx.QueryArgs().Peek("to"))
	statusCode, body, err := client.Get(fmt.Sprintf("/summary?from=%s&to=%s", from, to))
	if err != nil || statusCode != fasthttp.StatusOK {
		ctx.Error(fmt.Sprintf("Failed to get summary (%d): %v", statusCode, err), fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType("application/json")
	ctx.SetBody(body)
}
