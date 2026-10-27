// Shared pieces of load.js and stress.js: sign-in, the Load Test Plan traffic
// mix, and the error metric the performance NFRs are judged on.
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

export const BASE = __ENV.BASE_URL || 'http://localhost:8080';
const JSON_HEADERS = { 'Content-Type': 'application/json' };

// NFR: error rate < 1%. Only 5xx and timeouts (status 0) count; 409 (slot
// taken), 422 (quota reached) and 403 (penalty) are expected booking outcomes.
export const errors = new Rate('errors');
export const rateLimited = new Rate('rate_limited');

function record(res) {
  errors.add(res.status === 0 || res.status >= 500);
  rateLimited.add(res.status === 429);
  return res;
}

function post(url, body, token, kind) {
  const headers = token ? { ...JSON_HEADERS, Authorization: `Bearer ${token}` } : JSON_HEADERS;
  return record(http.post(url, body === null ? null : JSON.stringify(body), { headers, tags: { kind } }));
}

function get(url, token, kind, name) {
  const params = { tags: { kind, name: name || url } };
  if (token) params.headers = { Authorization: `Bearer ${token}` };
  return record(http.get(url, params));
}

// Anonymous requests share one rate-limit bucket per IP at the gateway, so many
// VUs signing in at once get 429; back off and retry instead of failing the VU.
function signIn(path, body) {
  for (let i = 0; i < 30; i++) {
    const res = post(`${BASE}${path}`, body, null, 'auth');
    if (res.status === 200) return res.json('token');
    if (res.status !== 429 && res.status !== 0 && res.status < 500) {
      throw new Error(`login ${path} failed: ${res.status} ${res.body}`);
    }
    sleep(1 + Math.random());
  }
  throw new Error(`login ${path} kept failing`);
}

export const login = (studentId) => signIn('/api/auth/login', { student_id: studentId });
export const adminLogin = () =>
  signIn('/api/auth/admin/login', {
    email: __ENV.ADMIN_EMAIL || 'admin@cu.ac.th',
    password: __ENV.ADMIN_PASSWORD || 'admin123',
  });

// setup(): library and room IDs for the mix, and an admin token for accepting.
// Bookings use only the first ROOMS rooms so concurrent requests fight over slots.
export function setupData() {
  const libs = http.get(`${BASE}/api/libraries`).json();
  if (!libs.length) throw new Error('no libraries seeded');
  const rooms = [];
  for (const l of libs) {
    for (const r of http.get(`${BASE}/api/libraries/${l.id}/rooms`).json()) rooms.push(r.id);
  }
  if (!rooms.length) throw new Error('no rooms seeded');
  return {
    libraries: libs.map((l) => l.id),
    rooms: rooms.slice(0, Number(__ENV.ROOMS || 3)),
    adminToken: adminLogin(),
  };
}

const pick = (a) => a[Math.floor(Math.random() * a.length)];
const TAGS = ['quiet', 'clean', 'strong_wifi', 'many_outlets', 'crowded'];

// One hour-long slot, 1..7 days ahead, 08:00..19:00 UTC: few enough slots that
// VUs collide, so the double-booking guard is exercised under load.
function randomSlot() {
  const d = new Date();
  d.setUTCDate(d.getUTCDate() + 1 + Math.floor(Math.random() * 7));
  d.setUTCHours(8 + Math.floor(Math.random() * 12), 0, 0, 0);
  return d.toISOString().replace('.000Z', 'Z');
}

// Per VU: each VU is its own signed-in student and, like the web app, sends its
// token on every call, so it gets its own rate-limit bucket at the gateway.
let token;

// One iteration of the Load Test Plan traffic mix (60/20/15/5).
export function userJourney(data) {
  if (!token) token = login(`${__ENV.ID_PREFIX || 'lt'}${__VU}`);
  const p = Math.random();

  if (p < 0.6) {
    const res = get(`${BASE}/api/occupancy`, token, 'read', 'GET /api/occupancy');
    check(res, { 'occupancy 200': (r) => r.status === 200 });
  } else if (p < 0.8) {
    const lib = pick(data.libraries);
    const a = get(`${BASE}/api/libraries/${lib}`, token, 'read', 'GET /api/libraries/{id}');
    const b = get(`${BASE}/api/reviews?library_id=${lib}`, token, 'read', 'GET /api/reviews');
    check(a, { 'library 200': (r) => r.status === 200 });
    check(b, { 'reviews 200': (r) => r.status === 200 });
  } else if (p < 0.95) {
    const res = post(
      `${BASE}/api/reservations`,
      { room_id: pick(data.rooms), start_time: randomSlot(), duration_minutes: 60 },
      token,
      'reservation',
    );
    check(res, { 'reserve 201/409/422': (r) => [201, 409, 422].includes(r.status) });
    if (res.status === 201) {
      // Admin accepts, then half are cancelled so the customer's quota (max 2
      // active) frees up and the create path keeps getting exercised.
      const id = res.json('id');
      const acc = post(`${BASE}/api/reservations/${id}/accept`, null, data.adminToken, 'reservation');
      check(acc, { 'accept 200': (r) => r.status === 200 });
      if (Math.random() < 0.5) {
        const c = post(`${BASE}/api/reservations/${id}/cancel`, null, token, 'reservation');
        check(c, { 'cancel 200': (r) => r.status === 200 });
      }
    }
  } else {
    const res = post(
      `${BASE}/api/reviews`,
      { library_id: pick(data.libraries), rating: 1 + Math.floor(Math.random() * 5), tags: [pick(TAGS)], comment: 'load test' },
      token,
      'write',
    );
    check(res, { 'review 201': (r) => r.status === 201 });
  }
  sleep(0.5 + Math.random()); // think time
}
