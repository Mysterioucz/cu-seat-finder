#!/bin/sh
# Operational and availability NFRs against the stack started by `make up`.
# Usage: loadtest/resilience.sh [uptime|freshness|backup]   (default: all three)
#
#   uptime    - crash one of two Library replicas while traffic runs; the
#               container must restart and turn healthy within 30 s, and
#               fewer than 1% of requests through the gateway may fail.
#   freshness - occupancy data is never older than two poll intervals.
#   backup    - scripts/backup.sh output restores into scratch databases with
#               the same row counts as the live ones.
set -u
cd "$(dirname "$0")/.."
BASE="${BASE_URL:-http://localhost:8080}"
POLL_INTERVAL="${POLL_INTERVAL:-30}"
fail=0
pass() { echo "PASS $*"; }
bad() { echo "FAIL $*"; fail=1; }

uptime() {
  echo "== uptime: crash a Library replica under traffic"
  docker compose --profile app up -d --scale library=2 --no-recreate >/dev/null 2>&1
  c=$(docker compose ps -q library | head -1)
  i=0
  until [ "$(docker compose ps -q library | xargs docker inspect -f '{{.State.Health.Status}}' | grep -c healthy)" -ge 2 ]; do
    i=$((i + 1)); [ $i -gt 60 ] && { bad "two healthy library replicas never came up"; return; }
    sleep 1
  done
  sleep 6 # let Consul's 5 s check see both replicas

  log=$(mktemp)
  # ~10 req/s for 60 s; GET /api/libraries goes gateway -> Library over gRPC.
  (end=$(($(date +%s) + 60)); while [ "$(date +%s)" -lt $end ]; do
    curl -s -o /dev/null -m 5 -w '%{http_code}\n' "$BASE/api/libraries" >> "$log"; sleep 0.1
  done) &
  traffic=$!

  sleep 10
  before=$(docker inspect -f '{{.RestartCount}}' "$c")
  # SIGQUIT makes the Go runtime abort (exit 2) without deregistering from
  # Consul: a real crash. `docker kill` would count as a manual stop and the
  # restart policy would (correctly) not bring it back.
  docker exec "$c" kill -QUIT 1
  t0=$(date +%s)
  until [ "$(docker inspect -f '{{.RestartCount}} {{.State.Health.Status}}' "$c")" = "$((before + 1)) healthy" ]; do
    [ $(($(date +%s) - t0)) -gt 60 ] && break
    sleep 1
  done
  recovered=$(($(date +%s) - t0))
  wait $traffic

  total=$(wc -l < "$log")
  errs=$(grep -vc '^200$' "$log")
  rm -f "$log"
  echo "restart + healthy after ${recovered}s; $errs/$total requests failed"
  [ "$recovered" -le 30 ] && pass "crashed replica back and healthy in ${recovered}s (<= 30s)" || bad "replica took ${recovered}s to recover"
  [ $((errs * 100)) -lt "$total" ] && pass "error rate during crash < 1%" || bad "error rate during crash >= 1% ($errs/$total)"
}

freshness() {
  window=$((POLL_INTERVAL * 5 / 2))
  limit=$((POLL_INTERVAL * 2))
  echo "== freshness: sample occupancy age for ${window}s (limit ${limit}s)"
  max=0
  end=$(($(date +%s) + window))
  while [ "$(date +%s)" -lt $end ]; do
    age=$(curl -s -m 5 "$BASE/api/occupancy" | jq '[.items[].updated_at | sub("\\.[0-9]+"; "") | fromdateiso8601] | now - min | floor')
    [ -n "$age" ] && [ "$age" -gt "$max" ] && max=$age
    sleep 2
  done
  echo "oldest reading seen: ${max}s"
  [ "$max" -le "$limit" ] && pass "occupancy never older than ${limit}s" || bad "occupancy reading was ${max}s old (> ${limit}s)"
}

backup() {
  echo "== backup: dump, restore into scratch databases, compare counts"
  backend/scripts/backup.sh >/dev/null || { bad "backup.sh failed"; return; }
  dir=$(ls -d backups/*/ | tail -1)
  psql() { docker compose exec -T postgres psql -U seat -tA "$@"; }
  mongo() { docker compose exec -T mongo mongosh --quiet "$@"; }

  psql -d postgres -c 'DROP DATABASE IF EXISTS seatfinder_restore' -c 'CREATE DATABASE seatfinder_restore' >/dev/null
  docker compose exec -T postgres pg_restore -U seat -d seatfinder_restore --no-owner < "$dir/postgres.dump" ||
    bad "pg_restore reported errors"
  for t in users.users users.penalties library.libraries library.rooms reservation.reservations occupancy.readings; do
    live=$(psql -d seatfinder -c "SELECT count(*) FROM $t")
    restored=$(psql -d seatfinder_restore -c "SELECT count(*) FROM $t")
    [ -n "$live" ] && [ "$live" = "$restored" ] && pass "$t restored ($restored rows)" || bad "$t: live $live, restored $restored"
  done
  psql -d postgres -c 'DROP DATABASE seatfinder_restore' >/dev/null

  docker compose exec -T mongo mongorestore --quiet --archive --gzip --drop \
    --nsFrom 'seatfinder.*' --nsTo 'seatfinder_restore.*' < "$dir/mongo.archive.gz" || bad "mongorestore failed"
  live=$(mongo seatfinder --eval 'db.reviews.countDocuments()')
  restored=$(mongo seatfinder_restore --eval 'db.reviews.countDocuments()')
  [ "$live" = "$restored" ] && pass "reviews restored ($restored documents)" || bad "reviews: live $live, restored $restored"
  mongo seatfinder_restore --eval 'db.dropDatabase()' >/dev/null
}

for step in ${1:-uptime freshness backup}; do $step; done
exit $fail
