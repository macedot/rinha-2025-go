package handler

import (
	"fmt"
	"os"

	"github.com/macedot/rinha-2025-go/pkg/client"
	"github.com/valyala/fasthttp"
)

func PurgePaymentsHandler(ctx *fasthttp.RequestCtx, client *client.SocketClient) {
	req := &ctx.Request
	resp := &ctx.Response
	if err := client.Do(req, resp); err != nil {
		fmt.Fprintf(os.Stderr, "Error sending request: %d %s", resp.StatusCode(), string(resp.Body()))
		ctx.Error(err.Error(), fasthttp.StatusBadGateway)
	}
}
