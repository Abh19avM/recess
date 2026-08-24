import http from 'k6/http';
import ws from 'k6/ws';
import { check, sleep, group } from 'k6';
import { Trend, Rate, Counter } from 'k6/metrics';

// Global Trends & Metrics
const apiLatency = new Trend('recess_full_api_latency', true);
const errorRate = new Rate('recess_full_errors');
const duelsCompleted = new Counter('recess_full_duels_completed');

export const options = {
  scenarios: {
    // Stage 1: API and Browsing Traffic (50 VUs)
    rest_traffic: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '5s', target: 25 },
        { duration: '15s', target: 50 },
        { duration: '5s', target: 0 },
      ],
      exec: 'restTraffic',
    },
    // Stage 2: Concurrent WebSocket Duels (25 VUs)
    ws_duels: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '5s', target: 15 },
        { duration: '15s', target: 25 },
        { duration: '5s', target: 0 },
      ],
      exec: 'wsDuels',
    },
  },
  thresholds: {
    'http_req_duration': ['p(50)<50', 'p(95)<150', 'p(99)<300'],
    'recess_full_errors': ['rate<0.02'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const WS_URL = __ENV.WS_URL || 'ws://localhost:8080/ws';

export function setup() {
  const guestRes = http.post(
    `${BASE_URL}/api/v1/auth/guest`,
    JSON.stringify({ nickname: 'BenchmarkFullStudent' }),
    { headers: { 'Content-Type': 'application/json' } }
  );

  let token = '';
  if (guestRes.status === 200 || guestRes.status === 201) {
    const body = JSON.parse(guestRes.body);
    token = body.data.tokens.access_token;
  }
  return { token };
}

export function restTraffic(data) {
  const authHeaders = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${data.token}`,
  };

  group('REST Syllabus & Leaderboards', function () {
    const res = http.get(`${BASE_URL}/api/v1/games`);
    const ok1 = check(res, { 'games status 200': (r) => r.status === 200 });
    apiLatency.add(res.timings.duration);
    errorRate.add(!ok1);

    const lbRes = http.get(`${BASE_URL}/api/v1/leaderboards`, { headers: authHeaders });
    const ok2 = check(lbRes, { 'leaderboard status 200': (r) => r.status === 200 });
    apiLatency.add(lbRes.timings.duration);
    errorRate.add(!ok2);

    const roomsRes = http.get(`${BASE_URL}/api/v1/rooms`, { headers: authHeaders });
    const ok3 = check(roomsRes, { 'rooms status 200': (r) => r.status === 200 });
    apiLatency.add(roomsRes.timings.duration);
    errorRate.add(!ok3);
  });

  sleep(0.2);
}

export function wsDuels() {
  group('WebSocket Game Duel', function () {
    const roomId = `room_full_desk_${__VU % 10}`;
    const url = `${WS_URL}?token=guest_tok_${__VU}&room_id=${roomId}&role=player`;

    const res = ws.connect(url, {}, function (socket) {
      socket.on('open', function () {
        socket.send(JSON.stringify({
          type: 'room.join',
          room_id: roomId,
          payload: { user_id: `player_${__VU}` },
        }));

        socket.setTimeout(function () {
          socket.send(JSON.stringify({
            type: 'player.ready',
            room_id: roomId,
            payload: { is_ready: true },
          }));
        }, 100);

        socket.setTimeout(function () {
          socket.send(JSON.stringify({
            type: 'game.move',
            room_id: roomId,
            payload: { action: 'move', cell: __VU % 9 },
          }));
          duelsCompleted.add(1);
        }, 500);

        socket.setTimeout(function () {
          socket.close();
        }, 2000);
      });
    });

    check(res, { 'ws connected': (r) => r && r.status === 101 });
  });

  sleep(0.5);
}
