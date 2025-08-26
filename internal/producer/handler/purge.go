package handler

// func PurgePaymentsHandler(ctx *fasthttp.RequestCtx, rdb *redis.Client) {
// 	if !ctx.IsPost() {
// 		ctx.Error("Method not allowed", fasthttp.StatusMethodNotAllowed)
// 	}
// 	if err := rdb.FlushDB(context.Background()).Err(); err != nil {
// 		ctx.Error("Failed to purge queues", fasthttp.StatusInternalServerError)
// 		return
// 	}
// 	ctx.SetStatusCode(fasthttp.StatusOK)
// }
