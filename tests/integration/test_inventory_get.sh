#!/bin/bash
set -e
# Inscription et login
URL_REG="http://localhost:8081/api/v1/auth/register"
URL_LOGIN="http://localhost:8081/api/v1/auth/login"
URL_CHAR="http://localhost:8082/api/v1/player/characters"
URL_INV_BASE="http://localhost:8084/api/v1/inventory/"
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
echo "Création personnage: $URL_CHAR"
CHAR_BODY='{"name":"Hero'$RAND'","class":"warrior","race":"human"}'
resp2=$(curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d "$CHAR_BODY" $URL_CHAR)
CHAR_ID=$(echo $resp2 | grep -oE '"id":"[^"]+' | head -n1 | cut -d'"' -f4)
if [ -z "$CHAR_ID" ]; then echo "[FAIL] Pas d'id de personnage"; exit 1; fi
URL_INV="$URL_INV_BASE$CHAR_ID"
echo "Récupération inventaire: $URL_INV"
resp3=$(curl -s -w "%{http_code}" -H "Authorization: Bearer $TOKEN" $URL_INV)
body3=$(echo $resp3 | head -c -3)
code3=$(echo $resp3 | tail -c 3)
if [ "$code3" != "200" ]; then
  echo "[FAIL] Code HTTP: $code3"; exit 1
fi
echo "$body3" | grep -E 'items|inventory|id' && echo "[OK] /inventory/{characterId}" || (echo "[FAIL] Body: $body3"; exit 1) 