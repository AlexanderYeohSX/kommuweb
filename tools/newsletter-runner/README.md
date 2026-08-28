# Kommu newsletter runner (Athena)

Self-hosted on **Athena** via Tailscale:

- **Subscribe API** — public HTTPS via Tailscale Funnel → writes KA Inventory **Newsletter** tab
- **Drip sender** — hourly cron → SMTP
- **Order sync** — every 30m copies purchaser emails from **Orders** tab

See [docs/newsletter-setup.md](../../docs/newsletter-setup.md).

## Quick start on Athena

```bash
cd /data/kommu/newsletter-runner   # or your sync path
cp config.example.env .env         # KA Inventory + SMTP + service account
chmod +x install-athena-newsletter.sh
sudo ./install-athena-newsletter.sh install
sudo ./install-athena-newsletter.sh funnel
```

Copy the Funnel URL into kommuweb `_config.yml`:

```yaml
newsletter_api_url: "https://athena.YOUR-TAILNET.ts.net/newsletter/subscribe"
```

Rebuild/redeploy the site, then test the homepage form.

## Commands

```bash
sudo ./install-athena-newsletter.sh status
sudo ./install-athena-newsletter.sh test-api
sudo ./install-athena-newsletter.sh test-drip
sudo ./install-athena-newsletter.sh test-sync
python import_subscribers.py --csv customers.csv --source import --dry-run
```

## Sync from Mac

```bash
rsync -az --exclude .venv --exclude logs \
  tools/newsletter-runner/ kommu@192.168.0.80:/data/kommu/newsletter-runner/
scp .env kommu@192.168.0.80:/data/kommu/newsletter-runner/.env
```
