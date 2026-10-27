// Load test (Load Test Plan, test type 1): expected peak use, 50 concurrent users.
// Pass/fail is the Performance NFRs: k6 exits non-zero if a threshold is missed.
import { setupData, userJourney } from './lib.js';

export const options = {
  stages: [
    { duration: '1m', target: 50 },
    { duration: '5m', target: 50 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    'http_req_duration{kind:reservation}': ['p(95)<1000'], // create, accept, cancel < 1 s
    'http_req_duration{kind:read}': ['p(95)<500'], // occupancy, library, review reads < 500 ms
    errors: ['rate<0.01'], // 5xx + timeouts < 1%
    rate_limited: ['rate<0.01'], // 429s would mean the test measured the limiter, not the system
  },
  summaryTrendStats: ['avg', 'p(50)', 'p(95)', 'p(99)', 'max'],
};

export const setup = setupData;
export default userJourney;
