#!/bin/bash
set -e
cd "$(dirname "$0")"
for test in test_*.sh; do
  echo "--- $test ---"
  bash $test
  echo
  sleep 1
done
echo "[OK] Tous les tests d'intégration sont passés." 