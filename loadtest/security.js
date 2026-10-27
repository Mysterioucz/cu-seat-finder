// Security NFR, black-box through the API Gateway: every protected route
// rejects a missing or forged JWT (401) and the wrong role (403), and services
// check the role again behind the gateway. One VU, one pass; any failed check
// fails the run.
import http from 'k6/http';
import { check } from 'k6';
import { BASE, login, adminLogin } from './lib.js';

export const options = {
  vus: 1,
  iterations: 1,
  thresholds: { checks: ['rate==1'] },
};

const X = 'nfr-nonexistent';

// [method, path, who may call it] — the gateway access table (Proposal Step 4,
// backend/cmd/gateway/routes.go accessFor).
const ROUTES = [
  ['GET', '/api/occupancy', 'public'],
  ['GET', '/api/libraries', 'public'],
  ['GET', '/api/reviews', 'public'],
  ['GET', '/api/users/me', 'signedIn'],
  ['GET', '/api/reservations/me', 'signedIn'],
  ['POST', '/api/reviews', 'signedIn'],
  ['POST', `/api/reservations/${X}/cancel`, 'signedIn'],
  ['GET', '/api/users/penalized', 'admin'],
  ['POST', `/api/users/${X}/unban`, 'admin'],
  ['GET', '/api/reservations', 'admin'],
  ['POST', `/api/reservations/${X}/accept`, 'admin'],
  ['POST', `/api/reservations/${X}/reject`, 'admin'],
  ['POST', '/api/reservations/checkin', 'admin'],
  ['GET', '/api/reviews/flagged', 'admin'],
  ['PATCH', `/api/reviews/${X}/moderate`, 'admin'],
  ['PUT', `/api/libraries/${X}`, 'admin'],
  ['DELETE', `/api/libraries/${X}`, 'admin'],
  ['PUT', `/api/rooms/${X}`, 'admin'],
  ['DELETE', `/api/rooms/${X}`, 'admin'],
];

function call(method, path, token) {
  const headers = { 'Content-Type': 'application/json' };
  if (token) headers.Authorization = `Bearer ${token}`;
  return http.request(method, `${BASE}${path}`, method === 'GET' ? null : '{}', { headers });
}

const allowed = (s) => s !== 401 && s !== 403 && s < 500;

export default function () {
  const customer = login('sec-customer');
  const admin = adminLogin();
  // Forged token: same header and claims, signature changed.
  const forged = admin.slice(0, -4) + (admin.endsWith('AAAA') ? 'BBBB' : 'AAAA');

  for (const [method, path, who] of ROUTES) {
    const name = `${method} ${path}`;
    const anon = call(method, path).status;
    const cust = call(method, path, customer).status;
    const adm = call(method, path, admin).status;
    const fake = call(method, path, forged).status;

    if (who === 'public') {
      check(anon, { [`${name}: public, anonymous allowed`]: allowed });
      continue;
    }
    check(anon, { [`${name}: anonymous -> 401`]: (s) => s === 401 });
    check(fake, { [`${name}: forged JWT -> 401`]: (s) => s === 401 });
    check(adm, { [`${name}: admin allowed (got ${adm})`]: allowed });
    if (who === 'signedIn') {
      check(cust, { [`${name}: customer allowed (got ${cust})`]: allowed });
    } else {
      check(cust, { [`${name}: customer -> 403`]: (s) => s === 403 });
    }
  }

  // Defence in depth: the gateway lets any signed-in user through to
  // POST /api/reservations, and the Reservation service itself accepts only
  // customers.
  const res = call('POST', '/api/reservations', admin);
  check(res, { 'POST /api/reservations: service rejects admin role -> 403': (r) => r.status === 403 });
}
