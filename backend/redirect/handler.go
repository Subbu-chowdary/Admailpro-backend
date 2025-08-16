package redirect

import (
	"context"
	"email-sender/backend/config"

	"github.com/go-redis/redis/v8"
	"github.com/valyala/fasthttp"
)

var rdb = redis.NewClient(&redis.Options{Addr: config.GetConfig().RedisAddr})

func Handler(ctx *fasthttp.RequestCtx) {
	id := string(ctx.QueryArgs().Peek("id"))
	if id == "" {
		ctx.Error("Missing id", fasthttp.StatusBadRequest)
		return
	}
	url, err := rdb.Get(context.Background(), id).Result()
	if err != nil {
		ctx.Error("Link not found", fasthttp.StatusNotFound)
		return
	}
	ctx.Redirect(url, fasthttp.StatusFound)
}