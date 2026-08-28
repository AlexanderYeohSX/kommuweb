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
- API URL from [`_config.yml`](../_config.yml) → `newsletter_api_url` (Funnel URL after Athena install)

```yaml
newsletter_api_url: "https://athena.YOUR-TAILNET.ts.net/newsletter/subscribe"
```

## 4. Athena — subscribe API + Funnel

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

## 8. Optional — aws.kommu.ai Lambda

If you prefer Lambda instead of Funnel, leave `newsletter_api_url` empty and implement [backend-api.md](backend-api.md#post-newslettersubscribe) on **CurlecGateway**.

## 9. Privacy

Signup requires marketing consent + link to [Privacy Policy](/privacy/).
