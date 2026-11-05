#!/bin/sh
# Live-demo helpers against the stack from `make up`.
#   demo.sh seed        sample reviews, plus a pending booking for the admin to accept
#   demo.sh ticket      a new student with an accepted booking inside its check-in window
#                       (run within 20 minutes of the demo: the window closes 25 minutes later)
#   demo.sh noshow REF  expire REF's check-in deadline so the demo needn't wait 15 minutes;
#                       the sweeper publishes reservation.no_show within a minute and the
#                       User service applies the penalty. Unban from the admin page afterwards.
# noshow writes to the reservations table directly: a demo shortcut, not a service API.
set -eu
cd "$(dirname "$0")/../.."
BASE="${BASE_URL:-http://localhost:8080}"

post() { # path body [token]
  curl -s -w '\n%{http_code}' -X POST "$BASE$1" -H 'content-type: application/json' ${3:+-H "authorization: Bearer $3"} -d "$2"
}
body() { sed '$d'; }
code() { tail -1; }
student() { post /api/auth/login "{\"student_id\":\"$1\"}" | body | jq -r .token; }
admin() { post /api/auth/admin/login '{"email":"admin@cu.ac.th","password":"admin123"}' | body | jq -r .token; }

# book TOKEN START: tries each room until one has the slot free; prints the reservation.
book() {
  for lib in $(curl -s "$BASE/api/libraries" | jq -r '.[].id'); do
    for room in $(curl -s "$BASE/api/libraries/$lib/rooms" | jq -r '.[].id'); do
      res=$(post /api/reservations "{\"room_id\":\"$room\",\"start_time\":\"$2\",\"duration_minutes\":60}" "$1")
      if [ "$(echo "$res" | code)" = 201 ]; then
        echo "$res" | body
        return 0
      fi
    done
  done
  echo "no room could be booked: $(echo "$res" | body)" >&2
  return 1
}

seed() {
  token=$(student demo-reviewer)
  if [ "$(curl -s "$BASE/api/reviews?library_id=lib-central" | jq .count)" = 0 ]; then
    for r in '4|["quiet","strong_wifi"]|เงียบ อ่านหนังสือได้ดี Wi-Fi แรง' '5|["clean","many_outlets"]|สะอาด ปลั๊กเยอะมาก' '2|["crowded"]|ช่วงสอบคนเยอะมาก หาที่นั่งยาก'; do
      rating=${r%%|*}; rest=${r#*|}; tags=${rest%%|*}; comment=${rest#*|}
      post /api/reviews "{\"library_id\":\"lib-central\",\"rating\":$rating,\"tags\":$tags,\"comment\":\"$comment\"}" "$token" >/dev/null
    done
    echo "reviews added to lib-central"
  fi
  id=$(jq -nr 'now | . + 86400 | floor | todate | .[0:10]')
  book "$(student demo-pending)" "${id}T03:00:00Z" | jq -r '"pending booking: room \(.room_id), student demo-pending"'
}

ticket() {
  who="demo-$(date +%H%M%S)"
  token=$(student "$who")
  # Starts in 10 minutes: check-in opens 15 minutes before the start.
  res=$(book "$token" "$(jq -nr 'now + 600 | floor | todate')")
  id=$(echo "$res" | jq -r .id)
  post "/api/reservations/$id/accept" '' "$(admin)" | body |
    jq -r --arg who "$who" '"sign in as student \($who) to see the QR ticket\nreference: \(.reference)\nroom: \(.room_id)"'
}

noshow() {
  ref="${1:?usage: demo.sh noshow CU-XXXXXX}"
  updated=$(docker compose exec -T postgres psql -U seat -d seatfinder -tA -c \
    "UPDATE reservation.reservations SET checkin_deadline = now() - interval '1 second'
     WHERE reference = '$ref' AND status = 'accepted' AND checked_in_at IS NULL RETURNING user_id" | sed -n 1p)
  [ -n "$updated" ] || { echo "no accepted, not-yet-checked-in booking with reference $ref" >&2; exit 1; }
  echo "deadline expired, waiting for the sweeper and RabbitMQ (up to a minute)..."
  token=$(admin)
  i=0
  until curl -s "$BASE/api/users/penalized" -H "authorization: Bearer $token" | jq -e --arg id "$updated" 'any(.[]; .id == $id)' >/dev/null; do
    i=$((i + 1)); [ $i -gt 45 ] && { echo "penalty did not appear after 90 s" >&2; exit 1; }
    sleep 2
  done
  echo "penalty applied: refresh the admin page"
}

case "${1:-}" in
  seed) seed ;;
  ticket) ticket ;;
  noshow) noshow "${2:-}" ;;
  *) echo "usage: demo.sh seed|ticket|noshow REF" >&2; exit 2 ;;
esac
