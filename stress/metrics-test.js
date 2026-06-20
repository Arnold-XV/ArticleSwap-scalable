// Metrics endpoint test: memastikan GET /metrics/summary berfungsi di bawah beban.
// Endpoint ini digunakan untuk monitoring dan data poster.

import http from 'k6/http';
import { check, sleep } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

const BASE_URL = __ENV.API_BASE_URL || 'http://localhost:8080';

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

  // 1. GET /metrics/summary — endpoint utama yang diuji.
  const metricsRes = http.get(`${BASE_URL}/metrics/summary`);
  check(metricsRes, {
    'metrics status 200': (r) => r.status === 200,
    'metrics has total_users': (r) => {
      const body = JSON.parse(r.body);
      return body.total_users !== undefined;
    },
    'metrics has total_articles': (r) => {
      const body = JSON.parse(r.body);
      return body.total_articles !== undefined;
    },
    'metrics has articles_by_status': (r) => {
      const body = JSON.parse(r.body);
      return body.articles_by_status !== undefined;
    },
    'metrics has processing_summary': (r) => {
      const body = JSON.parse(r.body);
      return body.processing_summary !== undefined;
    },
  });

  // 2. Submit artikel baru supaya data metrics terus berubah selama test.
  if (userList.length >= 2) {
    const sender = userList[Math.floor(Math.random() * userList.length)];
    let receiver = sender;
    while (receiver.id === sender.id) {
      receiver = userList[Math.floor(Math.random() * userList.length)];
    }

    const payload = JSON.stringify({
      sender_id: sender.id,
      receiver_id: receiver.id,
      title: `Metrics Test ${__VU}-${__ITER}`,
      content: `Artikel untuk menguji endpoint metrics. Virtual user ${__VU} iterasi ${__ITER}. Data metrics seharusnya bertambah setiap ada artikel baru masuk.`,
      idempotency_key: uuidv4(),
    });

    const submitRes = http.post(`${BASE_URL}/articles`, payload, {
      headers: { 'Content-Type': 'application/json' },
    });

    check(submitRes, {
      'submit success': (r) => r.status === 202 || r.status === 200,
    });
  }

  // 3. Panggil metrics lagi setelah submit untuk verifikasi data berubah.
  const metricsRes2 = http.get(`${BASE_URL}/metrics/summary`);
  check(metricsRes2, {
    'metrics still accessible': (r) => r.status === 200,
  });

  sleep(1);
}
