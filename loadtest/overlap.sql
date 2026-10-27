-- Run after every load/stress test: must return 0 rows (no meeting room double-booked).
SELECT a.room_id, a.id AS a_id, b.id AS b_id, a.start_time, a.end_time, b.start_time, b.end_time
FROM reservation.reservations a
JOIN reservation.reservations b
  ON a.room_id = b.room_id AND a.id < b.id
 AND tstzrange(a.start_time, a.end_time, '[)') && tstzrange(b.start_time, b.end_time, '[)')
WHERE a.status IN ('pending', 'accepted') AND b.status IN ('pending', 'accepted');
