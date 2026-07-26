package main

import "testing"

func TestCountryNameToISO(t *testing.T) {
	if got := countryNameToISO("Malaysia"); got != "my" {
		t.Fatalf("got %s", got)
	}
	if got := countryNameToISO("MY"); got != "my" {
		t.Fatalf("got %s", got)
	}
}

func TestVerifyPaymentSignatureRoundTrip(t *testing.T) {
	// Without keys, signature verification should fail closed for empty inputs.
	if verifyPaymentSignature("", "", "") {
		t.Fatal("empty should be invalid")
	}
}
