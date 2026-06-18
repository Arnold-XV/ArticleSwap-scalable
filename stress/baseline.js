// Stress test baseline: 10 virtual users selama 1 menit.
// Menguji endpoint utama ArticleSwap dengan beban ringan.

import http from 'k6/http';
import { check, sleep } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

const BASE_URL = __ENV.API_BASE_URL || 'http://localhost:8080';

export const options = {
  vus: 10,
  duration: '1m',
  thresholds: {
    http_req_duration: ['p(95)<2000', 'p(99)<5000'],
    http_req_failed: ['rate<0.05'],
  },
};

// Ambil daftar user saat setup.
let users = [];

export function setup() {
  const res = http.get(`${BASE_URL}/users`);
  if (res.status === 200) {
    return { users: JSON.parse(res.body) };
  }
  return { users: [] };
}

export default function (data) {
  const userList = data.users;
  if (userList.length < 2) {
    console.error('Butuh minimal 2 user untuk test');
    return;
  }

  // 1. Health check.
  const healthRes = http.get(`${BASE_URL}/health`);
  check(healthRes, {
    'health status 200': (r) => r.status === 200,
    'health status ok': (r) => JSON.parse(r.body).status === 'ok',
  });

  // 2. Get users.
  const usersRes = http.get(`${BASE_URL}/users`);
  check(usersRes, {
    'users status 200': (r) => r.status === 200,
    'users returns array': (r) => JSON.parse(r.body).length > 0,
  });

  // 3. Submit article.
  const sender = userList[Math.floor(Math.random() * userList.length)];
  let receiver = sender;
  while (receiver.id === sender.id) {
    receiver = userList[Math.floor(Math.random() * userList.length)];
  }

  const payload = JSON.stringify({
    sender_id: sender.id,
    receiver_id: receiver.id,
    title: `Test Article ${Date.now()}`,
    content: `Ini adalah artikel pengujian baseline yang dikirim pada waktu ${new Date().toISOString()}. Artikel ini berisi konten sederhana untuk menguji pipeline stemming dan word cloud.`,
    idempotency_key: uuidv4(),
  });

  const submitRes = http.post(`${BASE_URL}/articles`, payload, {
    headers: { 'Content-Type': 'application/json' },
  });

  check(submitRes, {
    'submit status 202': (r) => r.status === 202,
    'submit returns id': (r) => JSON.parse(r.body).id !== undefined,
  });

  if (submitRes.status === 202) {
    const article = JSON.parse(submitRes.body);

    // 4. Get article detail.
    const detailRes = http.get(`${BASE_URL}/articles/${article.id}`);
    check(detailRes, {
      'detail status 200': (r) => r.status === 200,
    });
  }

  // 5. Check inbox.
  const inboxRes = http.get(`${BASE_URL}/users/${receiver.id}/inbox`);
  check(inboxRes, {
    'inbox status 200': (r) => r.status === 200,
  });

  sleep(1);
}
