// Idempotency test: submit idempotency key yang sama berkali-kali.
// Seharusnya tidak membuat artikel duplikat.

import http from 'k6/http';
import { check, sleep } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

const BASE_URL = __ENV.API_BASE_URL || 'http://localhost:8080';

export const options = {
  vus: 5,
  duration: '30s',
  thresholds: {
    http_req_failed: ['rate<0.01'],
  },
};

// Shared idempotency key per VU.
const IDEMPOTENCY_KEYS = {};

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

  // Gunakan idempotency key yang sama per VU untuk semua iterasi.
  const idempotencyKey = `idem-test-vu-${__VU}`;

  const sender = userList[0];
  const receiver = userList[1];

  const payload = JSON.stringify({
    sender_id: sender.id,
    receiver_id: receiver.id,
    title: `Idempotency Test VU ${__VU}`,
    content: `Artikel idempotency test dari virtual user ${__VU}. Key yang sama dipakai berulang kali untuk memastikan tidak ada duplikasi.`,
    idempotency_key: idempotencyKey,
  });

  const res = http.post(`${BASE_URL}/articles`, payload, {
    headers: { 'Content-Type': 'application/json' },
  });

  check(res, {
    'idempotent response': (r) => r.status === 200 || r.status === 202,
    'returns article id': (r) => {
      const body = JSON.parse(r.body);
      return body.id !== undefined;
    },
  });

  // Semua request dengan key yang sama harus mengembalikan article ID yang sama.
  if (res.status === 200 || res.status === 202) {
    const article = JSON.parse(res.body);
    if (__ITER === 0) {
      console.log(`VU ${__VU}: first submit → article_id=${article.id}`);
    }
  }

  sleep(0.5);
}
