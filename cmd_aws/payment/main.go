package main

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	_ = ctx
	loadAppConfig()
	path := req.RawPath
	if path == "" && req.RequestContext.HTTP.Path != "" {
		path = req.RequestContext.HTTP.Path
	}
	method := strings.ToUpper(req.RequestContext.HTTP.Method)
	now := time.Now()
	var wg sync.WaitGroup

	curlecLog("request", map[string]string{"method": method, "path": path})

	switch {
	case method == "POST" && strings.HasSuffix(path, "/curlec/orders"):
		resp, err := handleCurlecCreateOrder(req, now, &wg)
		wg.Wait()
		return resp, err
	case method == "POST" && strings.HasSuffix(path, "/curlec/subscriptions"):
		resp, err := handleCurlecCreateSubscription(req, now, &wg)
		wg.Wait()
		return resp, err
	case method == "POST" && strings.HasSuffix(path, "/curlec/verify"):
		return handleCurlecVerify(req)
	case method == "POST" && strings.HasSuffix(path, "/curlec/callback"):
		return handleCurlecCallback(req)
	case method == "POST" && strings.HasSuffix(path, "/curlec/webhook"):
		resp, err := handleCurlecWebhook(req, now, &wg)
		wg.Wait()
		return resp, err
	case method == "GET" && (strings.HasSuffix(path, "/curlec/otp") || strings.HasSuffix(path, "/curlec/visa")):
		return jsonResponse(410, map[string]string{
			"error":   "Deprecated",
			"message": "Use POST /curlec/orders or /curlec/subscriptions with Standard Checkout",
		})
	case method == "OPTIONS":
		return events.APIGatewayV2HTTPResponse{StatusCode: 204}, nil
	default:
		return jsonResponse(404, map[string]string{"error": "not found", "path": path})
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	validateGoogleCredsJSON()
	lambda.Start(handler)
}
