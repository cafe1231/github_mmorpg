#!/bin/bash
set -e
URL="http://localhost:8083/api/v1/world/zones"
echo "Test: $URL"
resp=$(curl -s -w "%{http_code}" $URL)
body=$(echo $resp | head -c -3)
code=$(echo $resp | tail -c 3)
if [ "$code" != "200" ]; then
  echo "[FAIL] Code HTTP: $code"; exit 1
fi
echo "$body" | grep -E 'zone|id|name' && echo "[OK] /world/zones" || (echo "[FAIL] Body: $body"; exit 1) 