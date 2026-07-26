#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
PAY="$ROOT/payment"
cd "$PAY"
make zip
aws lambda update-function-code \
  --function-name CurlecGateway \
  --zip-file "fileb://$PAY/myFunction.zip" \
  --region ap-southeast-2
echo "Updated CurlecGateway code."
