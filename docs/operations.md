# Operations

## Production Lambda (`CurlecGateway`)

| Setting | Value |
|---------|--------|
| Function | `CurlecGateway` |
| Runtime | `provided.al2023`, ARM64, handler `bootstrap` |
| HTTP API | `Kommu Gateway` (`ifhdr5efvk`) |
| Domain | `https://aws.kommu.ai` |
| CORS origins | `http://127.0.0.1:4000`, `https://kommu.ai`, `https://www.kommu.ai` (API Gateway; includes `OPTIONS` for preflight) |

## Environment variables (Lambda)

| Variable | Description |
|----------|-------------|
| `CURLEC_KEY_ID` | Razorpay public key (Checkout); legacy alias `CURLEC_KEY` |
| `CURLEC_KEY_SECRET` | Server-only; legacy alias `CURLEC_SECRET` |
| `CURLEC_WEBHOOK_SECRET` | Webhook signature (optional) |
| `GOOGLE_SPREADSHEET_ID` | Google Sheets ID; legacy alias `GOOGLE_SHEET_ID` |
| `GOOGLE_CREDENTIALS_JSON` | Service account JSON (single line); **required** after secrets scrub (set in Lambda console) |
| `STRIPE_SECRET_KEY` | Stripe secret |
| `STRIPE_WEBHOOK_SECRET` | Stripe webhook signing secret |
| `CALLBACK_BASE_URL` | Default `https://kommu.ai` |

See [cmd_aws/payment/.env.example](../cmd_aws/payment/.env.example).

## Deploy Lambda code

```bash
cd cmd_aws/cdk && npm run deploy:code
```

Or CDK: `cd cmd_aws/cdk && npm install && npm run deploy` (bundles via local `make zip`).

## Razorpay Curlec Dashboard

1. Enable **automatic payment capture**.
2. Allowlist callback domain `kommu.ai`.
3. Webhook URL: prefer `https://aws.kommu.ai/curlec/webhook` for JSON events (`payment.captured`, etc.). Keep `https://aws.kommu.ai/curlec/callback` for Checkout form redirects only, or use one URL for both (handler dedupes retries).
4. Razorpay retries webhooks if the response takes **>5 seconds** or is non-2xx. The Lambda marks orders with `fulfilled_payment_id` in Razorpay order notes to avoid duplicate emails on retries.
5. Checkout redirect: `https://aws.kommu.ai/curlec/callback` (form POST from Standard Checkout).

## Deploy Jekyll site

```bash
bundle exec jekyll build
```

Publish `_site/` to hosting (e.g. GitHub Pages for `kommu.ai`).

## Test vs live keys

Use Razorpay test keys in staging; rotate any keys that were ever committed in source control.
