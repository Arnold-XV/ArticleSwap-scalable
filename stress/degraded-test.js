// Degraded mode test: menguji perilaku sistem saat wordcloud worker dimatikan.
// Artikel seharusnya tetap bisa dikirim dan dibaca. Status akhir menjadi 'degraded'.
//
// Cara menjalankan:
//   1. docker compose up --build
//   2. docker compose stop wordcloud-worker
//   3. k6 run stress/degraded-test.js

import http from 'k6/http';
import { check, sleep } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

const BASE_URL = __ENV.API_BASE_URL || 'http://localhost:8080';

export const options = {
  vus: 5,
  duration: '30s',
  thresholds: {
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
    title: `Degraded Test ${__VU}-${__ITER}`,
    content: `Artikel ini dikirim saat wordcloud worker dimatikan. Stemming seharusnya tetap berhasil. Status akhir artikel menjadi degraded karena wordcloud gagal.`,
    idempotency_key: uuidv4(),
  });

  const submitRes = http.post(`${BASE_URL}/articles`, payload, {
    headers: { 'Content-Type': 'application/json' },
  });

  check(submitRes, {
    'submit success': (r) => r.status === 202 || r.status === 200,
  });

  // Artikel harus tetap muncul di inbox meskipun wordcloud gagal.
  const inboxRes = http.get(`${BASE_URL}/users/${receiver.id}/inbox`);
  check(inboxRes, {
    'inbox accessible': (r) => r.status === 200,
  });

  // Cek health endpoint tetap berjalan.
  const healthRes = http.get(`${BASE_URL}/health`);
  check(healthRes, {
    'health reachable': (r) => r.status === 200 || r.status === 503,
  });

  sleep(1);
}
