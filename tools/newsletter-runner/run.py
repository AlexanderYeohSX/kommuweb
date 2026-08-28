#!/usr/bin/env python3
"""
Kommu newsletter drip runner — self-hosted on your own machine.

Reads subscribers from Google Sheet tab "Newsletter", sends the next email in
the sequence via SMTP, and updates sequence_step / last_sent_at.

Run manually or via cron, e.g. every hour:
  0 * * * * cd /path/to/kommuweb/tools/newsletter-runner && ./run.sh
"""

from __future__ import annotations

import json
import os
import re
import smtplib
import sys
from datetime import datetime, timedelta, timezone
from email.mime.multipart import MIMEMultipart
from email.mime.text import MIMEText
from pathlib import Path

try:
    import gspread
    from google.oauth2.service_account import Credentials
except ImportError:
    print("Install dependencies: pip install -r requirements.txt", file=sys.stderr)
    sys.exit(1)

ROOT = Path(__file__).resolve().parent
TEMPLATES = ROOT / "templates"
SEQUENCE_FILE = ROOT.parent.parent / "_data" / "newsletter_sequence.yaml"

SCOPES = [
    "https://www.googleapis.com/auth/spreadsheets",
    "https://www.googleapis.com/auth/drive.readonly",
]

# Sheet column headers (row 1) — must match docs/newsletter-setup.md
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


def load_env() -> dict:
    env_path = ROOT / ".env"
    if env_path.exists():
        for line in env_path.read_text().splitlines():
            line = line.strip()
            if not line or line.startswith("#") or "=" not in line:
                continue
            k, _, v = line.partition("=")
            os.environ.setdefault(k.strip(), v.strip().strip('"').strip("'"))
    required = [
        "GOOGLE_SPREADSHEET_ID",
        "GOOGLE_CREDENTIALS_JSON",
        "SMTP_HOST",
        "SMTP_PORT",
        "SMTP_USER",
        "SMTP_PASSWORD",
        "MAIL_FROM",
    ]
    missing = [k for k in required if not os.environ.get(k)]
    if missing:
        print(f"Missing env: {', '.join(missing)}", file=sys.stderr)
        print(f"Copy config.example.env to .env and fill in values.", file=sys.stderr)
        sys.exit(1)
    return os.environ


def load_sequence() -> list[dict]:
    if not SEQUENCE_FILE.exists():
        print(f"Sequence file not found: {SEQUENCE_FILE}", file=sys.stderr)
        sys.exit(1)
    import yaml

    data = yaml.safe_load(SEQUENCE_FILE.read_text())
    return sorted(data, key=lambda x: x["step"])


def get_sheet(env: dict):
    creds_json = env["GOOGLE_CREDENTIALS_JSON"]
    if creds_json.startswith("{"):
        info = json.loads(creds_json)
    else:
        info = json.loads(Path(creds_json).read_text())
    creds = Credentials.from_service_account_info(info, scopes=SCOPES)
    gc = gspread.authorize(creds)
    tab = env.get("NEWSLETTER_SHEET_TAB", "Newsletter")
    return gc.open_by_key(env["GOOGLE_SPREADSHEET_ID"]).worksheet(tab)


def parse_dt(value: str | None) -> datetime | None:
    if not value or not str(value).strip():
        return None
    s = str(value).strip().replace("Z", "+00:00")
    try:
        return datetime.fromisoformat(s)
    except ValueError:
        for fmt in ("%Y-%m-%d %H:%M:%S", "%Y-%m-%d"):
            try:
                return datetime.strptime(s, fmt).replace(tzinfo=timezone.utc)
            except ValueError:
                continue
    return None


def render_template(step_id: str, name: str) -> tuple[str, str]:
    html_path = TEMPLATES / f"{step_id}.html"
    txt_path = TEMPLATES / f"{step_id}.txt"
    greeting = name.strip() or "there"
    if html_path.exists():
        html = html_path.read_text().replace("{{name}}", greeting)
    else:
        html = f"<p>Hi {greeting},</p><p>(Add template: templates/{step_id}.html)</p>"
    if txt_path.exists():
        text = txt_path.read_text().replace("{{name}}", greeting)
    else:
        text = re.sub(r"<[^>]+>", "", html)
    return html, text


def send_email(env: dict, to: str, subject: str, html: str, text: str) -> None:
    msg = MIMEMultipart("alternative")
    msg["Subject"] = subject
    msg["From"] = env["MAIL_FROM"]
    msg["To"] = to
    msg.attach(MIMEText(text, "plain", "utf-8"))
    msg.attach(MIMEText(html, "html", "utf-8"))

    port = int(env.get("SMTP_PORT", "587"))
    use_tls = env.get("SMTP_TLS", "true").lower() in ("1", "true", "yes")
    with smtplib.SMTP(env["SMTP_HOST"], port, timeout=60) as smtp:
        if use_tls:
            smtp.starttls()
        smtp.login(env["SMTP_USER"], env["SMTP_PASSWORD"])
        smtp.sendmail(env["MAIL_FROM"], [to], msg.as_string())


def row_to_dict(headers: list[str], row: list[str]) -> dict:
    padded = row + [""] * (len(headers) - len(row))
    return dict(zip(headers, padded))


def main() -> None:
    env = load_env()
    sequence = load_sequence()
    sheet = get_sheet(env)
    rows = sheet.get_all_values()
    if not rows:
        print("Sheet is empty")
        return

    headers = [h.strip().lower() for h in rows[0]]
    if headers != HEADERS:
        print(f"Warning: expected headers {HEADERS}, got {headers}")

    now = datetime.now(timezone.utc)
    sent_count = 0

    for idx, row in enumerate(rows[1:], start=2):
        rec = row_to_dict(HEADERS, row)
        email = rec["email"].strip().lower()
        if not email or rec["status"].strip().lower() in ("unsubscribed", "completed"):
            continue

        step_num = int(rec["sequence_step"] or "0")
        next_step = step_num + 1
        if next_step > len(sequence):
            continue

        step = sequence[next_step - 1]
        subscribed_at = parse_dt(rec["subscribed_at"]) or now
        last_sent = parse_dt(rec["last_sent_at"])

        if next_step == 1:
            due_at = subscribed_at
        else:
            delay_days = int(step.get("delay_days", 2))
            base = last_sent or subscribed_at
            due_at = base.replace(hour=0, minute=0, second=0, microsecond=0)
            due_at = due_at + timedelta(days=delay_days)

        if now < due_at:
            continue

        name = rec["name"]
        html, text = render_template(step["id"], name)
        subject = step["subject"]

        print(f"Sending step {next_step} ({step['id']}) to {email}")
        send_email(env, email, subject, html, text)

        iso_now = now.strftime("%Y-%m-%dT%H:%M:%SZ")
        new_status = "completed" if next_step >= len(sequence) else "active"
        col = {h: i + 1 for i, h in enumerate(HEADERS)}
        sheet.update_cell(idx, col["sequence_step"], str(next_step))
        sheet.update_cell(idx, col["last_sent_at"], iso_now)
        sheet.update_cell(idx, col["status"], new_status)
        sheet.update_cell(idx, col["next_send_at"], "")
        sent_count += 1

        dry_run_limit = int(env.get("MAX_SENDS_PER_RUN", "50"))
        if sent_count >= dry_run_limit:
            print(f"Reached MAX_SENDS_PER_RUN={dry_run_limit}")
            break

    print(f"Done. Sent {sent_count} email(s).")


if __name__ == "__main__":
    main()
