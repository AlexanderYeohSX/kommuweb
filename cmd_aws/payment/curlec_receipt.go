package main

import (
	"encoding/json"
	"fmt"
	"html"
	"net/smtp"
	"os"
	"strings"
	"sync"
	"time"
)

// fulfillCurlecOneOffPayment sends receipt email and logs PaymentComplete.
func fulfillCurlecOneOffPayment(invNotes notes, entityID string, now time.Time, wg *sync.WaitGroup, clientIP, userAgent string) {
	sendMetaPurchaseFromNotes(invNotes, entityID, "MYR", 0, clientIP, userAgent)

	itemsJSON := invNotes.Items
	if itemsJSON == "" {
		chunks := []string{invNotes.Items1, invNotes.Items2, invNotes.Items3, invNotes.Items4}
		itemsJSON = strings.Join(chunks, "")
	}

	useAccessories := false
	rowsHTML := ""
	itemsText := ""
	var invLineItems []LineItem
	if itemsJSON != "" {
		if err := json.Unmarshal([]byte(itemsJSON), &invLineItems); err != nil {
			invLineItems = nil
		}
	}
	if len(invLineItems) > 0 {
		useAccessories = true
		parts := make([]string, 0, len(invLineItems))
		var b strings.Builder
		for _, it := range invLineItems {
			if it.Name == "" {
				continue
			}
			qty := it.Quantity
			if qty <= 0 {
				qty = 1
			}
			nameWithSel := it.Name
			if it.Subcategory != "" {
				nameWithSel = fmt.Sprintf("%s (%s)", it.Name, it.Subcategory)
			}
			if it.UnitPrice != "" {
				parts = append(parts, fmt.Sprintf("%s x%d @ %s", nameWithSel, qty, it.UnitPrice))
			} else {
				parts = append(parts, fmt.Sprintf("%s x%d", nameWithSel, qty))
			}
			unitF, ok := parseAmount(it.UnitPrice)
			unitDisplay := it.UnitPrice
			if ok {
				unitDisplay = formatFloat("MYR", unitF)
			}
			totalStr := it.UnitPrice
			if ok {
				totalStr = formatFloat("MYR", unitF*float64(qty))
			}
			b.WriteString("<tr>")
			b.WriteString("<td style=\"padding:8px 0\"><div style=\"font-weight:500;color:#fff\">")
			b.WriteString(html.EscapeString(it.Name))
			b.WriteString("</div>")
			if it.Subcategory != "" {
				b.WriteString("<div style=\"font-size:12px;color:#ccc\">")
				b.WriteString(html.EscapeString(it.Subcategory))
				b.WriteString("</div>")
			}
			if unitDisplay != "" {
				b.WriteString("<div style=\"font-size:12px;color:#ccc\">@ ")
				b.WriteString(html.EscapeString(unitDisplay))
				b.WriteString("</div>")
			}
			b.WriteString("</td>")
			b.WriteString("<td align=\"right\" style=\"padding:8px 0;color:#fff\">x")
			b.WriteString(fmt.Sprintf("%d", qty))
			b.WriteString("</td>")
			b.WriteString("<td align=\"right\" style=\"padding:8px 0;color:#fff\">")
			b.WriteString(html.EscapeString(totalStr))
			b.WriteString("</td>")
			b.WriteString("</tr>")
		}
		rowsHTML = b.String()
		itemsText = strings.Join(parts, "; ")
	}

	subject := "Kommu — Payment received"
	if isDevTestNotes(invNotes) {
		subject = "[DEV TEST] " + subject
	}

	body := buildReceiptEmailHTML(invNotes, rowsHTML, useAccessories)
	if invNotes.Email != "" {
		if err := sendMail(invNotes.Email, subject, body); err != nil {
			curlecLog("fulfill.email_error", map[string]string{"error": err.Error()})
		} else {
			curlecLog("fulfill.email_ok", map[string]string{"email": invNotes.Email})
		}
	}

	sheetMethod := "Curlec one-off"
	if isDevTestNotes(invNotes) {
		sheetMethod = "DEV_TEST " + sheetMethod
	}
	row := []string{
		now.Format("02/01/2006 03:04:05PM"), invNotes.Name, invNotes.Email, invNotes.Phone,
		invNotes.Address, invNotes.Postcode, "", "", invNotes.Duration, invNotes.Harness,
		sheetMethod, invNotes.Country, "MYR", invNotes.Total, invNotes.Delivery, invNotes.PromoCode,
		itemsText, invNotes.TradeIn, invNotes.CartRef, invNotes.Deposit, entityID,
	}
	wg.Add(1)
	go updateGoogleSheets(row, "PaymentComplete", wg)

	if !isDevTestNotes(invNotes) {
		go createMPAFromNotes(invNotes)
	} else {
		curlecLog("fulfill.skip_mpa_dev_test", map[string]string{"cart_ref": invNotes.CartRef})
	}
}

func fulfillCurlecSubscriptionAuth(n notes, subID string, now time.Time, wg *sync.WaitGroup, clientIP, userAgent string) {
	sendMetaPurchaseFromNotes(n, subID, "MYR", 0, clientIP, userAgent)

	subject := "Kommu — Subscription authorised"
	if isDevTestNotes(n) {
		subject = "[DEV TEST] " + subject
	}
	body := buildSubscriptionEmailHTML(n)
	if n.Email != "" {
		_ = sendMail(n.Email, subject, body)
	}

	sheetMethod := "Curlec subscription"
	if isDevTestNotes(n) {
		sheetMethod = "DEV_TEST " + sheetMethod
	}
	row := []string{
		now.Format("02/01/2006 03:04:05PM"), n.Name, n.Email, n.Phone,
		n.Address, n.Postcode, "", "", n.Duration, n.Harness,
		sheetMethod, n.Country, "MYR", n.Total, n.Delivery, n.PromoCode,
		"", n.TradeIn, n.CartRef, n.Deposit, subID, n.Device,
	}
	wg.Add(1)
	go updateGoogleSheets(row, "PaymentComplete", wg)

	if !isDevTestNotes(n) {
		go createMPAFromNotes(n)
	}
}

func buildReceiptEmailHTML(n notes, rowsHTML string, accessories bool) string {
	var b strings.Builder
	b.WriteString("<div style=\"font-family:Arial,sans-serif;background:#111;color:#fff;padding:24px\">")
	b.WriteString("<h2>Thank you for your order</h2>")
	b.WriteString("<p>Hi ")
	b.WriteString(html.EscapeString(n.Name))
	b.WriteString(",</p>")
	b.WriteString("<p>We received your payment of <strong>")
	b.WriteString(html.EscapeString(n.Total))
	b.WriteString(" MYR</strong>.</p>")
	if n.CartRef != "" {
		b.WriteString("<p>Reference: ")
		b.WriteString(html.EscapeString(n.CartRef))
		b.WriteString("</p>")
	}
	if accessories && rowsHTML != "" {
		b.WriteString("<table style=\"width:100%;border-collapse:collapse\">")
		b.WriteString(rowsHTML)
		b.WriteString("</table>")
	} else if n.Harness != "" {
		b.WriteString("<p>Device / harness: ")
		b.WriteString(html.EscapeString(n.Harness))
		b.WriteString("</p>")
	}
	if n.Delivery != "" {
		b.WriteString("<p>Installation: ")
		b.WriteString(html.EscapeString(n.Delivery))
		b.WriteString("</p>")
	}
	if n.TradeIn != "" {
		b.WriteString("<p>Trade-in: ")
		b.WriteString(html.EscapeString(n.TradeIn))
		b.WriteString("</p>")
	}
	b.WriteString("<p style=\"color:#aaa;font-size:12px\">Kommu · kommu.ai</p></div>")
	return b.String()
}

func buildSubscriptionEmailHTML(n notes) string {
	var b strings.Builder
	b.WriteString("<div style=\"font-family:Arial,sans-serif;background:#111;color:#fff;padding:24px\">")
	b.WriteString("<h2>Subscription authorised</h2>")
	b.WriteString("<p>Hi ")
	b.WriteString(html.EscapeString(n.Name))
	b.WriteString(",</p>")
	b.WriteString("<p>Your rent-to-own plan is authorised.</p>")
	if n.Device != "" {
		b.WriteString("<p>Monthly instalment: ")
		b.WriteString(html.EscapeString(n.Device))
		b.WriteString(" MYR</p>")
	}
	if n.Deposit != "" {
		b.WriteString("<p>Deposit: ")
		b.WriteString(html.EscapeString(n.Deposit))
		b.WriteString(" MYR</p>")
	}
	if n.CartRef != "" {
		b.WriteString("<p>Reference: ")
		b.WriteString(html.EscapeString(n.CartRef))
		b.WriteString("</p>")
	}
	b.WriteString("<p style=\"color:#aaa;font-size:12px\">Kommu · kommu.ai</p></div>")
	return b.String()
}

func sendMail(to, subject, htmlBody string) error {
	from := envOr("SMTP_FROM", "noreply@kommu.ai")
	host := envOr("SMTP_HOST", "smtp.gmail.com")
	port := envOr("SMTP_PORT", "587")
	user := envOr("SMTP_USER", from)
	pass := os.Getenv("SMTP_PASSWORD")
	if pass == "" {
		curlecLog("email.skip_no_smtp_password", nil)
		return fmt.Errorf("SMTP_PASSWORD not set")
	}
	addr := host + ":" + port
	msg := []byte("To: " + to + "\r\n" +
		"From: " + from + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n\r\n" +
		htmlBody)
	auth := smtp.PlainAuth("", user, pass, host)
	return smtp.SendMail(addr, auth, from, []string{to}, msg)
}
