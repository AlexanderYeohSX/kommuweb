package main

import (
	"fmt"
	"strings"
	"sync"
)

const fulfilledPaymentNoteKey = "fulfilled_payment_id"

var inflightPayments sync.Map // paymentID -> struct{}

func fetchOrderNotesMap(orderID string) (map[string]interface{}, error) {
	if orderID == "" {
		return nil, fmt.Errorf("empty order id")
	}
	fetched, err := curlecClient().Order.Fetch(orderID, nil, nil)
	if err != nil {
		return nil, err
	}
	raw, ok := fetched["notes"].(map[string]interface{})
	if !ok {
		return map[string]interface{}{}, nil
	}
	return raw, nil
}

func updateOrderNotes(orderID string, notes map[string]interface{}) error {
	_, err := curlecClient().Order.Update(orderID, map[string]interface{}{
		"notes": trimNotesMap(notes, maxRazorpayNotes),
	}, nil)
	return err
}

// claimPaymentFulfillment marks an order as processing this payment.
// Returns false if this payment was already fulfilled or is already being processed.
func claimPaymentFulfillment(orderID, paymentID string) bool {
	if paymentID == "" {
		return true
	}
	if _, loaded := inflightPayments.LoadOrStore(paymentID, true); loaded {
		curlecLog("fulfill.inflight_local", map[string]string{"payment_id": paymentID})
		return false
	}

	notes, err := fetchOrderNotesMap(orderID)
	if err != nil {
		curlecLog("fulfill.claim_notes_fetch_error", map[string]string{
			"order_id": orderID,
			"error":    err.Error(),
		})
		// Proceed cautiously — better to risk a rare duplicate than drop fulfillment.
		return true
	}
	if existing, ok := notes[fulfilledPaymentNoteKey]; ok {
		if strings.TrimSpace(fmt.Sprintf("%v", existing)) == paymentID {
			curlecLog("fulfill.already_done", map[string]string{"payment_id": paymentID})
			inflightPayments.Delete(paymentID)
			return false
		}
	}
	notes[fulfilledPaymentNoteKey] = paymentID
	if err := updateOrderNotes(orderID, notes); err != nil {
		curlecLog("fulfill.claim_notes_update_error", map[string]string{
			"order_id": orderID,
			"error":    err.Error(),
		})
	}
	return true
}
