import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Rate, Counter } from 'k6/metrics';

const mmQueueLatency = new Trend('recess_mm_queue_latency_ms', true);
const mmMatchLatency = new Trend('recess_mm_match_latency_ms', true);
const mmSuccessMatches = new Counter('recess_mm_matches_formed_total');
const mmErrors = new Rate('recess_mm_errors');

export const options = {
  stages: [
    { duration: '5s', target: 20 },   // 20 concurrent queue entries
    { duration: '15s', target: 40 },  // 40 concurrent players matching
    { duration: '5s', target: 0 },    // Cool down
  ],
  thresholds: {
    'recess_mm_errors': ['rate<0.02'],
    'http_req_duration': ['p(50)<50', 'p(95)<150'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const GAMES = ['hand_cricket', 'dots_boxes', 'tic_tac_toe', 'connect_4', 'paper_football', 'npat'];

// Setup: Create an authenticated student session
export function setup() {
  const res = http.post(
    `${BASE_URL}/api/v1/auth/guest`,
    JSON.stringify({ nickname: 'BenchmarkPlayer' }),
    { headers: { 'Content-Type': 'application/json' } }
  );

  let token = '';
  if (res.status === 200 || res.status === 201) {
    const data = JSON.parse(res.body);
    token = data.data.tokens.access_token;
  }
  return { token };
}

export default function (data) {
  const vuId = __VU;
  const gameType = GAMES[vuId % GAMES.length];
  const mode = vuId % 2 === 0 ? 'casual' : 'ranked';
  const token = data.token;

  const authHeaders = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
  };

  // 1. Enter Matchmaking Queue
  const queueStart = new Date().getTime();
  const queueRes = http.post(
    `${BASE_URL}/api/v1/matchmaking/join`,
    JSON.stringify({ game_type: gameType, mode: mode }),
    { headers: authHeaders }
  );

  const queueDuration = new Date().getTime() - queueStart;
  mmQueueLatency.add(queueDuration);

  const isQueued = check(queueRes, {
    'join queue status is 200 or 202': (r) => r.status === 200 || r.status === 202,
    'ticket_id returned': (r) => {
      try {
        const b = JSON.parse(r.body);
        return b.data && b.data.ticket_id !== undefined;
      } catch {
        return false;
      }
    },
  });

  if (!isQueued) {
    mmErrors.add(true);
    sleep(0.2);
    return;
  }
  mmErrors.add(false);

  const resBody = JSON.parse(queueRes.body);
  const ticketId = resBody.data.ticket_id;
  let matched = resBody.data.status === 'matched';

  if (matched) {
    mmSuccessMatches.add(1);
    mmMatchLatency.add(queueDuration);
  } else {
    // 2. Poll for Match Status
    for (let attempt = 0; attempt < 3; attempt++) {
      sleep(0.2);
      const statusRes = http.get(
        `${BASE_URL}/api/v1/matchmaking/status?ticket_id=${ticketId}`,
        { headers: authHeaders }
      );

      if (statusRes.status === 200) {
        try {
          const statusData = JSON.parse(statusRes.body);
          if (statusData.data && statusData.data.status === 'matched') {
            matched = true;
            const totalMatchTime = new Date().getTime() - queueStart;
            mmMatchLatency.add(totalMatchTime);
            mmSuccessMatches.add(1);
            break;
          }
        } catch {}
      }
    }

    // 3. Leave Queue if still waiting
    if (!matched) {
      http.post(
        `${BASE_URL}/api/v1/matchmaking/leave`,
        JSON.stringify({ game_type: gameType }),
        { headers: authHeaders }
      );
    }
  }

  // 4. Check Queue Size Endpoint
  const sizeRes = http.get(`${BASE_URL}/api/v1/matchmaking/size?game_type=${gameType}`, {
    headers: authHeaders,
  });
  check(sizeRes, {
    'size query status is 200': (r) => r.status === 200,
  });

  sleep(0.2);
}
