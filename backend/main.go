// backend/main.go
package main

import (
	"log"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttprouter"

	"email-sender/backend/api"
	"email-sender/backend/metrics"
	"email-sender/backend/queue"
	"email-sender/backend/redirect"
)

func main() {
	// Start background queue worker
	go queue.StartWorker()

	// Intialise fasthttp router
	router := fasthttprouter.New()

	// Helper: Wrap standard fasthttp handlers to fit into signature
	wrap := func(handler func(ctx *fasthttp.RequestCtx)) fasthttprouter.Handle {

		return func(ctx *fasthttp.RequestCtx, _ fasthttprouter.Params) {
			handler(ctx)
		}

	}

	// Public endpoints
	router.POST("/api/signup", wrap(api.SignupHandler))
	router.POST("/api/login", wrap(api.LoginHandler))

	// Protected endpoints with auth middleware
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

	// Email open redirect tracking
	router.GET("/api/redirect", wrap(redirect.Handler))

	// Health check
	router.GET("/healthz", func(ctx *fasthttp.RequestCtx, _ fasthttprouter.Params) {
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetBodyString("OK")
	})

	// CORS middleware wrapper
	corsHandler := func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
		ctx.Response.Header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		ctx.Response.Header.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if string(ctx.Method()) == "OPTIONS" {
			ctx.SetStatusCode(fasthttp.StatusOK)
			return
		}

		router.Handler(ctx)
	}

	log.Println("🚀 Server running on http://localhost:8080")
	log.Fatal(fasthttp.ListenAndServe(":8080", corsHandler))
}
