package main

import "testing"

func TestApplyMetaToNotes(t *testing.T) {
	n := notes{}
	applyMetaToNotes(&n, metaAttributionRequest{
		MetaEventID: "id", MetaFbp: "fbp", MetaFbc: "fbc", MetaSourceURL: "https://kommu.ai/checkout/",
	})
	if n.MetaEventID != "id" || n.MetaFbp != "fbp" {
		t.Fatalf("%#v", n)
	}
}
