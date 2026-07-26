package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

func curlecLog(event string, fields map[string]string) {
	if fields == nil {
		fields = map[string]string{}
	}
	b, _ := json.Marshal(fields)
	log.Printf("[curlec] %s %s", event, string(b))
}

func jsonResponse(status int, body interface{}) (events.APIGatewayV2HTTPResponse, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return events.APIGatewayV2HTTPResponse{StatusCode: 500, Body: `{"error":"encode"}`}, nil
	}
	return events.APIGatewayV2HTTPResponse{
		StatusCode: status,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(b),
	}, nil
}

func redirectResponse(status int, location string) (events.APIGatewayV2HTTPResponse, error) {
	return events.APIGatewayV2HTTPResponse{
		StatusCode: status,
		Headers:    map[string]string{"Location": location},
		Body:       "",
	}, nil
}

func requestBodyBytes(req events.APIGatewayV2HTTPRequest) []byte {
	return []byte(req.Body)
}

func handleCurlecCreateOrder(req events.APIGatewayV2HTTPRequest, now time.Time, wg *sync.WaitGroup) (events.APIGatewayV2HTTPResponse, error) {
	curlecLog("create_order.start", nil)
	var body orderCreateRequest
	if err := json.Unmarshal(requestBodyBytes(req), &body); err != nil {
		return jsonResponse(400, map[string]string{"error": "invalid JSON"})
	}

	amountStr := body.Amount
	devTest := false
	amountStr, devTest = applyDevTestOrderOverride(body.Email, amountStr)

	amountF, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amountF <= 0 {
		return jsonResponse(400, map[string]string{"error": "invalid amount"})
	}

	currency := strings.ToUpper(body.Currency)
	if currency == "" {
		currency = "MYR"
	}

	cartRef := genRef(6)
	address := strings.TrimSpace(body.Address1 + " " + body.Address2)
	tradeInVal := ""
	if body.DeviceProperties != "" {
		var dp map[string]string
		if err := json.Unmarshal([]byte(body.DeviceProperties), &dp); err == nil {
			tradeInVal = dp["tradeIn"]
		}
	}

	rawItems := ""
	if len(body.Items) > 0 {
		rawItems = string(body.Items)
	}

	paymentMethod := "Standard Checkout Order"
	if devTest {
		paymentMethod = "DEV_TEST " + paymentMethod
	}

	row := []string{
		now.Format("02/01/2006 03:04:05PM"), body.Name, body.Email, body.Mobile,
		address, body.Postcode, "", "", "", body.Harness, paymentMethod,
		body.Country, currency, amountStr, body.Installation, body.PromoCode,
		rawItems, body.DeviceProperties, cartRef, "",
	}
	wg.Add(1)
	go updateGoogleSheets(row, "PaymentGateway", wg)

	full := notes{
		Name: body.Name, Phone: body.Mobile, Address: address, Email: body.Email,
		Country: body.Country, Harness: body.Harness, Total: amountStr,
		Shipping: body.ShippingRate, Duration: "1", Device: body.DevicePrice,
		Postcode: body.Postcode, PromoCode: body.PromoCode, Delivery: body.Installation,
		TradeIn: tradeInVal, CartRef: cartRef, Items: rawItems,
	}
	if devTest {
		full.DevTest = "1"
	}
	applyMetaToNotes(&full, body.metaAttributionRequest)

	razorNotes := slimRazorpayNotes(full, false)
	receipt := "kommu_" + cartRef
	orderData := map[string]interface{}{
		"amount":   rmToSen(amountF),
		"currency": currency,
		"receipt":  receipt,
		"notes":    razorNotes,
	}
	created, err := curlecClient().Order.Create(orderData, nil)
	if err != nil {
		curlecLog("create_order.error", map[string]string{"error": err.Error()})
		return jsonResponse(500, map[string]string{"error": err.Error()})
	}

	orderID, _ := created["id"].(string)
	_ = putCheckoutContext(orderID, CheckoutContextPayload{
		Kind:     "order",
		CartRef:  cartRef,
		DevTest:  devTest,
		Notes:    full,
		RawItems: rawItems,
	})

	curlecLog("create_order.ok", map[string]string{"order_id": orderID, "cart_ref": cartRef, "dev_test": fmt.Sprintf("%v", devTest)})
	amountSen := rmToSen(amountF)
	if a, ok := created["amount"].(float64); ok {
		amountSen = int(a)
	}

	return jsonResponse(200, map[string]interface{}{
		"key_id":   curlecPublicKey(),
		"order_id": orderID,
		"amount":   amountSen,
		"currency": currency,
		"receipt":  receipt,
		"dev_test": devTest,
	})
}

func handleCurlecCreateSubscription(req events.APIGatewayV2HTTPRequest, now time.Time, wg *sync.WaitGroup) (events.APIGatewayV2HTTPResponse, error) {
	var body subscriptionCreateRequest
	if err := json.Unmarshal(requestBodyBytes(req), &body); err != nil {
		return jsonResponse(400, map[string]string{"error": "invalid JSON"})
	}

	planID := body.PlanID
	if planID == "" {
		planID = body.Subscription
	}
	totalCount := body.TotalCount
	if totalCount < 1 && body.Duration != "" {
		totalCount, _ = strconv.Atoi(body.Duration)
	}
	if planID == "" || totalCount < 1 {
		return jsonResponse(400, map[string]string{"error": "plan_id and total_count required"})
	}

	depositStr := body.Deposit
	devTest := false
	depositStr, planID, devTest = applyDevTestSubscriptionOverride(body.Email, depositStr, planID)

	depositF, _ := strconv.ParseFloat(depositStr, 64)
	currency := strings.ToUpper(body.Currency)
	if currency == "" {
		currency = "MYR"
	}

	cartRef := genRef(6)
	address := strings.TrimSpace(body.Address1)
	paymentMethod := fmt.Sprintf("Rent-to-own - %d months", totalCount)
	if devTest {
		paymentMethod = "DEV_TEST " + paymentMethod
	}

	row := []string{
		now.Format("02/01/2006 03:04:05PM"), body.Name, body.Email, body.Mobile,
		address, body.Postcode, body.NRIC, "", strconv.Itoa(totalCount), body.Harness,
		paymentMethod, "Malaysia", currency, body.SubDevicePrice, body.Installation,
		body.PromoCode, "", "", cartRef, depositStr,
	}
	wg.Add(1)
	go updateGoogleSheets(row, "PaymentGateway", wg)

	full := notes{
		Name: body.Name, Phone: body.Mobile, Address: address, Email: body.Email,
		Country: "Malaysia", Harness: body.Harness, Total: body.SubDevicePrice,
		Shipping: "Free", Duration: strconv.Itoa(totalCount), Device: body.MonthlyPayment,
		PromoCode: body.PromoCode, Delivery: body.Installation, CartRef: cartRef,
		Deposit: depositStr, Sim: body.Sim, TradeIn: body.TradeIn,
	}
	if devTest {
		full.DevTest = "1"
	}
	applyMetaToNotes(&full, body.metaAttributionRequest)

	data := map[string]interface{}{
		"plan_id":     planID,
		"total_count": totalCount,
		"notes":       slimRazorpayNotes(full, false),
	}
	if depositF > 0 {
		data["addons"] = []interface{}{
			map[string]interface{}{
				"item": map[string]interface{}{
					"name":     "Deposit",
					"amount":   rmToSen(depositF),
					"currency": currency,
				},
			},
		}
	}

	created, err := curlecClient().Subscription.Create(data, nil)
	if err != nil {
		return jsonResponse(500, map[string]string{"error": err.Error()})
	}

	subID, _ := created["id"].(string)
	_ = putCheckoutContext(subID, CheckoutContextPayload{
		Kind:    "subscription",
		CartRef: cartRef,
		DevTest: devTest,
		Notes:   full,
	})

	return jsonResponse(200, map[string]interface{}{
		"key_id":          curlecPublicKey(),
		"subscription_id": subID,
		"dev_test":        devTest,
	})
}

func handleCurlecVerify(req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	var body struct {
		OrderID   string `json:"razorpay_order_id"`
		PaymentID string `json:"razorpay_payment_id"`
		Signature string `json:"razorpay_signature"`
	}
	if err := json.Unmarshal(requestBodyBytes(req), &body); err != nil {
		return jsonResponse(400, map[string]string{"error": "invalid JSON"})
	}
	valid := verifyPaymentSignature(body.OrderID, body.PaymentID, body.Signature)
	return jsonResponse(200, map[string]interface{}{"valid": valid})
}

func handleCurlecCallback(req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	vals, _ := url.ParseQuery(req.Body)
	orderID := vals.Get("razorpay_order_id")
	paymentID := vals.Get("razorpay_payment_id")
	signature := vals.Get("razorpay_signature")
	subID := vals.Get("razorpay_subscription_id")

	success := callbackBaseURL() + "/trx-success/"
	failure := callbackBaseURL() + "/trx-failure/"

	if subID != "" {
		return redirectResponse(302, success+"?subscription_id="+url.QueryEscape(subID))
	}
	if orderID != "" && paymentID != "" && signature != "" && verifyPaymentSignature(orderID, paymentID, signature) {
		q := url.Values{}
		q.Set("razorpay_order_id", orderID)
		q.Set("razorpay_payment_id", paymentID)
		q.Set("razorpay_signature", signature)
		return redirectResponse(302, success+"?"+q.Encode())
	}
	return redirectResponse(302, failure)
}

func handleCurlecWebhook(req events.APIGatewayV2HTTPRequest, now time.Time, wg *sync.WaitGroup) (events.APIGatewayV2HTTPResponse, error) {
	body := requestBodyBytes(req)
	sig := ""
	if req.Headers != nil {
		sig = req.Headers["x-razorpay-signature"]
		if sig == "" {
			sig = req.Headers["X-Razorpay-Signature"]
		}
	}
	if !verifyWebhookSignature(body, sig) {
		return jsonResponse(400, map[string]string{"error": "invalid signature"})
	}

	var envelope struct {
		Event string `json:"event"`
		Payload struct {
			Payment struct {
				Entity map[string]interface{} `json:"entity"`
			} `json:"payment"`
			Order struct {
				Entity map[string]interface{} `json:"entity"`
			} `json:"order"`
			Subscription struct {
				Entity map[string]interface{} `json:"entity"`
			} `json:"subscription"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return jsonResponse(400, map[string]string{"error": "invalid JSON"})
	}

	clientIP := headerOr(req, "x-forwarded-for")
	userAgent := headerOr(req, "user-agent")

	switch envelope.Event {
	case "payment.captured":
		pay := envelope.Payload.Payment.Entity
		orderID, _ := pay["order_id"].(string)
		paymentID, _ := pay["id"].(string)
		if orderID == "" {
			orderID, _ = envelope.Payload.Order.Entity["id"].(string)
		}
		if !claimPaymentFulfillment(orderID, paymentID) {
			return jsonResponse(200, map[string]string{"status": "duplicate"})
		}
		n := loadOrderNotesHydrated(orderID)
		fulfillCurlecOneOffPayment(n, orderID, now, wg, clientIP, userAgent)
	case "subscription.authenticated", "subscription.activated":
		sub := envelope.Payload.Subscription.Entity
		subID, _ := sub["id"].(string)
		n := loadSubscriptionNotesHydrated(subID)
		fulfillCurlecSubscriptionAuth(n, subID, now, wg, clientIP, userAgent)
	}

	return jsonResponse(200, map[string]string{"status": "ok"})
}

func headerOr(req events.APIGatewayV2HTTPRequest, key string) string {
	if req.Headers == nil {
		return ""
	}
	if v := req.Headers[key]; v != "" {
		return v
	}
	return req.Headers[http.CanonicalHeaderKey(key)]
}

func loadOrderNotesHydrated(orderID string) notes {
	raw, err := fetchOrderNotesMap(orderID)
	base := notes{}
	if err == nil {
		base = notesFromMap(raw)
	}
	return hydrateNotesFromContext(orderID, base)
}

func loadSubscriptionNotesHydrated(subID string) notes {
	base := notes{}
	if subID != "" {
		fetched, err := curlecClient().Subscription.Fetch(subID, nil, nil)
		if err == nil {
			if raw, ok := fetched["notes"].(map[string]interface{}); ok {
				base = notesFromMap(raw)
			}
		}
	}
	return hydrateNotesFromContext(subID, base)
}
