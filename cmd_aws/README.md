# cmd_aws — Kommu payment Lambda + CDK

Go Lambda (`CurlecGateway`) serving `https://aws.kommu.ai` Curlec/Razorpay Standard Checkout, plus AWS CDK for deploy.

> **Important — source recovery:** The cloud agent that authored this tree did **not** have access to the wiped local `cmd_aws` history (`main.go` / `curlec_standard.go` / `curlec_receipt.go`). This is a **reconstructed** Standard Checkout + Meta CAPI + Sheets + email module that implements the notes-overflow fix and RM1 allowlist. Prefer this workflow before replacing production:
>
> 1. On the Mac that still has Cursor Local History (Jun 15 02:43: `t0eP.go` / `I2QA.go` / `NGLw.go`), restore those three files into your local `cmd_aws/payment/`.
> 2. Port into that restored tree: `checkout_context.go`, `rm1_allowlist.go`, slim `trimNotesMap` / `slimRazorpayNotes`, create-path Dynamo put, fulfill hydrate, and CDK DynamoDB table.
> 3. Run `go test` + `make zip` against the restored+ported tree, then deploy.
>
> Deploying **this** reconstructed zip will replace the Jun 15 production binary; only do that if you accept the narrower feature surface (Standard Checkout paths are covered; legacy Stripe/HTML richness may differ).

> **Repo split:** Cloud `gh` cannot create `kommuai/cmd_aws`. To publish the private org repo from a machine with org admin:
>
> ```bash
> git subtree split -P cmd_aws -b cmd-aws-export
> mkdir -p ../cmd_aws && git archive cmd-aws-export | tar -x -C ../cmd_aws
> cd ../cmd_aws && git init && git add . && git commit -m "Initial import" && \
>   gh repo create kommuai/cmd_aws --private --source=. --remote=origin --push
> ```

## Layout

| Path | Role |
|------|------|
| `payment/` | Go Lambda (`provided.al2023` ARM64 `bootstrap`) |
| `cdk/` | CDK stack: code deploy + `kommu-checkout-context` DynamoDB table |

## Key behaviour

- **Checkout context in DynamoDB** — full cart items + Meta attribution stored by `order_id` / `subscription_id`. Razorpay `notes` stay slim (≤14 keys) so the 15-pair limit never drops fulfillment data.
- **RM1 prod test** — emails `alexanderyeoh@kommu.ai` and `keanwei@kommu.ai` are charged RM1 (one-off) / RM1 deposit + `CURLEC_DEV_TEST_PLAN_ID` (subscriptions). Marked `DEV_TEST`; skips MPA + Meta CAPI (unless test event code set).

## Deploy

```bash
cd payment && make test && make zip
cd ../cdk && npm install && npm run deploy:code
# or full stack (table + code): npm run deploy
```

Set Lambda env (console or CDK): see `payment/.env.example`. Required new vars:

- `CHECKOUT_CONTEXT_TABLE=kommu-checkout-context`
- `CURLEC_DEV_TEST_PLAN_ID` — live Curlec plan id for RM1/month (create once in dashboard)
- `SMTP_PASSWORD` — for receipt email

## Local tests

```bash
cd payment && go test ./...
```
