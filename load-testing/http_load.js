import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Rate, Counter } from 'k6/metrics';

// Custom metrics
const httpDuration = new Trend('recess_http_req_duration', true);
const errorRate = new Rate('recess_http_errors');
const successCounter = new Counter('recess_http_success_total');

export const options = {
  stages: [
    { duration: '5s', target: 25 },   // Ramp up to 25 VUs
    { duration: '15s', target: 50 },  // Sustained 50 concurrent VUs
    { duration: '5s', target: 0 },    // Ramp down
  ],
  thresholds: {
    'http_req_duration': ['p(50)<50', 'p(95)<150', 'p(99)<300'],
    'recess_http_errors': ['rate<0.01'], // < 1% error rate
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Setup: Pre-generate a master student token for authenticated read tests
export function setup() {
  const guestRes = http.post(
    `${BASE_URL}/api/v1/auth/guest`,
    JSON.stringify({ nickname: 'BenchmarkStudent' }),
    { headers: { 'Content-Type': 'application/json' } }
  );

  let token = '';
  let userId = 'usr_demo_headmaster';
  if (guestRes.status === 200) {
    const body = JSON.parse(guestRes.body);
    if (body.data && body.data.tokens) {
      token = body.data.tokens.access_token;
      userId = body.data.user.id;
    }
  }
  return { token, userId };
}

export default function (data) {
  const authHeaders = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${data.token}`,
  };

  // 1. Health check
  const healthRes = http.get(`${BASE_URL}/health`);
  const healthOk = check(healthRes, {
    'health status is 200': (r) => r.status === 200,
  });
  httpDuration.add(healthRes.timings.duration);
  errorRate.add(!healthOk);

  // 2. Fetch Syllabus (Games List)
  const gamesRes = http.get(`${BASE_URL}/api/v1/games`);
  const gamesOk = check(gamesRes, {
    'games list status is 200': (r) => r.status === 200,
    'games list has 6 games': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.data && body.data.games && body.data.games.length >= 6;
      } catch {
        return false;
      }
    },
  });
  httpDuration.add(gamesRes.timings.duration);
  errorRate.add(!gamesOk);

  // 3. Query Open Classrooms / Rooms
  const roomsRes = http.get(`${BASE_URL}/api/v1/rooms`, { headers: authHeaders });
  const roomsOk = check(roomsRes, {
    'rooms status is 200': (r) => r.status === 200,
  });
  httpDuration.add(roomsRes.timings.duration);
  errorRate.add(!roomsOk);

  // 4. Query Student Profile
  const profileRes = http.get(`${BASE_URL}/api/v1/users/usr_demo_headmaster`, { headers: authHeaders });
  const profileOk = check(profileRes, {
    'profile status is 200': (r) => r.status === 200,
  });
  httpDuration.add(profileRes.timings.duration);
  errorRate.add(!profileOk);

  // 5. Query Global Leaderboard
  const lbGlobalRes = http.get(`${BASE_URL}/api/v1/leaderboards`, { headers: authHeaders });
  const lbGlobalOk = check(lbGlobalRes, {
    'global leaderboard status is 200': (r) => r.status === 200,
  });
  httpDuration.add(lbGlobalRes.timings.duration);
  errorRate.add(!lbGlobalOk);

  // 6. Query Game-specific Leaderboard (Hand Cricket)
  const lbGameRes = http.get(`${BASE_URL}/api/v1/leaderboards/hand_cricket`, { headers: authHeaders });
  const lbGameOk = check(lbGameRes, {
    'game leaderboard status is 200': (r) => r.status === 200,
  });
  httpDuration.add(lbGameRes.timings.duration);
  errorRate.add(!lbGameOk);

  // 7. Prometheus Metrics Scraping
  const metricsRes = http.get(`${BASE_URL}/metrics`);
  const metricsOk = check(metricsRes, {
    'metrics status is 200': (r) => r.status === 200,
  });
  httpDuration.add(metricsRes.timings.duration);
  errorRate.add(!metricsOk);

  if (healthOk && gamesOk && roomsOk && profileOk && lbGlobalOk && lbGameOk && metricsOk) {
    successCounter.add(1);
  }

  sleep(0.1);
}
