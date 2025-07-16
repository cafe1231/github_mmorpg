#!/bin/bash
set -e
# Inscription et login
URL_REG="http://localhost:8081/api/v1/auth/register"
URL_LOGIN="http://localhost:8081/api/v1/auth/login"
URL_CHAT="http://localhost:8087/api/v1/chat/channels"
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
resp=$(curl -s -H "Content-Type: application/json" -d "$BODY_LOGIN" $URL_LOGIN)
TOKEN=$(echo $resp | grep -oE '"token":"[^"]+' | cut -d'"' -f4)
if [ -z "$TOKEN" ]; then echo "[FAIL] Pas de token"; exit 1; fi
echo "Récupération canaux: $URL_CHAT"
resp2=$(curl -s -w "%{http_code}" -H "Authorization: Bearer $TOKEN" $URL_CHAT)
body2=$(echo $resp2 | head -c -3)
code2=$(echo $resp2 | tail -c 3)
if [ "$code2" != "200" ]; then
  echo "[FAIL] Code HTTP: $code2"; exit 1
fi
echo "$body2" | grep -E 'channels|id|name' && echo "[OK] /chat/channels" || (echo "[FAIL] Body: $body2"; exit 1) 