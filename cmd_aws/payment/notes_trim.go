package main

import (
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"
)

var timeNow = time.Now

func rmToSen(amount float64) int {
	return int(math.Round(amount * 100))
}

func parseAmount(s string) (float64, bool) {
	s = strings.TrimSpace(strings.ReplaceAll(s, ",", ""))
	if s == "" {
		return 0, false
	}
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	if err != nil {
		return 0, false
	}
	return f, true
}

func formatFloat(currency string, v float64) string {
	return fmt.Sprintf("%s %.2f", currency, v)
}

func chunkRunes(s string, size int) []string {
	if size <= 0 {
		return []string{s}
	}
	var out []string
	var b strings.Builder
	count := 0
	for _, r := range s {
		if count >= size {
			out = append(out, b.String())
			b.Reset()
			count = 0
		}
		b.WriteRune(r)
		count++
	}
	if b.Len() > 0 {
		out = append(out, b.String())
	}
	return out
}

// trimNotesMap keeps at most max non-empty Razorpay note pairs.
// Priority puts operational fulfillment fields first; Meta is NOT included
// (Meta lives in DynamoDB checkout context).
func trimNotesMap(notes map[string]interface{}, max int) map[string]interface{} {
	if max <= 0 || len(notes) == 0 {
		return map[string]interface{}{}
	}
	priority := []string{
		"fulfilled_payment_id",
		"cart_ref", "email", "name", "phone", "address", "total", "country",
		"items", "items_1", "items_2", "items_3", "items_4",
		"delivery", "trade_in", "deposit", "sim", "device",
		"harness", "shipping", "duration", "postcode", "promoCode", "dev_test",
	}
	out := make(map[string]interface{})
	add := func(k string) {
		if len(out) >= max {
			return
		}
		v, ok := notes[k]
		if !ok {
			return
		}
		s := fmt.Sprintf("%v", v)
		if s == "" {
			return
		}
		if utf8.RuneCountInString(s) > curlecValueLimit {
			s = string([]rune(s)[:curlecValueLimit])
		}
		out[k] = s
	}
	for _, k := range priority {
		add(k)
	}
	for k := range notes {
		if _, ok := out[k]; ok {
			continue
		}
		// Never put meta_* into Razorpay notes — DynamoDB owns them.
		if strings.HasPrefix(k, "meta_") {
			continue
		}
		add(k)
	}
	return out
}

// slimRazorpayNotes builds the ≤14-key notes map for Razorpay entities
// (leaves room for fulfilled_payment_id).
func slimRazorpayNotes(full notes, includeItems bool) map[string]interface{} {
	m := map[string]interface{}{
		"cart_ref": full.CartRef,
		"email":    full.Email,
		"name":     full.Name,
		"phone":    full.Phone,
		"address":  full.Address,
		"total":    full.Total,
		"country":  full.Country,
		"device":   full.Device,
		"shipping": full.Shipping,
		"delivery": full.Delivery,
		"harness":  full.Harness,
		"postcode": full.Postcode,
		"trade_in": full.TradeIn,
		"deposit":  full.Deposit,
		"sim":      full.Sim,
		"duration": full.Duration,
		"dev_test": full.DevTest,
	}
	if includeItems && full.Items != "" && utf8.RuneCountInString(full.Items) <= curlecValueLimit {
		// Optional short items for debugging; full cart is in DynamoDB.
		m["items"] = full.Items
	}
	return trimNotesMap(m, maxRazorpayNotes-1) // reserve 1 slot for fulfilled_payment_id
}

func genRef(n int) string {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	var b strings.Builder
	now := fmt.Sprintf("%d", nowUnixNano())
	for i := 0; i < n; i++ {
		b.WriteByte(alphabet[(int(now[len(now)-1-i%len(now)])+i*7)%len(alphabet)])
	}
	return b.String()
}

func nowUnixNano() int64 {
	return timeNow().UnixNano()
}
