#!/usr/bin/env python3
"""Render newsletter HTML locally for browser preview (no SMTP, no sheet)."""

from __future__ import annotations

import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent
TEMPLATES = ROOT / "templates"
PREVIEW_DIR = ROOT / "preview"


def wrap_email_html(body: str) -> str:
    shell_path = TEMPLATES / "_email_shell.html"
    if not shell_path.exists():
        return body
    return shell_path.read_text().replace("{{content}}", body)


def render(step_id: str, name: str = "there") -> str:
    html_path = TEMPLATES / f"{step_id}.html"
    if not html_path.exists():
        raise FileNotFoundError(html_path)
    body = html_path.read_text().replace("{{name}}", name)
    return wrap_email_html(body)


def main() -> None:
    default_steps = [
        "branding",
        "product_features",
        "use_cases",
        "product_background",
        "hardware_specs",
        "technical_info",
    ]
    steps = sys.argv[1:] or default_steps
    PREVIEW_DIR.mkdir(exist_ok=True)

    for step_id in steps:
        out = PREVIEW_DIR / f"{step_id}.html"
        out.write_text(render(step_id, "Kean"))
        print(out)

    print(f"\nOpen any file in {PREVIEW_DIR} in your browser.")


if __name__ == "__main__":
    main()
