package router

import (
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttprouter"

	"email-sender/backend/api"
	"email-sender/backend/metrics"
	"email-sender/backend/redirect"
	"email-sender/backend/smtp"
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
	router.POST("/api/send-email", wrap(api.AuthMiddleware(smtp.SendEmailHandler)))
	router.POST("/api/upload-csv", wrap(api.AuthMiddleware(api.UploadCSVHandler)))
	router.POST("/api/campaigns", wrap(api.AuthMiddleware(api.CreateCampaignHandler)))
	router.POST("/api/send-campaign", wrap(api.AuthMiddleware(smtp.SendCampaignHandler)))

	// --- Subdomain management (CRUD) ---
	router.GET("/api/subdomains", wrap(api.AuthMiddleware(api.GetSubdomains)))
	router.POST("/api/subdomains", wrap(api.AuthMiddleware(api.AddSubdomain)))
	router.PUT("/api/subdomains/:id", wrap(api.AuthMiddleware(api.UpdateSubdomain)))
	router.DELETE("/api/subdomains/:id", wrap(api.AuthMiddleware(api.DeleteSubdomain)))

	// --- Campaign management (CRUD) ---
	router.GET("/api/campaigns", wrap(api.AuthMiddleware(api.GetCampaignsHandler)))
	router.PUT("/api/campaigns/:id", wrap(api.AuthMiddleware(api.UpdateCampaignHandler)))
	router.DELETE("/api/campaigns/:id", wrap(api.AuthMiddleware(api.DeleteCampaignHandler)))
	router.GET("/api/campaigns/:id", wrap(api.AuthMiddleware(api.GetCampaignHandler)))

	// --- RecipientList management (CRUD) ---
	router.GET("/api/recipient-lists", wrap(api.AuthMiddleware(api.GetRecipientListsHandler)))
	router.GET("/api/recipient-lists/:id", wrap(api.AuthMiddleware(api.GetRecipientListHandler)))
	router.PUT("/api/recipient-lists/:id", wrap(api.AuthMiddleware(api.UpdateRecipientListHandler)))
	router.DELETE("/api/recipient-lists/:id", wrap(api.AuthMiddleware(api.DeleteRecipientListHandler)))
	
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