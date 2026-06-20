// Pipeline end-to-end test: submit artikel lalu polling sampai status 'completed'.
// Membuktikan bahwa pipeline asinkron (stemmer → wordcloud → aggregator) berjalan penuh.
//
// Test ini lebih lambat karena perlu menunggu worker memproses.
// Pastikan semua worker sudah berjalan sebelum menjalankan test ini.

import http from 'k6/http';
import { check, sleep } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';
import { Counter, Trend } from 'k6/metrics';

const BASE_URL = __ENV.API_BASE_URL || 'http://localhost:8080';

// Custom metrics.
const pipelineCompleted = new Counter('pipeline_completed');
const pipelineFailed = new Counter('pipeline_failed');
const pipelineDuration = new Trend('pipeline_duration_ms');

export const options = {
  vus: 5,
  duration: '2m',
  thresholds: {
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

  // 1. Submit artikel.
  const submitTime = Date.now();
  const payload = JSON.stringify({
    sender_id: sender.id,
    receiver_id: receiver.id,
    title: `Pipeline E2E ${__VU}-${__ITER}`,
    content: `Artikel end-to-end test untuk memverifikasi pipeline lengkap. Virtual user ${__VU} iterasi ${__ITER}. Konten ini akan melalui stemming, word cloud, dan aggregator sebelum status menjadi completed.`,
    idempotency_key: uuidv4(),
  });

  const submitRes = http.post(`${BASE_URL}/articles`, payload, {
    headers: { 'Content-Type': 'application/json' },
  });

  check(submitRes, {
    'submit accepted': (r) => r.status === 202 || r.status === 200,
  });

  if (submitRes.status !== 202 && submitRes.status !== 200) {
    pipelineFailed.add(1);
    return;
  }

  const article = JSON.parse(submitRes.body);
  const articleId = article.id;

  // 2. Polling: cek status artikel sampai completed atau timeout.
  const maxWait = 15; // maksimal 15 detik menunggu pipeline selesai.
  let completed = false;

  for (let i = 0; i < maxWait; i++) {
    sleep(1);

    const detailRes = http.get(`${BASE_URL}/articles/${articleId}`);
    if (detailRes.status !== 200) continue;

    const detail = JSON.parse(detailRes.body);

    // Cek apakah processing sudah selesai.
    if (detail.status === 'completed') {
      completed = true;

      const duration = Date.now() - submitTime;
      pipelineDuration.add(duration);
      pipelineCompleted.add(1);

      check(detailRes, {
        'article completed': () => true,
        'has stemmed content': () =>
          detail.processing && detail.processing.stemmed_content !== null,
        'has word frequencies': () =>
          detail.processing && detail.processing.word_frequencies !== null,
        'stemming status completed': () =>
          detail.processing && detail.processing.stemming_status === 'completed',
        'wordcloud status completed': () =>
          detail.processing && detail.processing.wordcloud_status === 'completed',
      });

      break;
    }

    // Jika sudah 'failed', langsung keluar.
    if (detail.status === 'failed') {
      pipelineFailed.add(1);
      break;
    }
  }

  if (!completed) {
    pipelineFailed.add(1);
  }

  // 3. Verifikasi artikel muncul di inbox penerima.
  const inboxRes = http.get(`${BASE_URL}/users/${receiver.id}/inbox`);
  check(inboxRes, {
    'inbox accessible': (r) => r.status === 200,
    'article in inbox': (r) => {
      const inbox = JSON.parse(r.body);
      return inbox.some((a) => a.id === articleId);
    },
  });
}

export function handleSummary(data) {
  const completed =
    data.metrics.pipeline_completed
      ? data.metrics.pipeline_completed.values.count
      : 0;
  const failed =
    data.metrics.pipeline_failed
      ? data.metrics.pipeline_failed.values.count
      : 0;
  const total = completed + failed;

  const avgDuration =
    data.metrics.pipeline_duration_ms
      ? data.metrics.pipeline_duration_ms.values.avg.toFixed(0)
      : 'N/A';
  const p95Duration =
    data.metrics.pipeline_duration_ms
      ? data.metrics.pipeline_duration_ms.values['p(95)'].toFixed(0)
      : 'N/A';

  console.log(`\n=== PIPELINE END-TO-END SUMMARY ===`);
  console.log(`Total articles   : ${total}`);
  console.log(`Completed        : ${completed}`);
  console.log(`Failed/Timeout   : ${failed}`);
  console.log(
    `Success rate     : ${total > 0 ? ((completed / total) * 100).toFixed(1) : 0}%`
  );
  console.log(`Avg pipeline time: ${avgDuration} ms`);
  console.log(`P95 pipeline time: ${p95Duration} ms`);
  console.log(`====================================\n`);

  return {};
}
