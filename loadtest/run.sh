#!/bin/sh
# Runs one k6 script against the stack started by `make up`, then the
# double-booking check. Usage: loadtest/run.sh load|stress|security
# Results land in loadtest/results/<test>-<timestamp>.{json,html,stats.txt}:
# the k6 summary (req/s, p50/p95/p99 per tag, error rate), an HTML report with
# charts over time (open in a browser), and docker stats
# (CPU and memory per container) sampled every 5 s during the run.
# Uses a local k6 if installed, otherwise the grafana/k6 image.
set -eu
test="${1:?usage: run.sh load|stress|security}"
cd "$(dirname "$0")"
mkdir -p results
out="results/$test-$(date +%Y%m%d-%H%M%S)"

if command -v k6 >/dev/null; then
  k6="k6"
else
  k6="docker run --rm -i --network host -u $(id -u):$(id -g) -v $PWD:/loadtest -w /loadtest grafana/k6:1.3.0"
fi

if [ "$test" != security ]; then
  (while :; do docker stats --no-stream --format '{{.Name}} {{.CPUPerc}} {{.MemUsage}}' >> "$out.stats.txt"; sleep 5; done) &
  sampler=$!
  trap 'kill $sampler 2>/dev/null || true' EXIT
fi

status=0
$k6 run -e BASE_URL="${BASE_URL:-http://localhost:8080}" --summary-export "$out.json" --out "web-dashboard=export=$out.html" "$test.js" || status=$?

if [ "$test" != security ]; then
  overlaps=$(docker compose exec -T postgres psql -U seat -d seatfinder -tA -f - < overlap.sql | wc -l)
  echo "overlapping active reservations: $overlaps"
  [ "$overlaps" -eq 0 ] || status=1
fi
echo "results: loadtest/$out.*"
exit $status
