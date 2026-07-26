package main

import (
	"encoding/json"
	"strings"
)

const (
	curlecValueLimit = 200
	maxItemChunks    = 4
	maxRazorpayNotes = 15
)

// notes is the fulfillment view of checkout context / Razorpay notes.
type notes struct {
	Name         string `json:"name"`
	Phone        string `json:"phone"`
	Address      string `json:"address"`
	Email        string `json:"email"`
	Country      string `json:"country"`
	Harness      string `json:"harness"`
	Total        string `json:"total"`
	Shipping     string `json:"shipping"`
	Duration     string `json:"duration"`
	Device       string `json:"device"`
	Postcode     string `json:"postcode"`
	PromoCode    string `json:"promoCode"`
	Delivery     string `json:"delivery"`
	TradeIn      string `json:"trade_in"`
	CartRef      string `json:"cart_ref"`
	Items        string `json:"items"`
	Items1       string `json:"items_1"`
	Items2       string `json:"items_2"`
	Items3       string `json:"items_3"`
	Items4       string `json:"items_4"`
	Deposit      string `json:"deposit"`
	Sim          string `json:"sim"`
	DevTest      string `json:"dev_test"`
	MetaEventID  string `json:"meta_event_id"`
	MetaFbp      string `json:"meta_fbp"`
	MetaFbc      string `json:"meta_fbc"`
	MetaSourceURL string `json:"meta_source_url"`
}

type orderCreateRequest struct {
	Source           string          `json:"source"`
	Name             string          `json:"name"`
	Email            string          `json:"email"`
	Mobile           string          `json:"mobile"`
	IC               string          `json:"ic"`
	Address1         string          `json:"address1"`
	Address2         string          `json:"address2"`
	Postcode         string          `json:"postcode"`
	Country          string          `json:"country"`
	Currency         string          `json:"currency"`
	Amount           string          `json:"amount"`
	DevicePrice      string          `json:"deviceprice"`
	ShippingRate     string          `json:"shippingrate"`
	Harness          string          `json:"harness"`
	Installation     string          `json:"installation"`
	PromoCode        string          `json:"promoCode"`
	DeviceProperties string          `json:"deviceProperties"`
	Items            json.RawMessage `json:"items"`
	metaAttributionRequest
}

type subscriptionCreateRequest struct {
	PlanID          string `json:"plan_id"`
	Subscription    string `json:"subscription"`
	TotalCount      int    `json:"total_count"`
	Duration        string `json:"duration"`
	Deposit         string `json:"deposit"`
	Currency        string `json:"currency"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	Mobile          string `json:"mobile"`
	Address1        string `json:"address1"`
	Postcode        string `json:"postcode"`
	NRIC            string `json:"nric"`
	Harness         string `json:"harness"`
	Installation    string `json:"installation"`
	PromoCode       string `json:"promoCode"`
	SubDevicePrice  string `json:"sub_device_price"`
	MonthlyPayment  string `json:"monthlyPayment"`
	Sim             string `json:"sim"`
	TradeIn         string `json:"tradeIn"`
	metaAttributionRequest
}

type LineItem struct {
	Name        string `json:"name"`
	Quantity    int    `json:"quantity"`
	UnitPrice   string `json:"unitPrice"`
	Subcategory string `json:"subcategory"`
}

func notesFromMap(m map[string]interface{}) notes {
	get := func(k string) string {
		if v, ok := m[k]; ok {
			return strings.TrimSpace(asString(v))
		}
		return ""
	}
	return notes{
		Name: get("name"), Phone: get("phone"), Address: get("address"), Email: get("email"),
		Country: get("country"), Harness: get("harness"), Total: get("total"), Shipping: get("shipping"),
		Duration: get("duration"), Device: get("device"), Postcode: get("postcode"), PromoCode: get("promoCode"),
		Delivery: get("delivery"), TradeIn: get("trade_in"), CartRef: get("cart_ref"),
		Items: get("items"), Items1: get("items_1"), Items2: get("items_2"), Items3: get("items_3"), Items4: get("items_4"),
		Deposit: get("deposit"), Sim: get("sim"), DevTest: get("dev_test"),
		MetaEventID: get("meta_event_id"), MetaFbp: get("meta_fbp"), MetaFbc: get("meta_fbc"), MetaSourceURL: get("meta_source_url"),
	}
}

func notesToMap(n notes) map[string]interface{} {
	m := map[string]interface{}{
		"name": n.Name, "phone": n.Phone, "address": n.Address, "email": n.Email,
		"country": n.Country, "harness": n.Harness, "total": n.Total, "shipping": n.Shipping,
		"duration": n.Duration, "device": n.Device, "postcode": n.Postcode, "promoCode": n.PromoCode,
		"delivery": n.Delivery, "trade_in": n.TradeIn, "cart_ref": n.CartRef,
		"items": n.Items, "items_1": n.Items1, "items_2": n.Items2, "items_3": n.Items3, "items_4": n.Items4,
		"deposit": n.Deposit, "sim": n.Sim, "dev_test": n.DevTest,
		"meta_event_id": n.MetaEventID, "meta_fbp": n.MetaFbp, "meta_fbc": n.MetaFbc, "meta_source_url": n.MetaSourceURL,
	}
	return m
}

func mergeNotes(base notes, overlay notes) notes {
	m := notesToMap(base)
	for k, v := range notesToMap(overlay) {
		s := asString(v)
		if s != "" {
			m[k] = s
		}
	}
	return notesFromMap(m)
}

func asString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case nil:
		return ""
	default:
		b, _ := json.Marshal(t)
		return string(b)
	}
}

func isDevTestNotes(n notes) bool {
	return n.DevTest == "1" || strings.EqualFold(n.DevTest, "true")
}
