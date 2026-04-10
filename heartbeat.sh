#!/bin/bash

PING_URL="http://localhost:8080/api/ping/f43e62ce-7a0f-4cb0-ae2b-81ee48f1bda6"
INTERVAL=10  # secondes entre chaque ping

echo "🟢 Heartbeat simulator démarré"
echo "   URL : $PING_URL"
echo "   Intervalle : ${INTERVAL}s"
echo "   Ctrl+C pour arrêter"
echo ""

while true; do
  RESPONSE=$(curl -fsS -o /dev/null -w "%{http_code}" "$PING_URL" 2>&1)

  if [ "$RESPONSE" = "200" ]; then
    echo "✅ $(date '+%H:%M:%S') — Ping envoyé → HTTP $RESPONSE"
  else
    echo "❌ $(date '+%H:%M:%S') — Échec → HTTP $RESPONSE"
  fi

  sleep $INTERVAL
done
