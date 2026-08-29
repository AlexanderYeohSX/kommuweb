#!/usr/bin/env bash
# Set unsubscribe env vars on Athena (same Apps Script /exec URL + shared secret).
#
# Usage (on Athena):
#   cd ~/kommuweb/tools/newsletter-runner   # or your clone path
#   ./configure-unsubscribe.sh
#
# Also set UNSUBSCRIBE_SECRET in Apps Script → Project settings → Script properties.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
ENV_FILE="$ROOT/.env"

set_env_var() {
  local key="$1" value="$2"
  if grep -qE "^${key}=" "$ENV_FILE"; then
    sed -i "s|^${key}=.*|${key}=${value}|" "$ENV_FILE"
  else
    echo "${key}=${value}" >>"$ENV_FILE"
  fi
}

if [[ ! -f "$ENV_FILE" ]]; then
  echo "Missing $ENV_FILE"
  echo "Copy config.example.env to .env and fill SMTP / Google credentials first."
  exit 1
fi

echo "Newsletter unsubscribe — Athena .env setup"
echo "======================================="
echo
echo "Use the same Google Apps Script /exec URL as kommuweb _config.yml → newsletter_api_url"
echo "Example: https://script.google.com/macros/s/AKfycb.../exec"
echo

read -r -p "Apps Script URL: " APPS_URL
APPS_URL="$(echo "$APPS_URL" | tr -d '[:space:]')"

if [[ -z "$APPS_URL" ]]; then
  echo "URL is required." >&2
  exit 1
fi
if [[ "$APPS_URL" != https://script.google.com/macros/s/* ]]; then
  echo "Warning: expected a script.google.com/macros/s/.../exec URL" >&2
fi

echo
echo "UNSUBSCRIBE_SECRET must match Apps Script → Project settings → Script properties."
echo "Generate one: openssl rand -hex 32"
echo
read -r -s -p "Unsubscribe secret: " SECRET
echo

if [[ -z "$SECRET" ]]; then
  echo "Secret is required." >&2
  exit 1
fi

set_env_var "NEWSLETTER_APPS_SCRIPT_URL" "$APPS_URL"
set_env_var "UNSUBSCRIBE_SECRET" "$SECRET"

echo
echo "Saved to $ENV_FILE:"
echo "  NEWSLETTER_APPS_SCRIPT_URL=$APPS_URL"
echo "  UNSUBSCRIBE_SECRET=(hidden)"
echo

if [[ -x "$ROOT/.venv/bin/python" ]]; then
  echo "Sample unsubscribe link (test email):"
  ROOT="$ROOT" "$ROOT/.venv/bin/python" -c "
import os, sys
sys.path.insert(0, os.environ['ROOT'])
from sheet_store import load_dotenv, unsubscribe_url
load_dotenv()
print(' ', unsubscribe_url('you@example.com'))
"
fi

echo
echo "Next steps:"
echo "  1. Apps Script → Script properties → UNSUBSCRIBE_SECRET = same secret"
echo "  2. Redeploy web app if you updated newsletter-subscribe-api.gs"
echo "  3. Test drip: ./run.sh  (or ./install-athena-newsletter.sh test-drip)"
