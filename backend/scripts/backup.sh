#!/bin/sh
# Dumps Postgres (user, library, occupancy, reservation) and MongoDB (reviews)
# into backups/<timestamp>/ at the repo root. Operational NFR: run daily, e.g.
#   0 2 * * * cd /path/to/cu-seat-finder/backend && make backup
# Production uses the managed hosts' own backups instead (ADR-006, ADR-008).
set -eu
cd "$(dirname "$0")/../.."
dir="backups/$(date +%Y%m%d-%H%M%S)"
mkdir -p "$dir"
docker compose exec -T postgres pg_dump -U seat -Fc seatfinder > "$dir/postgres.dump"
docker compose exec -T mongo mongodump --quiet --db seatfinder --archive --gzip > "$dir/mongo.archive.gz"
ls -lh "$dir"
