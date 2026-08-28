# Newsletter signup + automated email chain

Homepage signup, **KA Inventory** subscriber list (including purchasers), and drip emails on **Athena** (Tailscale + Funnel).

## Architecture (Athena)

```
┌─────────────────┐   POST /newsletter/subscribe   ┌─────────────────────────────┐
│  kommu.ai       │ ─────────────────────────────► │ Tailscale Funnel (public)   │
│  (homepage)     │                                  │        ↓                    │
└─────────────────┘                                  │ Athena :8787 subscribe_api│
                                                       └─────────────┬─────────────┘
                                                                     │ upsert
┌─────────────────┐   every 30m sync_orders.py                     ▼
│  Orders tab     │ ─────────────────────────────► ┌─────────────────────────────┐
│  (KA Inventory) │                                  │ KA Inventory → Newsletter   │
└─────────────────┘                                  └─────────────┬─────────────┘
                                                                     │
┌─────────────────┐   hourly run.py (drip)                           │
│  Athena SMTP    │ ◄────────────────────────────────────────────────┘
└─────────────────┘
```

**Optional fallback:** set `newsletter_api_url` empty in `_config.yml` → form posts to `aws.kommu.ai` (requires Lambda).

## 1. Google Sheet — KA Inventory → `Newsletter` tab

Row 1 headers (exact order):

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

Share **KA Inventory** with your Google service account (Editor).

## 2. Email sequence

Defined in [`_data/newsletter_sequence.yaml`](../_data/newsletter_sequence.yaml). Edit copy in [`tools/newsletter-runner/templates/`](../tools/newsletter-runner/templates/).

## 3. Frontend (kommuweb)

- [`_includes/newsletter_signup.html`](../_includes/newsletter_signup.html) on `index.html`
- API URL from [`_config.yml`](../_config.yml) → `newsletter_api_url`

**Do not point the browser at `*.ts.net` Funnel URLs** — many networks (office Wi‑Fi, some ISPs, DNS filters) block Tailscale domains. The form should call **`https://aws.kommu.ai/newsletter/subscribe`** or a **Google Apps Script** web app instead.

```yaml
newsletter_api_url: "https://aws.kommu.ai/newsletter/subscribe"
```

### Option A — API Gateway proxy to Athena (recommended with `aws.kommu.ai`)

Browsers → `aws.kommu.ai` → API Gateway HTTP proxy → Athena Funnel → `subscribe_api.py` → KA Inventory.

In **AWS Console** → API Gateway → **Kommu Gateway** (`ifhdr5efvk`):

1. **Routes** → create `POST /newsletter/subscribe`
2. Integration: **HTTP proxy** → `https://athena.tail3f9a13.ts.net/newsletter/subscribe`
3. Ensure **CORS** allows `POST` + `Content-Type` from `kommu.ai`, `www.kommu.ai`, `alexanderyeohsx.github.io`
4. Deploy the API stage

Test:

```bash
curl -X POST https://aws.kommu.ai/newsletter/subscribe \
  -H 'Content-Type: application/json' \
  -H 'Origin: https://kommu.ai' \
  -d '{"email":"you@example.com","name":"Test","source":"homepage"}'
```

Reference Node proxy: [`server/src/routes/newsletter.js`](../server/src/routes/newsletter.js) (for local dev or future Lambda deploy).

### Option B — Google Apps Script (no AWS, no Tailscale in browser)

Same pattern as [installers](installers-google-sheet.md). Paste [`docs/scripts/newsletter-subscribe-api.gs`](scripts/newsletter-subscribe-api.gs) into **KA Inventory** → Extensions → Apps Script → Deploy → Web app (Anyone).

Set `_config.yml`:

```yaml
newsletter_api_url: "https://script.google.com/macros/s/YOUR_DEPLOYMENT_ID/exec"
```

Athena drip/sync timers are unchanged; only the homepage subscribe path uses Apps Script.

## 4. Athena — subscribe API + Funnel (server-side)

Location: [`tools/newsletter-runner/`](../tools/newsletter-runner/)

### Install on Athena

```bash
# Sync from Mac (example)
rsync -az --exclude .venv --exclude logs \
  tools/newsletter-runner/ kommu@192.168.0.80:/data/kommu/newsletter-runner/
scp path/to/.env kommu@192.168.0.80:/data/kommu/newsletter-runner/.env

# On Athena
cd /data/kommu/newsletter-runner
cp config.example.env .env   # edit: KA Inventory ID, service account, SMTP
sudo ./install-athena-newsletter.sh install
sudo ./install-athena-newsletter.sh funnel
```

`install` enables:

| Unit | Role |
|------|------|
| `kommu-newsletter-api.service` | Subscribe API on `127.0.0.1:8787` |
| `kommu-newsletter-drip.timer` | Hourly drip emails |
| `kommu-newsletter-sync.timer` | Every 30m: **Orders** tab → **Newsletter** tab |

`funnel` runs:

```bash
sudo tailscale funnel --bg 8787
tailscale funnel status   # copy https URL
```

Paste that URL (with path `/newsletter/subscribe`) into `_config.yml`, rebuild the site, test the homepage form.

### API

| Method | Path | Body |
|--------|------|------|
| `POST` | `/newsletter/subscribe` | `{ "email", "name?", "source?" }` |
| `GET` | `/health` | — |

CORS origins: `NEWSLETTER_CORS_ORIGINS` in `.env` (default includes `kommu.ai`, `alexanderyeohsx.github.io`).

### Test

```bash
sudo ./install-athena-newsletter.sh test-api
curl -X POST https://YOUR-FUNNEL-HOST/newsletter/subscribe \
  -H 'Content-Type: application/json' \
  -d '{"email":"you@example.com","name":"Test","source":"homepage"}'
```

## 5. Drip emails

Hourly on Athena via `run.py`. Manual test:

```bash
sudo ./install-athena-newsletter.sh test-drip
```

Logs: `logs/newsletter-drip.log`, `logs/subscribe-api.log`

## 6. Purchasers + import

- **Automatic:** `sync_orders.py` reads **Orders** tab (`ORDERS_SHEET_TAB`, default `Orders`), upserts with `source=checkout`
- **One-time CSV:** `python import_subscribers.py --csv customers.csv --source import`

## 7. Unsubscribe

Set `status=unsubscribed` in the sheet when users reply or email support@kommu.ai.

## 8. Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| “Could not reach the subscribe service” | Browser/network blocks `*.ts.net` | Use `aws.kommu.ai` proxy (§3 Option A) or Apps Script (§3 Option B) |
| `404` on `aws.kommu.ai/newsletter/subscribe` | API Gateway route not added | §3 Option A |
| Row missing in sheet | Service account / Apps Script permissions | Share KA Inventory with service account or run Apps Script as sheet owner |

## 9. Privacy

Signup requires marketing consent + link to [Privacy Policy](/privacy/).
