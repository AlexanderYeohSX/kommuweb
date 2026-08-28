# Kommu newsletter runner

Self-hosted drip email sender. Reads subscribers from Google Sheets and sends the sequence defined in `_data/newsletter_sequence.yaml`.

See [docs/newsletter-setup.md](../docs/newsletter-setup.md) for full setup.

## Quick start

```bash
cp config.example.env .env
# Edit .env
./run.sh
```

## Import existing customers

```bash
python import_subscribers.py --csv customers.csv --source import --dry-run
python import_subscribers.py --csv customers.csv --source import
```

CSV columns: `email`, optional `name`.
