#!/usr/bin/env python3
"""One-time import of existing customer emails into the Newsletter sheet tab."""

from __future__ import annotations

import argparse
import csv

from sheet_store import load_dotenv, upsert_subscriber


def main() -> None:
    parser = argparse.ArgumentParser(description="Import emails into Newsletter sheet")
    parser.add_argument("--csv", required=True, help="CSV with email and optional name columns")
    parser.add_argument("--source", default="import", choices=["import", "checkout", "homepage"])
    parser.add_argument("--dry-run", action="store_true")
    args = parser.parse_args()

    load_dotenv()
    added = 0
    skipped = 0

    with open(args.csv, newline="", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        for row in reader:
            email = (row.get("email") or row.get("Email") or "").strip().lower()
            name = (row.get("name") or row.get("Name") or "").strip()
            if not email:
                continue
            if args.dry_run:
                print(f"Would upsert: {email}")
                added += 1
                continue
            result = upsert_subscriber(email, name, args.source)
            if result.get("created"):
                added += 1
                print(f"Added {email}")
            else:
                skipped += 1

    print(f"Added {added}, skipped {skipped} (already in sheet)")


if __name__ == "__main__":
    main()
