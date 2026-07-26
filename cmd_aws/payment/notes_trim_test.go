package main

import (
	"encoding/json"
	"testing"
)

func TestTrimNotesMapExcludesMetaAndKeepsOps(t *testing.T) {
	notes := map[string]interface{}{
		"meta_event_id": "e1", "meta_fbp": "fbp", "meta_fbc": "fbc", "meta_source_url": "https://x",
		"name": "A", "phone": "1", "address": "addr", "email": "a@b.c", "country": "Malaysia",
		"harness": "h", "total": "100", "shipping": "10", "duration": "1", "device": "90",
		"postcode": "50000", "promoCode": "X", "delivery": "Self", "trade_in": "yes",
		"cart_ref": "ABC123", "items": `[{"name":"Cable","quantity":1,"unitPrice":"50.00"}]`,
	}
	out := trimNotesMap(notes, 15)
	if _, ok := out["meta_event_id"]; ok {
		t.Fatalf("meta_event_id must not be in Razorpay notes")
	}
	if out["cart_ref"] != "ABC123" {
		t.Fatalf("cart_ref should be kept, got %#v", out["cart_ref"])
	}
	if out["items"] == nil || out["items"] == "" {
		t.Fatalf("items should be kept when Meta is excluded")
	}
	if len(out) > 15 {
		t.Fatalf("expected <=15 keys, got %d", len(out))
	}
}

func TestSlimRazorpayNotesUnderLimit(t *testing.T) {
	full := notes{
		Name: "Alex", Phone: "012", Address: "x", Email: "alexanderyeoh@kommu.ai",
		Country: "Malaysia", Harness: "h", Total: "1.00", Shipping: "0", Duration: "1",
		Device: "1", Postcode: "1", Delivery: "Self", TradeIn: "t", CartRef: "REF001",
		Deposit: "1", Sim: "s", DevTest: "1",
		MetaEventID: "e", MetaFbp: "b", MetaFbc: "c", MetaSourceURL: "u",
		Items: `[{"name":"Item","quantity":1,"unitPrice":"1.00"}]`,
	}
	slim := slimRazorpayNotes(full, true)
	if len(slim) > 14 {
		t.Fatalf("slim notes must leave room for fulfilled_payment_id, got %d keys: %#v", len(slim), slim)
	}
	for k := range slim {
		if len(k) > 4 && k[:5] == "meta_" {
			t.Fatalf("meta key leaked into slim notes: %s", k)
		}
	}
}

func TestDevTestAllowlist(t *testing.T) {
	if !isDevTestEmail("AlexanderYeoh@kommu.ai") {
		t.Fatal("expected allowlisted email")
	}
	if isDevTestEmail("customer@example.com") {
		t.Fatal("customer must not be allowlisted")
	}
	amt, ok := applyDevTestOrderOverride("keanwei@kommu.ai", "5199.00")
	if !ok || amt != "1.00" {
		t.Fatalf("expected RM1 override, got %s ok=%v", amt, ok)
	}
	amt2, ok2 := applyDevTestOrderOverride("x@y.com", "5199.00")
	if ok2 || amt2 != "5199.00" {
		t.Fatalf("non-allowlisted must keep amount")
	}
}

func TestMergeNotesPrefersOverlay(t *testing.T) {
	base := notes{Email: "a@b.c", CartRef: "R1", Total: "10"}
	over := notes{Items: `[{"name":"X"}]`, MetaEventID: "eid", Total: ""}
	m := mergeNotes(base, over)
	if m.Email != "a@b.c" || m.CartRef != "R1" {
		t.Fatalf("base fields lost: %#v", m)
	}
	if m.Items == "" || m.MetaEventID != "eid" {
		t.Fatalf("overlay fields missing: %#v", m)
	}
	if m.Total != "10" {
		t.Fatalf("empty overlay should not wipe total")
	}
}

func TestNotesJSONRoundTrip(t *testing.T) {
	n := notes{Name: "A", MetaEventID: "e", Items: "[]", DevTest: "1"}
	b, err := json.Marshal(n)
	if err != nil {
		t.Fatal(err)
	}
	var back notes
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if back.Name != "A" || back.MetaEventID != "e" || back.DevTest != "1" {
		t.Fatalf("roundtrip failed: %#v", back)
	}
}
