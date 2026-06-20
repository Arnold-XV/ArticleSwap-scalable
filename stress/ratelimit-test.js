// Rate limit test: menguji fitur rate limiting berbasis Redis.
// Mengirim request melebihi batas API_RATE_LIMIT_PER_MINUTE untuk memastikan
// server menolak dengan HTTP 429 Too Many Requests.
//
// Cara menjalankan:
//   Pastikan .env berisi API_RATE_LIMIT_PER_MINUTE=60 (default).
//   k6 run stress/ratelimit-test.js

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter } from 'k6/metrics';

const BASE_URL = __ENV.API_BASE_URL || 'http://localhost:8080';

// Custom metrics untuk mencatat jumlah 429 yang diterima.
const rateLimited = new Counter('rate_limited_responses');
const allowed = new Counter('allowed_responses');

export const options = {
  // 1 VU saja agar semua request datang dari 1 IP.
  // Dengan rate limit default 60/menit, 80 request dalam 30 detik pasti kena limit.
  vus: 1,
  duration: '30s',
  thresholds: {
    // Kita MENGHARAPKAN ada yang kena 429, jadi threshold lebih longgar.
    http_req_failed: ['rate<0.60'],
  },
};

export default function () {
  // Kirim request cepat-cepat tanpa sleep agar tembus rate limit.
  const res = http.get(`${BASE_URL}/health`);

  if (res.status === 429) {
    rateLimited.add(1);
    check(res, {
      'rate limited has Retry-After header': (r) =>
        r.headers['Retry-After'] !== undefined,
      'rate limited body contains error': (r) =>
        r.body.includes('rate limit'),
    });
  } else {
    allowed.add(1);
    check(res, {
      'request allowed': (r) => r.status === 200,
    });
  }

  // Sedikit jeda supaya k6 tidak throttle di level transport.
  sleep(0.1);
}

export function handleSummary(data) {
  const rateLimitedCount =
    data.metrics.rate_limited_responses
      ? data.metrics.rate_limited_responses.values.count
      : 0;
  const allowedCount =
    data.metrics.allowed_responses
      ? data.metrics.allowed_responses.values.count
      : 0;
  const total = rateLimitedCount + allowedCount;

  console.log(`\n=== RATE LIMIT TEST SUMMARY ===`);
  console.log(`Total requests   : ${total}`);
  console.log(`Allowed (200)    : ${allowedCount}`);
  console.log(`Blocked (429)    : ${rateLimitedCount}`);
  console.log(
    `Block rate       : ${total > 0 ? ((rateLimitedCount / total) * 100).toFixed(1) : 0}%`
  );

  if (rateLimitedCount > 0) {
    console.log(`✅ Rate limiting AKTIF — server menolak request berlebihan.`);
  } else {
    console.log(
      `⚠️  Tidak ada request yang di-rate-limit. Pastikan API_RATE_LIMIT_PER_MINUTE < 300 di .env.`
    );
  }
  console.log(`===============================\n`);

  return {};
}
