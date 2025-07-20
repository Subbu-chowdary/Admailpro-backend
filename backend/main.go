package main

import (
	"log"

	"github.com/valyala/fasthttp"

	"email-sender/backend/queue"
	"email-sender/backend/router"
)

func main() {
	// Start worker (for background tasks)
	go queue.StartWorker()

	log.Println(" Server running on http://localhost:8080")
	log.Fatal(fasthttp.ListenAndServe(":8080", router.SetupRouter()))
}
