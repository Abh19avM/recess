import ws from 'k6/ws';
import { check, sleep } from 'k6';
import { Trend, Rate, Counter } from 'k6/metrics';

// Custom WebSocket metrics
const wsConnectionDuration = new Trend('recess_ws_session_duration', true);
const wsMessagesReceived = new Counter('recess_ws_messages_received_total');
const wsMessagesSent = new Counter('recess_ws_messages_sent_total');
const wsErrors = new Rate('recess_ws_errors');

export const options = {
  stages: [
    { duration: '5s', target: 20 },   // Connect 20 players
    { duration: '15s', target: 50 },  // 50 concurrent duels & spectators
    { duration: '5s', target: 0 },    // Disconnect
  ],
  thresholds: {
    'recess_ws_errors': ['rate<0.02'], // < 2% error rate
  },
};

const WS_URL = __ENV.WS_URL || 'ws://localhost:8080/ws';

export default function () {
  const vuId = __VU;
  const iterId = __ITER;
  const isSpectator = vuId % 5 === 0; // 20% spectators, 80% active players
  const role = isSpectator ? 'spectator' : 'player';
  const roomId = `room_load_desk_${vuId % 10}`; // 10 parallel classrooms
  const userId = `student_${vuId}_${iterId}`;

  const url = `${WS_URL}?token=guest_tok_${userId}&room_id=${roomId}&role=${role}`;

  const res = ws.connect(url, {}, function (socket) {
    const startTime = new Date().getTime();

    socket.on('open', function () {
      // 1. Join room
      const joinMsg = JSON.stringify({
        type: 'room.join',
        room_id: roomId,
        payload: {
          user_id: userId,
          role: role,
        },
      });
      socket.send(joinMsg);
      wsMessagesSent.add(1);

      // 2. If player, set ready and submit moves periodically
      if (!isSpectator) {
        socket.setTimeout(function () {
          socket.send(JSON.stringify({
            type: 'player.ready',
            room_id: roomId,
            payload: { is_ready: true },
          }));
          wsMessagesSent.add(1);
        }, 200);

        // Periodically submit simulated game moves (e.g. Hand Cricket numbers 1-6 or XO cell 0-8)
        let moveCount = 0;
        socket.setInterval(function () {
          if (moveCount < 6) {
            moveCount++;
            socket.send(JSON.stringify({
              type: 'game.move',
              room_id: roomId,
              payload: {
                action: 'choose_number',
                value: (moveCount % 6) + 1,
              },
            }));
            wsMessagesSent.add(1);
          }
        }, 1000);
      }

      // 3. Heartbeat / ping
      socket.setInterval(function () {
        socket.send(JSON.stringify({
          type: 'ping',
          room_id: roomId,
          timestamp: new Date().getTime(),
        }));
        wsMessagesSent.add(1);
      }, 3000);

      // 4. Session duration
      socket.setTimeout(function () {
        socket.close();
      }, 5000);
    });

    socket.on('message', function (data) {
      wsMessagesReceived.add(1);
      try {
        const msg = JSON.parse(data);
        check(msg, {
          'has message type': (m) => m.type !== undefined,
        });
      } catch (e) {
        wsErrors.add(1);
      }
    });

    socket.on('close', function () {
      const duration = new Date().getTime() - startTime;
      wsConnectionDuration.add(duration);
    });

    socket.on('error', function (e) {
      wsErrors.add(1);
    });
  });

  check(res, {
    'ws connection established (101 Switching Protocols)': (r) => r && r.status === 101,
  });

  sleep(0.5);
}
