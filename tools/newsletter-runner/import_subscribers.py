#!/usr/bin/env python3
"""One-time import of existing customer emails into the Newsletter sheet tab."""

from __future__ import annotations

import argparse
import csv
import json
import sys
from datetime import datetime, timezone
from pathlib import Path

try:
    import gspread
    from google.oauth2.service_account import Credentials
except ImportError:
    print("pip install -r requirements.txt", file=sys.stderr)
    sys.exit(1)

ROOT = Path(__file__).resolve().parent
HEADERS = [
    "email",
    "name",
    "source",
    "subscribed_at",
    "sequence_step",
    "last_sent_at",
    "status",
    "next_send_at",
]
SCOPES = [
    "https://www.googleapis.com/auth/spreadsheets",
    "https://www.googleapis.com/auth/drive.readonly",
]


def load_env() -> dict:
    import os

    env_path = ROOT / ".env"
    if env_path.exists():
        for line in env_path.read_text().splitlines():
            line = line.strip()
            if not line or line.startswith("#") or "=" not in line:
                continue
            k, _, v = line.partition("=")
            os.environ.setdefault(k.strip(), v.strip().strip('"').strip("'"))
    return os.environ


def get_sheet(env: dict):
    creds_json = env["GOOGLE_CREDENTIALS_JSON"]
    info = json.loads(creds_json) if creds_json.startswith("{") else json.loads(Path(creds_json).read_text())
    creds = Credentials.from_service_account_info(info, scopes=SCOPES)
    gc = gspread.authorize(creds)
    tab = env.get("NEWSLETTER_SHEET_TAB", "Newsletter")
    return gc.open_by_key(env["GOOGLE_SPREADSHEET_ID"]).worksheet(tab)


def main() -> None:
    parser = argparse.ArgumentParser(description="Import emails into Newsletter sheet")
    parser.add_argument("--csv", required=True, help="CSV with email and optional name columns")
    parser.add_argument("--source", default="import", choices=["import", "checkout"])
    parser.add_argument("--dry-run", action="store_true")
    args = parser.parse_args()

    env = load_env()
    sheet = get_sheet(env)
    existing = {r[0].strip().lower() for r in sheet.get_all_values()[1:] if r and r[0].strip()}
    now = datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
    added = 0
    skipped = 0

    with open(args.csv, newline="", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        for row in reader:
            email = (row.get("email") or row.get("Email") or "").strip().lower()
            name = (row.get("name") or row.get("Name") or "").strip()
            if not email:
                continue
            if email in existing:
                skipped += 1
                continue
            line = [email, name, args.source, now, "0", "", "active", ""]
            if args.dry_run:
                print(f"Would add: {email}")
            else:
                sheet.append_row(line, value_input_option="USER_ENTERED")
            existing.add(email)
            added += 1

    print(f"Added {added}, skipped {skipped} (already in sheet)")


if __name__ == "__main__":
    main()
