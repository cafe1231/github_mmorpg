#!/bin/bash
set -e
# Inscription et login
URL_REG="http://localhost:8081/api/v1/auth/register"
URL_LOGIN="http://localhost:8081/api/v1/auth/login"
URL_EVENT="http://localhost:8088/api/v1/analytics/events"
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
echo "Enregistrement événement: $URL_EVENT"
EVENT_BODY='{"type":"login","payload":"{\"ip\":\"127.0.0.1\"}"}'
resp2=$(curl -s -w "%{http_code}" -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d "$EVENT_BODY" $URL_EVENT)
body2=$(echo $resp2 | head -c -3)
code2=$(echo $resp2 | tail -c 3)
if [ "$code2" != "200" ] && [ "$code2" != "201" ]; then
  echo "[FAIL] Code HTTP: $code2"; exit 1
fi
echo "$body2" | grep -E 'event|id|type|success' && echo "[OK] /analytics/events (create)" || (echo "[FAIL] Body: $body2"; exit 1) 