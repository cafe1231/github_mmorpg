#!/bin/bash
set -e
URL="http://localhost:8081/api/v1/auth/register"
RAND=$RANDOM
EMAIL="testuser$RAND@example.com"
BODY='{"email":"'$EMAIL'","password":"Test1234!","username":"testuser'$RAND'"}'
echo "Test: $URL"
resp=$(curl -s -w "%{http_code}" -H "Content-Type: application/json" -d "$BODY" $URL)
body=$(echo $resp | head -c -3)
code=$(echo $resp | tail -c 3)
if [ "$code" != "200" ] && [ "$code" != "201" ]; then
  echo "[FAIL] Code HTTP: $code"; exit 1
fi
echo "$body" | grep -E 'user|id|success|token' && echo "[OK] /auth/register" || (echo "[FAIL] Body: $body"; exit 1) 