#!/bin/bash
set -e
# Inscription et login
URL_REG="http://localhost:8081/api/v1/auth/register"
URL_LOGIN="http://localhost:8081/api/v1/auth/login"
URL_CREATE="http://localhost:8082/api/v1/player/characters"
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
echo "Création personnage: $URL_CREATE"
CHAR_BODY='{"name":"Hero'$RAND'","class":"warrior","race":"human"}'
resp2=$(curl -s -w "%{http_code}" -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d "$CHAR_BODY" $URL_CREATE)
body2=$(echo $resp2 | head -c -3)
code2=$(echo $resp2 | tail -c 3)
if [ "$code2" != "200" ] && [ "$code2" != "201" ]; then
  echo "[FAIL] Code HTTP: $code2"; exit 1
fi
echo "$body2" | grep -E 'character|id|name' && echo "[OK] /player/characters (create)" || (echo "[FAIL] Body: $body2"; exit 1) 