package main

import (
	"sync"

	razorpay "github.com/razorpay/razorpay-go"
)

var (
	rzpClient     *razorpay.Client
	rzpClientOnce sync.Once
)

func curlecClient() *razorpay.Client {
	rzpClientOnce.Do(func() {
		rzpClient = razorpay.NewClient(curlecPublicKey(), curlecSecret())
	})
	return rzpClient
}
