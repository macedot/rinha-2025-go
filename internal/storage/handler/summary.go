package handler

import (
	"fmt"
	"time"

	"github.com/macedot/rinha-2025-go/internal/storage/store"
	"github.com/valyala/fasthttp"
)

func SummaryHandler(ctx *fasthttp.RequestCtx, d *store.PaymentDB) {
	from := string(ctx.QueryArgs().Peek("from"))
	to := string(ctx.QueryArgs().Peek("to"))
	var (
		fromTime time.Time
		toTime   time.Time
		err      error
	)
	if from != "" {
		fromTime, err = parseTime(from)
		if err != nil {
			ctx.Error(err.Error(), fasthttp.StatusBadRequest)
			return
		}
	} else {
		fromTime = time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	if to != "" {
		toTime, err = parseTime(to)
		if err != nil {
			ctx.Error(err.Error(), fasthttp.StatusBadRequest)
			return
		}
	} else {
		toTime = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
	}
	if fromTime.After(toTime) {
		ctx.Error("'from' time cannot be after 'to' time", fasthttp.StatusBadRequest)
		return
	}
	summary := d.QuerySummary(fromTime, toTime)
	body, _ := summary.MarshalJSON()
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody(body)
}

func parseTime(s string) (time.Time, error) {
	// Try standard RFC3339 format first
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	// Try without timezone but with milliseconds
	if t, err := time.Parse("2006-01-02T15:04:05.999", s); err == nil {
		return t, nil
	}
	// Try without milliseconds
	if t, err := time.Parse("2006-01-02T15:04:05", s); err == nil {
		return t, nil
	}
	// Try with just date
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	// Try Unix timestamp
	if t, err := time.Parse(time.UnixDate, s); err == nil {
		return t, nil
	}
	// Return a proper error
	return time.Time{}, fmt.Errorf("unrecognized time format: %s", s)
}
