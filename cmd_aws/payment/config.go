package main

import (
	"os"
	"strings"
	"sync"
)

// AppConfig holds runtime configuration from environment variables (Lambda / local).
type AppConfig struct {
	CurlecKeyID           string
	CurlecKeySecret       string
	CurlecWebhookSecret   string
	CurlecDevTestPlanID   string
	GoogleSpreadsheetID   string
	GoogleCredentialsJSON string
	CallbackBaseURL       string
	StorefrontOrigin      string
	MetaPixelID           string
	MetaCAPIAccessToken   string
	MetaCAPITestEventCode string
	CheckoutContextTable  string
	StripeSecretKey       string
	StripeWebhookSecret   string
}

var (
	appConfig     AppConfig
	appConfigOnce sync.Once
)

func loadAppConfig() AppConfig {
	appConfigOnce.Do(func() {
		appConfig = AppConfig{
			CurlecKeyID:           envOr("CURLEC_KEY_ID", os.Getenv("CURLEC_KEY")),
			CurlecKeySecret:       envOr("CURLEC_KEY_SECRET", os.Getenv("CURLEC_SECRET")),
			CurlecWebhookSecret:   os.Getenv("CURLEC_WEBHOOK_SECRET"),
			CurlecDevTestPlanID:   os.Getenv("CURLEC_DEV_TEST_PLAN_ID"),
			GoogleSpreadsheetID:   envOr("GOOGLE_SPREADSHEET_ID", envOr("GOOGLE_SHEET_ID", "")),
			GoogleCredentialsJSON: os.Getenv("GOOGLE_CREDENTIALS_JSON"),
			CallbackBaseURL:       envOr("CALLBACK_BASE_URL", "https://kommu.ai"),
			StorefrontOrigin:      envOr("STOREFRONT_ORIGIN", "https://kommu.ai"),
			MetaPixelID:           envOr("META_PIXEL_ID", "586857186218202"),
			MetaCAPIAccessToken:   os.Getenv("META_CAPI_ACCESS_TOKEN"),
			MetaCAPITestEventCode: os.Getenv("META_CAPI_TEST_EVENT_CODE"),
			CheckoutContextTable:  envOr("CHECKOUT_CONTEXT_TABLE", "kommu-checkout-context"),
			StripeSecretKey:       os.Getenv("STRIPE_SECRET_KEY"),
			StripeWebhookSecret:   os.Getenv("STRIPE_WEBHOOK_SECRET"),
		}
	})
	return appConfig
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func spreadsheetID() string {
	return loadAppConfig().GoogleSpreadsheetID
}

func callbackBaseURL() string {
	return strings.TrimRight(strings.TrimSpace(loadAppConfig().CallbackBaseURL), "/")
}

func checkoutContextTable() string {
	return loadAppConfig().CheckoutContextTable
}
