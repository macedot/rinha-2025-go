package handler

import (
	"fmt"
	"log"
	"time"

	"github.com/macedot/rinha-2025-go/pkg/client"
	"github.com/valyala/fasthttp"
)

func SummaryHandler(ctx *fasthttp.RequestCtx, client *client.HttpClient) {
	if !ctx.IsGet() {
		ctx.Error("Method not allowed", fasthttp.StatusMethodNotAllowed)
		return
	}
	from := string(ctx.QueryArgs().Peek("from"))
	to := string(ctx.QueryArgs().Peek("to"))
	log.Println("Get summary from", from, "to", to)
	statusCode, body, err := client.GetTimeout(fmt.Sprintf("http://unix/summary?from=%s&to=%s", from, to), 1*time.Minute)
	log.Println("Get summary result", statusCode, string(body))
	if err != nil || statusCode != fasthttp.StatusOK {
		ctx.Error(fmt.Sprintf("Failed to get summary (%d): %v", statusCode, err), fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType("application/json")
	ctx.SetBody(body)
}
