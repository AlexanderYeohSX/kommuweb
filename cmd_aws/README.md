# cmd_aws — Kommu payment Lambda + CDK

Go Lambda (`CurlecGateway`) serving `https://aws.kommu.ai` Curlec/Razorpay Standard Checkout, plus AWS CDK for deploy.

> **Repo split:** This tree currently lives inside [kommuweb](https://github.com/AlexanderYeohSX/kommuweb) because the cloud agent cannot create `kommuai/cmd_aws`. To publish the private org repo from a machine with org admin + `gh` write access:
>
> ```bash
> # from kommuweb root
> git subtree split -P cmd_aws -b cmd-aws-export
> gh repo create kommuai/cmd_aws --private --source=. --remote=cmdaws
> # or: mkdir ../cmd_aws && git archive cmd-aws-export | tar -x -C ../cmd_aws
> cd /path/to/cmd_aws && git init && git add payment cdk README.md && git commit -m "Initial import" && \
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
