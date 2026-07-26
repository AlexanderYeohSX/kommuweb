# Operations

## Production Lambda (`CurlecGateway`)

| Setting | Value |
|---------|--------|
| Function | `CurlecGateway` |
| Runtime | `provided.al2023`, ARM64, handler `bootstrap` |
| HTTP API | `Kommu Gateway` (`ifhdr5efvk`) |
| Domain | `https://aws.kommu.ai` |
| CORS origins | `http://127.0.0.1:4000`, `https://kommu.ai`, `https://www.kommu.ai`, `https://alexanderyeohsx.github.io` (API Gateway; includes `OPTIONS` for preflight) |

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
| `META_PIXEL_ID` | Meta pixel ID (default `586857186218202`) |
| `META_CAPI_ACCESS_TOKEN` | Meta Conversions API access token (**secret**) |
| `META_CAPI_TEST_EVENT_CODE` | Optional; Events Manager test code for staging |
| `CHECKOUT_CONTEXT_TABLE` | DynamoDB table for full checkout payload (default `kommu-checkout-context`) |
| `CURLEC_DEV_TEST_PLAN_ID` | Live Curlec plan id for RM1/month developer subscription tests |
| `SMTP_PASSWORD` | SMTP password for receipt email (`SMTP_HOST`/`SMTP_USER` optional) |

See [meta-conversions-api.md](meta-conversions-api.md) for event mapping, notes overflow fix, and RM1 smoke testing.

See [cmd_aws/payment/.env.example](../cmd_aws/payment/.env.example).

## Deploy Lambda code

```bash
cd cmd_aws/cdk && npm run deploy:code
```

Or CDK: `cd cmd_aws/cdk && npm install && npm run deploy` (creates `kommu-checkout-context` table at 1 RCU/1 WCU + bundles via local `make zip`).

After first CDK deploy, merge these into the existing Lambda environment in the console (do **not** replace the whole map blindly):

- `CHECKOUT_CONTEXT_TABLE=kommu-checkout-context`
- `CURLEC_DEV_TEST_PLAN_ID=<your RM1 plan id>`

## Razorpay Curlec Dashboard

1. Enable **automatic payment capture**.
2. Allowlist callback domains `kommu.ai` and (for staging) `alexanderyeohsx.github.io` in Razorpay Checkout settings.
3. Webhook URL: prefer `https://aws.kommu.ai/curlec/webhook` for JSON events (`payment.captured`, etc.). Keep `https://aws.kommu.ai/curlec/callback` for Checkout form redirects only, or use one URL for both (handler dedupes retries).
4. Razorpay retries webhooks if the response takes **>5 seconds** or is non-2xx. The Lambda marks orders with `fulfilled_payment_id` in Razorpay order notes to avoid duplicate emails on retries.
5. Checkout redirect: `https://aws.kommu.ai/curlec/callback` (form POST from Standard Checkout).
6. Create a **RM1/month** live plan once; put its `plan_…` id in `CURLEC_DEV_TEST_PLAN_ID`.

## Deploy Jekyll site

```bash
bundle exec jekyll build
```

Publish `_site/` to hosting (e.g. GitHub Pages for `kommu.ai`). `cmd_aws` / `server` / `docs` are excluded from the Jekyll build.

## Test vs live keys

Use Razorpay test keys in staging; rotate any keys that were ever committed in source control.

For **production flow smoke tests** after each code deploy, checkout with `alexanderyeoh@kommu.ai` or `keanwei@kommu.ai` — the Lambda forces RM1 and marks the run `DEV_TEST` (no AWB / no prod CAPI).
