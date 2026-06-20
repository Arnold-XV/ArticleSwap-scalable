// Connection pool saturation test: membanjiri database query untuk menguji
// batas connection pool PostgreSQL (DB_POOL_MAX_CONNS).
// Menguji apakah pooling tetap stabil di bawah beban tinggi.

import http from 'k6/http';
import { check, sleep } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';
import { Counter } from 'k6/metrics';

const BASE_URL = __ENV.API_BASE_URL || 'http://localhost:8080';

const dbErrors = new Counter('db_errors');

export const options = {
  stages: [
    { duration: '15s', target: 20 },    // Ramp up.
    { duration: '30s', target: 20 },    // Tahan 20 VUs — mendekati batas pool.
    { duration: '15s', target: 50 },    // Naikkan melebihi pool size (default MAX_CONNS=20).
    { duration: '30s', target: 50 },    // Tahan 50 VUs.
    { duration: '15s', target: 80 },    // Naikkan lagi.
    { duration: '30s', target: 80 },    // Tahan 80 VUs — jauh di atas pool size.
    { duration: '15s', target: 0 },     // Ramp down.
  ],
  thresholds: {
    http_req_duration: ['p(95)<10000'],
    http_req_failed: ['rate<0.20'],
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

  // Campuran operasi yang semuanya mengakses database secara bersamaan.

  // 1. Submit artikel baru (INSERT ke articles + processing_results + idempotency_keys).
  const payload = JSON.stringify({
    sender_id: sender.id,
    receiver_id: receiver.id,
    title: `Pool Test ${__VU}-${__ITER}`,
    content: `Connection pool stress test dari VU ${__VU}. Test ini mengirim banyak query bersamaan untuk menguji stabilitas pgxpool saat koneksi mendekati atau melampaui MAX_CONNS.`,
    idempotency_key: uuidv4(),
  });

  const submitRes = http.post(`${BASE_URL}/articles`, payload, {
    headers: { 'Content-Type': 'application/json' },
  });

  check(submitRes, {
    'submit ok': (r) => r.status === 202 || r.status === 200,
  });

  if (submitRes.status >= 500) {
    dbErrors.add(1);
  }

  // 2. Query inbox (SELECT with JOIN).
  const inboxRes = http.get(`${BASE_URL}/users/${receiver.id}/inbox`);
  check(inboxRes, {
    'inbox ok': (r) => r.status === 200,
  });
  if (inboxRes.status >= 500) {
    dbErrors.add(1);
  }

  // 3. Query metrics (multiple aggregation queries).
  const metricsRes = http.get(`${BASE_URL}/metrics/summary`);
  check(metricsRes, {
    'metrics ok': (r) => r.status === 200,
  });
  if (metricsRes.status >= 500) {
    dbErrors.add(1);
  }

  // 4. Health check (ping DB).
  const healthRes = http.get(`${BASE_URL}/health`);
  check(healthRes, {
    'health ok': (r) => r.status === 200 || r.status === 503,
  });

  // Tanpa sleep agar membanjiri pool secepat mungkin.
}

export function handleSummary(data) {
  const dbErrorCount =
    data.metrics.db_errors ? data.metrics.db_errors.values.count : 0;
  const totalReqs =
    data.metrics.http_reqs ? data.metrics.http_reqs.values.count : 0;

  console.log(`\n=== CONNECTION POOL SATURATION SUMMARY ===`);
  console.log(`Total HTTP requests : ${totalReqs}`);
  console.log(`DB-related errors   : ${dbErrorCount}`);
  console.log(
    `DB error rate       : ${totalReqs > 0 ? ((dbErrorCount / totalReqs) * 100).toFixed(2) : 0}%`
  );

  if (dbErrorCount === 0) {
    console.log(`✅ Connection pool stabil — tidak ada error database.`);
  } else {
    console.log(
      `⚠️  Ada ${dbErrorCount} error database. Pertimbangkan naikkan DB_POOL_MAX_CONNS.`
    );
  }
  console.log(`==========================================\n`);

  return {};
}
