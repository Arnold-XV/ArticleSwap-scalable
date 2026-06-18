// Cache test: submit konten identik berulang kali untuk membuktikan Redis cache.
// Artikel dengan content_hash yang sama seharusnya memanfaatkan cache.

import http from 'k6/http';
import { check, sleep } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

const BASE_URL = __ENV.API_BASE_URL || 'http://localhost:8080';

// Konten tetap yang sama untuk setiap request agar content_hash identik.
const SAME_CONTENT = `Ini adalah artikel dengan konten yang sama persis untuk menguji fitur Redis cache pada ArticleSwap. Konten ini akan di-hash menggunakan SHA-256 dan hasilnya digunakan sebagai cache key. Saat worker memproses artikel kedua dengan konten yang sama, worker seharusnya mendapatkan cache hit dari Redis tanpa perlu mengulangi proses stemming dan word cloud.`;

export const options = {
  vus: 10,
  duration: '1m',
  thresholds: {
    http_req_duration: ['p(95)<2000'],
    http_req_failed: ['rate<0.05'],
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

  // Submit article dengan konten yang sama (idempotency key berbeda).
  const payload = JSON.stringify({
    sender_id: sender.id,
    receiver_id: receiver.id,
    title: `Cache Test ${__VU}-${__ITER}`,
    content: SAME_CONTENT,
    idempotency_key: uuidv4(), // Key berbeda → artikel baru, tapi konten sama → cache hit.
  });

  const submitRes = http.post(`${BASE_URL}/articles`, payload, {
    headers: { 'Content-Type': 'application/json' },
  });

  check(submitRes, {
    'submit success': (r) => r.status === 202 || r.status === 200,
  });

  if (submitRes.status === 202) {
    const article = JSON.parse(submitRes.body);

    // Tunggu sebentar lalu cek hasilnya.
    sleep(2);

    const detailRes = http.get(`${BASE_URL}/articles/${article.id}`);
    check(detailRes, {
      'detail status 200': (r) => r.status === 200,
    });
  }

  sleep(1);
}
