#!/bin/bash
set -e
URL_REG="http://localhost:8081/api/v1/auth/register"
URL_LOGIN="http://localhost:8081/api/v1/auth/login"
RAND=$RANDOM
EMAIL="testuser$RAND@example.com"
USER="testuser$RAND"
PASS="Test1234!"
BODY_REG='{"email":"'$EMAIL'","password":"'$PASS'","username":"'$USER'"}'
BODY_LOGIN='{"email":"'$EMAIL'","password":"'$PASS'"}'
echo "Inscription: $URL_REG"
curl -s -H "Content-Type: application/json" -d "$BODY_REG" $URL_REG > /dev/null
sleep 1
echo "Login: $URL_LOGIN"
resp=$(curl -s -w "%{http_code}" -H "Content-Type: application/json" -d "$BODY_LOGIN" $URL_LOGIN)
body=$(echo $resp | head -c -3)
code=$(echo $resp | tail -c 3)
if [ "$code" != "200" ]; then
  echo "[FAIL] Code HTTP: $code"; exit 1
fi
echo "$body" | grep -E 'token|access|jwt' && echo "[OK] /auth/login" || (echo "[FAIL] Body: $body"; exit 1) 