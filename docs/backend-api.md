# Backend API (aws.kommu.ai)

Implementation: Go Lambda in [`cmd_aws/payment/`](../cmd_aws/payment/). Deploy with [`cmd_aws/cdk/`](../cmd_aws/cdk/) to the existing **CurlecGateway** function behind `https://aws.kommu.ai`.

The [`server/`](../server/) Node package is a reference implementation only (not production).

## Authentication

- **Orders / Subscriptions**: HTTP Basic with `RAZORPAY_KEY_ID` / `RAZORPAY_KEY_SECRET` (server-side only).
- **Browser**: receives only `key_id` and `order_id` or `subscription_id`.

## POST /curlec/orders

Creates a Razorpay order. **Amount in request body is major units (RM)**; server converts to sen.

**Request (JSON):**

```json
{
  "source": "product",
  "name": "Customer",
  "email": "a@example.com",
  "mobile": "0123456789",
  "amount": "5199.00",
  "currency": "MYR",
  "deviceprice": "4999.00",
  "shippingrate": "200.00",
  "harness": "Toyota Camry 2024",
  "promoCode": ""
}
```

**Response:**

```json
{
  "key_id": "rzp_test_xxx",
  "order_id": "order_xxx",
  "amount": 519900,
  "currency": "MYR",
  "receipt": "kommu_..."
}
```

## POST /curlec/subscriptions

Creates a Razorpay subscription with optional upfront addon (deposit).

**Request (JSON):**

```json
{
  "plan_id": "plan_ST3HOHumxjDqrv",
  "total_count": 24,
  "deposit": "1999.00",
  "currency": "MYR",
  "harness": "...",
  "sub_device_price": "6199.00",
  "monthlyPayment": "199.00",
  "promoCode": ""
}
```

Subscription `notes` mirror legacy `/curlec/visa`: `total` = plan value (`sub_device_price`), `device` = monthly instalment (`monthlyPayment`) for receipt `SubAmount`.

**Response:**

```json
{
  "key_id": "rzp_test_xxx",
  "subscription_id": "sub_xxx"
}
```

## POST /curlec/verify

**Request:**

```json
{
  "razorpay_order_id": "order_xxx",
  "razorpay_payment_id": "pay_xxx",
  "razorpay_signature": "..."
}
```

**Response:** `{ "valid": true }`

HMAC: `sha256(order_id + "|" + payment_id, key_secret)`.

## POST /curlec/callback

Form POST from Razorpay Checkout (`redirect: true`). Verifies signature and redirects to:

- `trx-success/?razorpay_order_id=...` (one-off orders)
- `trx-success/?subscription_id=...` (rent-to-own subscriptions)

## POST /curlec/webhook

Raw JSON body; header `X-Razorpay-Signature` verified with `RAZORPAY_WEBHOOK_SECRET`.

## Deprecated

- `GET /curlec/otp` → 410
- `GET /curlec/visa` → 410
