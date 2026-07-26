package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func curlecPublicKey() string {
	return loadAppConfig().CurlecKeyID
}

func curlecSecret() string {
	return loadAppConfig().CurlecKeySecret
}

func verifyPaymentSignature(orderID, paymentID, signature string) bool {
	if orderID == "" || paymentID == "" || signature == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(curlecSecret()))
	mac.Write([]byte(orderID + "|" + paymentID))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(strings.ToLower(expected)), []byte(strings.ToLower(signature)))
}

func verifyWebhookSignature(body []byte, signature string) bool {
	secret := loadAppConfig().CurlecWebhookSecret
	if secret == "" {
		return true // optional in some setups
	}
	if signature == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(strings.ToLower(expected)), []byte(strings.ToLower(signature)))
}
