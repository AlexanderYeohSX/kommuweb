# kommuweb

Jekyll site for [kommu.ai](https://kommu.ai).

## Local development

```bash
bundle install
bundle exec jekyll serve
```

## Payments

- **MYR**: Razorpay Curlec Standard Checkout via [`server/`](server/) API on `aws.kommu.ai`
- **International**: Stripe redirect (unchanged)

Documentation: [`docs/README.md`](docs/README.md)

## API server

```bash
cd server && npm install && cp .env.example .env && npm start
```
