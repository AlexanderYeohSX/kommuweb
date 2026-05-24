# Checkout flow

## Entry points

1. **Product** — `product.html` → `goToCheckout()` sets `sessionStorage.checkoutSelections` and `fromProductPage=true` → `/checkout?mode=product`
2. **Cart** — `cart.html` → `goToCheckoutFromCart()` → `/checkout?mode=cart`

## Rule matching

`computePlan()` builds a context `{ source, currency, payment, dropdowns }` and calls `KommuPricing.matchRule()` against `_data/checkout_rules.yaml`.

**Providers (after migration):**

| Provider | Use case |
|----------|----------|
| `razorpay-order` | MYR one-off product or cart |
| `razorpay-subscription` | MYR rent-to-own |
| `stripe` | Non-MYR |

Legacy provider names `curlec-otp` / `curlec-sub` are still accepted in `checkout.html` for backward compatibility.

## `checkoutSubmit()` sequence

1. Validate form fields.
2. `computePricingForSelections()` — device price, deposit, monthly, `subDevicePrice`, promos.
3. Branch on `plan.provider`:
   - **Subscription** — `POST /curlec/subscriptions` → Razorpay Checkout with `subscription_id` and `callback_url` (FPX-safe `redirect: true`).
   - **Order** — `POST /curlec/orders` → Checkout with `order_id`, amount in sen from server.
   - **Stripe** — `window.open` to `/stripe/checkout?` (query params unchanged).

## Amount fields (rent-to-own)

| Field | Meaning |
|-------|---------|
| `devicePrice` / gateway `amount` (legacy) | One-off device + bundled add-ons for metadata |
| `sub_device_price` | Total subscription plan value (YAML or computed) |
| `deposit` | Upfront addon on `/v1/subscriptions` |
| `monthlyPayment` | Plan instalment (from Razorpay plan + promos) |

## Old vs new parameter mapping

| Legacy GET (`/curlec/otp`) | New |
|----------------------------|-----|
| `amount`, `deviceprice`, `shippingrate` | `POST /curlec/orders` JSON body |
| `items`, `deviceProperties` | Same keys in JSON |
| Customer fields | Same, server stores in order `notes` |

| Legacy GET (`/curlec/visa`) | New |
|-----------------------------|-----|
| `subscription` | `plan_id` |
| `duration` | `total_count` |
| `deposit` | `addons[0].item.amount` (sen) |
| `sub_device_price`, `harness`, etc. | `notes` on subscription |
