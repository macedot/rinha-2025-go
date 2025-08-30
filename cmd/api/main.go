package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/pprof"

	"github.com/macedot/rinha-2025-go/internal/store"
	"github.com/macedot/rinha-2025-go/internal/types"
	"github.com/macedot/rinha-2025-go/pkg/util"
)

func main() {
	runtime.GOMAXPROCS(1)
	appCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	apiSocketPath := util.GetEnv("SOCKET_PATH")
	//pollWorkerSize, _ := strconv.Atoi(util.GetEnvOr("WORKERS_POOL_SIZE", "10"))
	paymentDB, err := store.NewPaymentDB(appCtx)
	if err != nil {
		log.Fatalf("Unable to connect to the database: %v", err)
	}
	defer paymentDB.Close()
	h := server.New(
		server.WithNetwork("unix"),
		server.WithHostPorts(apiSocketPath),
	)
	pprof.Register(h)
	h.POST("/payments", func(c context.Context, ctx *app.RequestContext) {
		var payment types.PaymentRequest
		err := payment.UnmarshalJSON(ctx.Request.Body())
		if err != nil {
			ctx.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
			return
		}
		payment.RequestedAt = time.Now().UTC().Truncate(time.Millisecond)
		paymentDB.SavePayment(context.Background(), &payment, 1)
		ctx.JSON(consts.StatusOK, utils.H{"payment": "added"})
	})
	h.GET("/payments-summary", func(c context.Context, ctx *app.RequestContext) {
		args := ctx.QueryArgs()
		from, to := string(args.Peek("from")), string(args.Peek("to"))
		summary := paymentDB.GetSummary(context.Background(), from, to)
		ctx.JSON(consts.StatusOK, summary)
	})
	h.POST("/purge-payments", func(c context.Context, ctx *app.RequestContext) {
		paymentDB.Purge(context.Background())
		ctx.JSON(consts.StatusOK, utils.H{"purge": "OK"})
	})
	h.Spin()
}
