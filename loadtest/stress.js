// Stress test (Load Test Plan, test type 2): steps of 50, 100, 200 and 400 VUs,
// 2 min each, then a light recovery phase. Each step is its own scenario, so
// k6 tags every request with it and the summary shows p95 and error rate per
// step: the first step marked ✗ is the breaking point. A ✗ here is a finding,
// not a failed run.
import { setupData, userJourney } from './lib.js';

const steps = [50, 100, 200, 400];
const scenarios = {};
steps.forEach((vus, i) => {
  scenarios[`step_${vus}`] = { executor: 'constant-vus', vus, duration: '2m', startTime: `${i * 2}m`, gracefulStop: '10s' };
});
// Recovery: does the system answer normally again once load drops?
scenarios.recovery = { executor: 'constant-vus', vus: 10, duration: '1m', startTime: `${steps.length * 2}m` };

const thresholds = {};
for (const s of Object.keys(scenarios)) {
  thresholds[`http_req_duration{scenario:${s},kind:reservation}`] = ['p(95)<1000'];
  thresholds[`http_req_duration{scenario:${s},kind:read}`] = ['p(95)<500'];
  thresholds[`errors{scenario:${s}}`] = ['rate<0.01'];
  thresholds[`rate_limited{scenario:${s}}`] = ['rate<0.01'];
}

export const options = { scenarios, thresholds, summaryTrendStats: ['avg', 'p(50)', 'p(95)', 'p(99)', 'max'] };
export const setup = setupData;
export default userJourney;
