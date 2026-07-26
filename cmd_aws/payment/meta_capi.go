package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"log"
)

const metaGraphAPIVersion = "v21.0"

var digitsOnly = regexp.MustCompile(`\D`)

// MetaPurchaseInput is the normalized payload for a Purchase CAPI event.
type MetaPurchaseInput struct {
	EventID   string
	Email     string
	Phone     string
	Name      string
	Country   string
	Fbp       string
	Fbc       string
	SourceURL string
	Currency  string
	Value     float64
	OrderID   string
	ClientIP  string
	UserAgent string
}

type metaAttributionRequest struct {
	MetaEventID   string `json:"meta_event_id"`
	MetaFbp       string `json:"meta_fbp"`
	MetaFbc       string `json:"meta_fbc"`
	MetaSourceURL string `json:"meta_source_url"`
}

// applyMetaToNotes writes Meta attribution into the full context notes (NOT Razorpay slim notes).
func applyMetaToNotes(n *notes, meta metaAttributionRequest) {
	if meta.MetaEventID != "" {
		n.MetaEventID = meta.MetaEventID
	}
	if meta.MetaFbp != "" {
		n.MetaFbp = meta.MetaFbp
	}
	if meta.MetaFbc != "" {
		n.MetaFbc = meta.MetaFbc
	}
	if meta.MetaSourceURL != "" {
		n.MetaSourceURL = meta.MetaSourceURL
	}
}

func metaPurchaseFromNotes(n notes, entityID, currency string, value float64) MetaPurchaseInput {
	curr := strings.ToUpper(strings.TrimSpace(currency))
	if curr == "" {
		curr = "MYR"
	}
	if value <= 0 {
		if v, ok := parseAmount(n.Total); ok {
			value = v
		}
	}
	return MetaPurchaseInput{
		EventID:   n.MetaEventID,
		Email:     n.Email,
		Phone:     n.Phone,
		Name:      n.Name,
		Country:   countryNameToISO(n.Country),
		Fbp:       n.MetaFbp,
		Fbc:       n.MetaFbc,
		SourceURL: n.MetaSourceURL,
		Currency:  curr,
		Value:     value,
		OrderID:   entityID,
	}
}

func sendMetaPurchaseFromNotes(n notes, entityID, currency string, value float64, clientIP, userAgent string) {
	if isDevTestNotes(n) && loadAppConfig().MetaCAPITestEventCode == "" {
		curlecLog("meta_capi.skip_dev_test", map[string]string{"entity_id": entityID})
		return
	}
	in := metaPurchaseFromNotes(n, entityID, currency, value)
	in.ClientIP = clientIP
	in.UserAgent = userAgent
	sendMetaPurchase(in)
}

func sendMetaPurchase(in MetaPurchaseInput) {
	cfg := loadAppConfig()
	if cfg.MetaCAPIAccessToken == "" || cfg.MetaPixelID == "" {
		curlecLog("meta_capi.skip_unconfigured", nil)
		return
	}
	userData := map[string]interface{}{}
	if in.Email != "" {
		userData["em"] = []string{sha256Hex(strings.ToLower(strings.TrimSpace(in.Email)))}
	}
	if in.Phone != "" {
		phone := digitsOnly.ReplaceAllString(in.Phone, "")
		if phone != "" {
			userData["ph"] = []string{sha256Hex(phone)}
		}
	}
	if in.Fbp != "" {
		userData["fbp"] = in.Fbp
	}
	if in.Fbc != "" {
		userData["fbc"] = in.Fbc
	}
	if in.ClientIP != "" {
		userData["client_ip_address"] = in.ClientIP
	}
	if in.UserAgent != "" {
		userData["client_user_agent"] = in.UserAgent
	}
	if in.Country != "" {
		userData["country"] = []string{sha256Hex(strings.ToLower(in.Country))}
	}

	custom := map[string]interface{}{
		"currency": in.Currency,
		"value":    in.Value,
	}
	if in.OrderID != "" {
		custom["order_id"] = in.OrderID
	}

	event := map[string]interface{}{
		"event_name":      "Purchase",
		"event_time":      time.Now().Unix(),
		"action_source":   "website",
		"user_data":       userData,
		"custom_data":     custom,
		"event_source_url": firstNonEmpty(in.SourceURL, cfg.StorefrontOrigin+"/checkout/"),
	}
	if in.EventID != "" {
		event["event_id"] = in.EventID
	}

	body := map[string]interface{}{
		"data": []interface{}{event},
	}
	if cfg.MetaCAPITestEventCode != "" {
		body["test_event_code"] = cfg.MetaCAPITestEventCode
	}

	raw, _ := json.Marshal(body)
	url := fmt.Sprintf("https://graph.facebook.com/%s/%s/events?access_token=%s",
		metaGraphAPIVersion, cfg.MetaPixelID, cfg.MetaCAPIAccessToken)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		log.Printf("meta capi purchase: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		curlecLog("meta_capi.error", map[string]string{"error": err.Error()})
		return
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	curlecLog("meta_capi.ok", map[string]string{
		"status":    fmt.Sprintf("%d", resp.StatusCode),
		"event_id":  in.EventID,
		"order_id":  in.OrderID,
		"response":  string(respBody),
	})
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
