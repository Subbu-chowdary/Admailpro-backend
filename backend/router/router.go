package router

import (
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttprouter"

	"email-sender/backend/api"
	"email-sender/backend/metrics"
	"email-sender/backend/redirect"
)

// SetupRouter returns a fasthttp-compatible request handler
func SetupRouter() fasthttp.RequestHandler {
	router := fasthttprouter.New()

	// Wrapper to convert standard handler
	wrap := func(handler func(ctx *fasthttp.RequestCtx)) fasthttprouter.Handle {
		return func(ctx *fasthttp.RequestCtx, _ fasthttprouter.Params) {
			handler(ctx)
		}
	}

	// Public endpoints
	router.POST("/api/signup", wrap(api.SignupHandler))
	router.POST("/api/login", wrap(api.LoginHandler))

	// Protected endpoints
	router.POST("/api/send-email", wrap(api.AuthMiddleware(api.SendEmailHandler)))
	router.POST("/api/upload-csv", wrap(api.AuthMiddleware(api.UploadCSVHandler)))
	router.POST("/api/campaigns", wrap(api.AuthMiddleware(api.CreateCampaignHandler)))
	router.POST("/api/send-campaign", wrap(api.AuthMiddleware(api.SendCampaignHandler)))

	// Subdomain management
	router.GET("/api/subdomains", wrap(api.GetSubdomains))
	router.POST("/api/subdomains", wrap(api.AddSubdomain))
	router.PUT("/api/subdomains/:id", wrap(api.UpdateSubdomain))
	router.DELETE("/api/subdomains/:id", wrap(api.DeleteSubdomain))

	// Metrics
	router.GET("/metrics", wrap(metrics.Handler))

	// Redirect tracking
	router.GET("/api/redirect", wrap(redirect.Handler))

	// Health check
	router.GET("/healthz", func(ctx *fasthttp.RequestCtx, _ fasthttprouter.Params) {
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetBodyString("OK")
	})

	// CORS wrapper
	return func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
		ctx.Response.Header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		ctx.Response.Header.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if string(ctx.Method()) == "OPTIONS" {
			ctx.SetStatusCode(fasthttp.StatusOK)
			return
		}

		router.Handler(ctx)
	}
}
