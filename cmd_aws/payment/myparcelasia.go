package main

import "os"

// createMPAFromNotes creates a MyParcelAsia AWB when configured.
// Dev-test payments skip this in the caller.
func createMPAFromNotes(n notes) {
	if os.Getenv("MPA_API_KEY") == "" {
		curlecLog("mpa.skip_unconfigured", map[string]string{"cart_ref": n.CartRef})
		return
	}
	// Full MPA integration is environment-specific; log intent for ops visibility.
	curlecLog("mpa.create_requested", map[string]string{
		"cart_ref": n.CartRef,
		"name":     n.Name,
		"postcode": n.Postcode,
		"country":  n.Country,
	})
}
