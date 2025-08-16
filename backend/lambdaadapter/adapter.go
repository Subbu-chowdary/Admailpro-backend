package lambdaadapter

import (
	"context"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/valyala/fasthttp"
)

type Adapter struct {
	handler fasthttp.RequestHandler
}

func NewAdapter(handler fasthttp.RequestHandler) *Adapter {
	return &Adapter{handler: handler}
}

func (a *Adapter) HandleRequest(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fReq := fasthttp.AcquireRequest()
	fRes := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(fReq)
	defer fasthttp.ReleaseResponse(fRes)

	fReq.SetRequestURI(req.Path)
	fReq.Header.SetMethod(req.HTTPMethod)
	fReq.SetBodyString(req.Body)

	// Headers
	for k, v := range req.Headers {
		fReq.Header.Set(k, v)
	}

	// Query params
	if len(req.QueryStringParameters) > 0 {
		qs := ""
		for k, v := range req.QueryStringParameters {
			qs += k + "=" + v + "&"
		}
		fReq.URI().SetQueryString(strings.TrimRight(qs, "&"))
	}

	// Simulate request
	ctxFast := &fasthttp.RequestCtx{}
	ctxFast.Init(fReq, nil, nil)
	a.handler(ctxFast)

	respHeaders := make(map[string]string)
	ctxFast.Response.Header.VisitAll(func(k, v []byte) {
		respHeaders[string(k)] = string(v)
	})

	return events.APIGatewayProxyResponse{
		StatusCode: ctxFast.Response.StatusCode(),
		Headers:    respHeaders,
		Body:       string(ctxFast.Response.Body()),
	}, nil
}
