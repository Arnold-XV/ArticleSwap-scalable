// Stress test bertahap: beban naik dari 10 sampai 100 virtual users.
// Menguji batas kapasitas sistem ArticleSwap.

import http from 'k6/http';
import { check, sleep } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

const BASE_URL = __ENV.API_BASE_URL || 'http://localhost:8080';

export const options = {
  stages: [
    { duration: '30s', target: 10 },   // Ramp up ke 10 VUs.
    { duration: '1m', target: 10 },     // Tahan 10 VUs.
    { duration: '30s', target: 50 },    // Ramp up ke 50 VUs.
    { duration: '1m', target: 50 },     // Tahan 50 VUs.
    { duration: '30s', target: 100 },   // Ramp up ke 100 VUs.
    { duration: '1m', target: 100 },    // Tahan 100 VUs.
    { duration: '30s', target: 0 },     // Ramp down.
  ],
  thresholds: {
    http_req_duration: ['p(95)<5000'],
    http_req_failed: ['rate<0.10'],
  },
};

export function setup() {
  const res = http.get(`${BASE_URL}/users`);
  if (res.status === 200) {
    return { users: JSON.parse(res.body) };
  }
  return { users: [] };
}

export default function (data) {
  const userList = data.users;
  if (userList.length < 2) return;

  const sender = userList[Math.floor(Math.random() * userList.length)];
  let receiver = sender;
  while (receiver.id === sender.id) {
    receiver = userList[Math.floor(Math.random() * userList.length)];
  }

  // Submit article.
  const payload = JSON.stringify({
    sender_id: sender.id,
    receiver_id: receiver.id,
    title: `Stress Test ${__VU}-${__ITER}`,
    content: `Artikel stress test dari virtual user ${__VU} iterasi ${__ITER}. Konten ini dibuat untuk menguji performa sistem pada beban tinggi. Sistem harus mampu menangani banyak request secara bersamaan tanpa mengalami penurunan kinerja signifikan.`,
    idempotency_key: uuidv4(),
  });

  const submitRes = http.post(`${BASE_URL}/articles`, payload, {
    headers: { 'Content-Type': 'application/json' },
  });

  check(submitRes, {
    'submit success': (r) => r.status === 202 || r.status === 200,
  });

  // Get inbox.
  http.get(`${BASE_URL}/users/${receiver.id}/inbox`);

  sleep(0.5);
}
