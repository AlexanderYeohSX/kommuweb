# Newsletter signup + automated email chain

Homepage signup, Google Sheet subscriber list (including purchasers), and a self-hosted drip runner on your own machine.

## Architecture

```
┌─────────────────┐     POST /newsletter/subscribe     ┌──────────────────┐
│  kommu.ai       │ ───────────────────────────────► │ aws.kommu.ai     │
│  (homepage form)│                                    │ (Lambda)         │
└─────────────────┘                                    └────────┬─────────┘
                                                                │ upsert
┌─────────────────┐     payment.captured webhook              ▼
│  Checkout       │ ───────────────────────────────► ┌──────────────────┐
└─────────────────┘                                    │ Google Sheet     │
                                                       │ tab: Newsletter  │
┌─────────────────┐     cron (your device)             └────────┬─────────┘
│ newsletter-     │ ◄──────────────── read / update ────────────┘
│ runner (SMTP)   │
└─────────────────┘
```

## 1. Google Sheet — `Newsletter` tab

Create a tab named **Newsletter** in the same spreadsheet used for orders (`GOOGLE_SPREADSHEET_ID`). Row 1 headers (exact order):

| Column | Description |
|--------|-------------|
| `email` | Primary key (lowercase) |
| `name` | Display name |
| `source` | `homepage`, `checkout`, or `import` |
| `subscribed_at` | ISO8601 UTC when first added |
| `sequence_step` | `0` = none sent yet; `1`–`6` = last step sent |
| `last_sent_at` | ISO8601 UTC of last drip email |
| `status` | `active`, `completed`, or `unsubscribed` |
| `next_send_at` | Optional; runner can leave blank |

**Upsert rule:** match on `email`. Never reset `sequence_step` if the row already exists (so a purchaser who signed up on the homepage keeps progress).

Share the spreadsheet with your Google service account email (Editor).

## 2. Email sequence

Defined in [`_data/newsletter_sequence.yaml`](../_data/newsletter_sequence.yaml):

| Step | ID | Delay after previous | Topic |
|------|-----|----------------------|--------|
| 1 | `branding` | 0 days (on subscribe) | Welcome + brand |
| 2 | `product_features` | 2 days | Product features |
| 3 | `use_cases` | 2 days | Use cases + value |
| 4 | `product_background` | 3 days | Company / product story |
| 5 | `hardware_specs` | 3 days | Hardware in the box |
| 6 | `technical_info` | 4 days | Specs, safety, compatibility |

Edit copy in [`tools/newsletter-runner/templates/`](../tools/newsletter-runner/templates/). Use `{{name}}` for personalization.

## 3. Frontend (kommuweb)

- Homepage section: [`_includes/newsletter_signup.html`](../_includes/newsletter_signup.html) (included on `index.html`)
- Submits `POST https://aws.kommu.ai/newsletter/subscribe` with `{ email, name?, source: "homepage" }`

## 4. Backend API (aws.kommu.ai)

See [backend-api.md](backend-api.md#post-newslettersubscribe).

### Lambda implementation checklist

Add to **CurlecGateway** (`cmd_aws/payment/`):

1. **`POST /newsletter/subscribe`**
   - Validate email
   - Upsert row in `Newsletter` tab via existing Google Sheets client
   - If new row: `sequence_step=0`, `status=active`, `subscribed_at=now`, `source` from body
   - If existing row: update `name` if provided; do **not** decrement step or change `status` if `unsubscribed`

2. **On `payment.captured` / successful checkout**
   - Upsert purchaser email with `source=checkout`, same rules as above
   - Purchasers appear in the same list as homepage signups

3. **CORS** — allow `POST /newsletter/subscribe` from `kommu.ai`, `www.kommu.ai`, GitHub Pages staging origin

Reference Node handler: [`server/src/routes/newsletter.js`](../server/src/routes/newsletter.js).

## 5. Self-hosted email runner

Location: [`tools/newsletter-runner/`](../tools/newsletter-runner/)

```bash
cd tools/newsletter-runner
cp config.example.env .env
# Edit .env — same GOOGLE_SPREADSHEET_ID + service account as Lambda
pip install -r requirements.txt
python run.py
```

Or use `./run.sh` (creates venv on first run).

### Cron example (hourly)

```cron
0 * * * * /path/to/kommuweb/tools/newsletter-runner/run.sh >> /var/log/kommu-newsletter.log 2>&1
```

The runner:

1. Reads rows where `status=active` and `sequence_step < 6`
2. Checks delay from `subscribed_at` (step 1) or `last_sent_at` (steps 2+)
3. Sends via SMTP
4. Updates `sequence_step`, `last_sent_at`, and sets `status=completed` after step 6

### Unsubscribe

Until a dedicated endpoint exists, handle manually: set `status=unsubscribed` in the sheet when users reply "unsubscribe" or email support@kommu.ai.

## 6. Import existing / purchased customers

One-time backfill from your orders sheet or CSV:

1. Export unique emails + names from order history
2. Run import script (dry-run first):

```bash
cd tools/newsletter-runner
python import_subscribers.py --csv /path/to/customers.csv --source import
```

Or paste rows manually with `source=import`, `sequence_step=0`, `status=active`, `subscribed_at=<today UTC>`.

Purchasers who should **skip** the drip can be set to `status=completed` and `sequence_step=6`.

## 7. Testing

1. Deploy Lambda `/newsletter/subscribe`
2. Submit the homepage form → confirm row in Sheet
3. Set `subscribed_at` to yesterday and run `python run.py` → step 1 sends
4. Complete a test checkout → confirm purchaser row with `source=checkout`
5. Verify SMTP / SPF / DKIM on your sending domain

## 8. Privacy

The signup form links to the [Privacy Policy](/privacy/). Marketing consent checkbox is required before submit.
