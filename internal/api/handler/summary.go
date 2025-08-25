package handler

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/macedot/rinha-2025-go/internal/types"
	"github.com/redis/go-redis/v9"
	"github.com/valyala/fasthttp"
)

func SummaryHandler(ctx *fasthttp.RequestCtx, rdb *redis.Client) {
	if !ctx.IsGet() {
		ctx.Error("Method not allowed", fasthttp.StatusMethodNotAllowed)
		return
	}

	fromStr := string(ctx.QueryArgs().Peek("from"))
	toStr := string(ctx.QueryArgs().Peek("to"))

	from := time.Unix(0, 0).UTC()
	to := time.Now().UTC()
	var err error

	if fromStr != "" {
		from, err = time.Parse(time.RFC3339, fromStr)
		if err != nil {
			ctx.Error("Invalid 'from' timestamp", fasthttp.StatusBadRequest)
			return
		}
	}

	if toStr != "" {
		to, err = time.Parse(time.RFC3339, toStr)
		if err != nil {
			ctx.Error("Invalid 'to' timestamp", fasthttp.StatusBadRequest)
			return
		}
	}

	var summary types.SummaryResponse
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		summary.Default = getSummary(rdb, from, to, "default_payments")
	}()
	go func() {
		defer wg.Done()
		summary.Fallback = getSummary(rdb, from, to, "fallback_payments")
	}()
	wg.Wait()

	responseJSON, _ := summary.MarshalJSON()
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType("application/json")
	ctx.SetBody(responseJSON)
}

func getSummary(rdb *redis.Client, from, to time.Time, key string) types.SummaryServer {
	results, err := rdb.ZRangeByScore(context.Background(), key, &redis.ZRangeBy{
		Min: fmt.Sprintf("%d", from.UnixNano()),
		Max: fmt.Sprintf("%d", to.UnixNano()),
	}).Result()
	if err != nil {
		log.Printf("Failed to get summary (%s): %v", key, err)
		return types.SummaryServer{}
	}
	count := len(results)
	totalAmount := int64(0)
	for _, result := range results {
		amount, _ := strconv.ParseInt(result, 10, 64)
		totalAmount += amount
	}
	return types.SummaryServer{
		TotalRequests: count,
		TotalAmount:   float64(totalAmount) / 100,
	}
}
