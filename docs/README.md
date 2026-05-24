# Kommuweb documentation

Architecture and checkout documentation for the Jekyll storefront and Curlec (Razorpay) payment API.

## Contents

| Document | Description |
|----------|-------------|
| [architecture.md](architecture.md) | System overview and data files |
| [checkout-flow.md](checkout-flow.md) | Product/cart checkout and payment routing |
| [backend-api.md](backend-api.md) | `aws.kommu.ai` Curlec API contract |
| [operations.md](operations.md) | Dashboard, env vars, deployment |
| [qa-checklist.md](qa-checklist.md) | Manual regression checklist |

## Build PDF

Requires [Pandoc](https://pandoc.org/) and a PDF engine (e.g. `pdflatex` or `wkhtmltopdf`).

```bash
./docs/scripts/build-pdf.sh
```

Output: `docs/output/kommu-architecture.pdf`

Without Pandoc, read the Markdown files directly or use your editor’s export to PDF.
