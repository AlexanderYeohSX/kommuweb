#!/usr/bin/env python3
"""Copy purchaser emails from the orders tab into Newsletter (source=checkout)."""

from __future__ import annotations

import json
import os
import re
from pathlib import Path

try:
    import gspread
except ImportError:
    raise SystemExit("pip install -r requirements.txt")

from sheet_store import get_credentials, upsert_subscriber, load_dotenv, normalize_email

EMAIL_RE = re.compile(r"^[^\s@]+@[^\s@]+\.[^\s@]+$")


def get_orders_sheet():
    load_dotenv()
    sheet_id = os.environ["GOOGLE_SPREADSHEET_ID"]
    tab = os.environ.get("ORDERS_SHEET_TAB", "Orders")
    gc = gspread.authorize(get_credentials())
    return gc.open_by_key(sheet_id).worksheet(tab)


def find_email_column(headers: list[str]) -> int | None:
    for i, h in enumerate(headers):
        if h.strip().lower() in ("email", "customer email", "customer_email", "e-mail"):
            return i
    return None


def find_name_column(headers: list[str]) -> int | None:
    for i, h in enumerate(headers):
        if h.strip().lower() in ("name", "customer name", "customer_name"):
            return i
    return None


def main() -> None:
    load_dotenv()
    orders = get_orders_sheet()
    rows = orders.get_all_values()
    if not rows:
        print("Orders sheet is empty")
        return

    headers = [h.strip().lower() for h in rows[0]]
    email_col = find_email_column(headers)
    name_col = find_name_column(headers)
    if email_col is None:
        print(f"No email column in orders tab. Headers: {rows[0]}", flush=True)
        return

    added = 0
    seen: set[str] = set()
    for row in rows[1:]:
        if email_col >= len(row):
            continue
        email = normalize_email(row[email_col])
        if not email or email in seen or not EMAIL_RE.match(email):
            continue
        seen.add(email)
        name = row[name_col].strip() if name_col is not None and name_col < len(row) else ""
        result = upsert_subscriber(email, name, "checkout")
        if result.get("created"):
            added += 1
            print(f"Added {email}")

    print(f"Done. {added} new subscriber(s) from orders.")


if __name__ == "__main__":
    main()
