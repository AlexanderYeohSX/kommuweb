# QA checklist

Run after deploy with **test** Razorpay keys.

## Automated (repo)

```bash
cd cmd_aws/payment && go test ./...
cd ../../cmd_aws/cdk && npm install && npm run synth
cd ../../server && npm test
cd .. && bundle exec jekyll build
```

## Manual

| # | Scenario | Currency | Expected |
|---|----------|----------|----------|
| 1 | Product one-off, no trade-in | MYR | Checkout modal opens; payment → `trx-success` |
| 2 | Product one-off + trade-in promo | MYR | Order amount matches promo price |
| 3 | Product rent-to-own | MYR | Subscription Checkout; → `trx-success` with `subscription_id` |
| 4 | Cart accessories | MYR | Order with line items |
| 5 | Product/cart | USD | Stripe tab still opens |
| 6 | FPX (if enabled) | MYR | Redirect via `/curlec/callback` |
| 7 | Cancel Checkout modal | MYR | No charge; button re-enabled |
| 8 | Locale `en-MY` / `en-US` | — | Prices show MYR / USD consistently |
| 9 | `/accessories` | — | Redirects to `/store` |
| 10 | Invalid tampered signature on success URL | — | Error banner on `trx-success` |

Record payment IDs from Dashboard for any failed row.
