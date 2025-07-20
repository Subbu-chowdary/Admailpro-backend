package metrics

import (
	"fmt"
	"sync/atomic"

	"github.com/valyala/fasthttp"
)

var sentEmails int64

// Handler exposes Prometheus-compatible metric endpoint
func Handler(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("text/plain; charset=utf-8")
	ctx.SetStatusCode(fasthttp.StatusOK)
	_, _ = ctx.WriteString(fmt.Sprintf("sent_emails_total %d\n", atomic.LoadInt64(&sentEmails)))
}

// IncrementSentEmails increments the email sent count
func IncrementSentEmails() {
	atomic.AddInt64(&sentEmails, 1)
}
