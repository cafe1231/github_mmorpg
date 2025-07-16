#!/bin/bash
set -e
URL="http://localhost:8080/gateway/status"
echo "Test: $URL"
resp=$(curl -s -w "%{http_code}" $URL)
body=$(echo $resp | head -c -3)
code=$(echo $resp | tail -c 3)
if [ "$code" != "200" ]; then
  echo "[FAIL] Code HTTP: $code"; exit 1
fi
echo "$body" | grep -q 'status' && echo "[OK] /gateway/status" || (echo "[FAIL] Body: $body"; exit 1) 