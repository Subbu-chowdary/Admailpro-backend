package main

import (
	"email-sender/backend/lambdaadapter"
	"email-sender/backend/router"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	// Set up the fasthttp router handler
	handler := router.SetupRouter()

	// Start Lambda adapter with handler
	adapter := lambdaadapter.NewAdapter(handler)
	lambda.Start(adapter.HandleRequest)
}
