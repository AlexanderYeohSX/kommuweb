# Meta Conversions API (CAPI)

Server-side conversion events for Meta Ads, paired with the browser Meta Pixel on kommuweb. Both channels use the same `event_id` so Meta deduplicates `Purchase` events.

See also [operations.md](operations.md) for Lambda env vars.

## Events

| Event | Browser (Pixel) | Server (CAPI) | When |
|-------|-----------------|---------------|------|
| `PageView` | All pages | — | Page load |
| `InitiateCheckout` | `checkout.html` | — | Valid checkout submit |
| `Purchase` | `trx-success.html` | Razorpay `payment.captured` + Stripe `checkout.session.completed` | Payment captured |

## Attribution passthrough

At checkout, the storefront sends:

- `meta_event_id` — UUID for deduplication
- `meta_fbp` — `_fbp` cookie
- `meta_fbc` — `_fbc` cookie or `fbclid`-derived value
- `meta_source_url` — checkout page URL

**Razorpay (MYR):** stored in Razorpay order/subscription `notes`.

**Stripe (non-MYR):** stored in Checkout Session `metadata`; `success_url` includes `meta_event_id`, `meta_value`, `meta_currency` for the browser `Purchase` on return.

## Lambda configuration

| Variable | Required | Description |
|----------|----------|-------------|
| `META_PIXEL_ID` | Yes (default `586857186218202`) | Events Manager pixel ID |
| `META_CAPI_ACCESS_TOKEN` | Yes for CAPI | From Events Manager → Conversions API |
| `META_CAPI_TEST_EVENT_CODE` | Staging only | Test Events tab code; omit in production |

Generate the access token in [Meta Events Manager](https://business.facebook.com/events_manager) → Pixel → Settings → Conversions API.

**Never commit** `META_CAPI_ACCESS_TOKEN` to git.

## Testing

1. Set `META_CAPI_TEST_EVENT_CODE` on staging Lambda.
2. Complete a MYR test checkout → confirm one deduplicated `Purchase` in Test Events.
3. Complete a non-MYR Stripe checkout → confirm `Purchase` with correct currency/value.
4. Verify Event Match Quality includes hashed email/phone and `fbp` where available.
5. Replay a webhook → confirm no duplicate CAPI send (Razorpay `fulfilled_payment_id` gate; Stripe session id dedupe).

## Troubleshooting

| Issue | Check |
|-------|--------|
| No server `Purchase` | `META_CAPI_ACCESS_TOKEN` set on Lambda; CloudWatch `[curlec] meta_capi.*` logs |
| Double counting | Same `meta_event_id` on Pixel and CAPI; Events Manager dedup window |
| Stripe browser `Purchase` missing | `success_url` query params; open `/trx-success/?meta_event_id=…` |
| Missing match quality | Ensure checkout collects email/phone; `_fbp` cookie present |

## Implementation files

| Layer | Files |
|-------|--------|
| Jekyll | `_includes/head.html`, `_includes/meta_pixel_helpers.html`, `checkout.html`, `trx-success.html` |
| Lambda | `cmd_aws/payment/meta_capi.go`, `curlec_standard.go`, `curlec_receipt.go`, `main.go` (Stripe) |
