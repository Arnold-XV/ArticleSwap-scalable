# Dokumentasi API ArticleSwap

Base URL: `http://localhost:8080`

Semua response menggunakan `Content-Type: application/json`.

---

## GET /health

Mengecek konektivitas ke PostgreSQL, Redis, dan RabbitMQ.

**Response 200 OK**

```json
{
  "status": "ok",
  "postgres": "up",
  "redis": "up",
  "rabbitmq": "up",
  "time": "2026-06-18T09:00:00Z"
}
```

**Response 503 Service Unavailable** (jika salah satu service down)

```json
{
  "status": "degraded",
  "postgres": "up",
  "redis": "down",
  "rabbitmq": "up",
  "time": "2026-06-18T09:00:00Z"
}
```

---

## GET /users

Mengembalikan daftar semua user yang terdaftar.

**Response 200 OK**

```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Maulana Faris Al Ghifari",
    "email": "maulana.faris@example.local",
    "created_at": "2026-06-18T09:00:00Z"
  },
  {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "name": "Raditya Nathaniel Nugroho",
    "email": "raditya.nathaniel@example.local",
    "created_at": "2026-06-18T09:00:00Z"
  }
]
```

---

## POST /articles

Mengirim artikel baru dari satu user ke user lain. Endpoint ini:

1. Mengecek idempotency key untuk mencegah duplikasi.
2. Menghitung `content_hash` (SHA-256) dari isi artikel.
3. Menyimpan artikel ke PostgreSQL dengan status `queued`.
4. Mempublikasikan job ke RabbitMQ untuk diproses worker.

**Request Body**

```json
{
  "sender_id": "uuid-pengirim",
  "receiver_id": "uuid-penerima",
  "title": "Judul Artikel",
  "content": "Isi lengkap artikel yang akan dikirim.",
  "idempotency_key": "unique-key-dari-frontend"
}
```

| Field | Tipe | Wajib | Keterangan |
| --- | --- | --- | --- |
| `sender_id` | UUID string | Ya | ID user pengirim |
| `receiver_id` | UUID string | Ya | ID user penerima |
| `title` | string | Ya | Judul artikel |
| `content` | string | Ya | Isi artikel |
| `idempotency_key` | string | Ya | Key unik untuk mencegah submit duplikat |

**Response 202 Accepted** (artikel baru berhasil dibuat)

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440099",
  "sender_id": "uuid-pengirim",
  "receiver_id": "uuid-penerima",
  "title": "Judul Artikel",
  "content": "Isi lengkap artikel yang akan dikirim.",
  "content_hash": "a1b2c3d4e5f6...",
  "status": "queued",
  "created_at": "2026-06-18T09:00:00Z",
  "updated_at": "2026-06-18T09:00:00Z",
  "processing": {
    "article_id": "550e8400-e29b-41d4-a716-446655440099",
    "stemmed_content": null,
    "word_frequencies": null,
    "stemming_status": "queued",
    "wordcloud_status": "queued",
    "processing_started_at": null,
    "processing_finished_at": null
  }
}
```

**Response 200 OK** (idempotency key sudah pernah dipakai — mengembalikan artikel yang sama)

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440099",
  "sender_id": "uuid-pengirim",
  "receiver_id": "uuid-penerima",
  "title": "Judul Artikel",
  "content": "Isi lengkap artikel yang akan dikirim.",
  "content_hash": "a1b2c3d4e5f6...",
  "status": "processed",
  "created_at": "2026-06-18T09:00:00Z",
  "updated_at": "2026-06-18T09:01:00Z",
  "processing": {
    "article_id": "550e8400-e29b-41d4-a716-446655440099",
    "stemmed_content": "hasil stemming...",
    "word_frequencies": {"kata": 5, "artikel": 3},
    "stemming_status": "done",
    "wordcloud_status": "done",
    "processing_started_at": "2026-06-18T09:00:01Z",
    "processing_finished_at": "2026-06-18T09:00:05Z"
  }
}
```

**Response 400 Bad Request**

```json
{"error": "sender_id, receiver_id, title, and content are required"}
```

```json
{"error": "idempotency_key is required"}
```

---

## GET /articles/:id

Mengambil detail artikel beserta status dan hasil pemrosesan.

**Parameter**

| Parameter | Lokasi | Tipe | Keterangan |
| --- | --- | --- | --- |
| `id` | path | UUID string | ID artikel |

**Response 200 OK**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440099",
  "sender_id": "uuid-pengirim",
  "receiver_id": "uuid-penerima",
  "title": "Judul Artikel",
  "content": "Isi lengkap artikel.",
  "content_hash": "a1b2c3d4e5f6...",
  "status": "processed",
  "created_at": "2026-06-18T09:00:00Z",
  "updated_at": "2026-06-18T09:01:00Z",
  "processing": {
    "article_id": "550e8400-e29b-41d4-a716-446655440099",
    "stemmed_content": "hasil stemming dari isi artikel",
    "word_frequencies": {
      "artikel": 3,
      "lengkap": 2,
      "hasil": 1
    },
    "stemming_status": "done",
    "wordcloud_status": "done",
    "processing_started_at": "2026-06-18T09:00:01Z",
    "processing_finished_at": "2026-06-18T09:00:05Z"
  }
}
```

**Nilai `status` artikel:**

| Status | Keterangan |
| --- | --- |
| `queued` | Artikel baru masuk, belum diproses |
| `processed` | Stemming dan word cloud selesai |
| `degraded` | Salah satu proses gagal, yang lain berhasil |
| `failed` | Kedua proses gagal |

**Nilai `stemming_status` / `wordcloud_status`:**

| Status | Keterangan |
| --- | --- |
| `queued` | Belum diproses |
| `done` | Berhasil diproses |
| `failed` | Gagal setelah retry |

**Response 404 Not Found**

```json
{"error": "article not found"}
```

---

## GET /users/:id/inbox

Mengambil daftar artikel yang diterima oleh user tertentu, diurutkan dari yang terbaru.

**Parameter**

| Parameter | Lokasi | Tipe | Keterangan |
| --- | --- | --- | --- |
| `id` | path | UUID string | ID user penerima |

**Response 200 OK**

```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440099",
    "sender_id": "uuid-pengirim",
    "sender_name": "Maulana Faris Al Ghifari",
    "title": "Judul Artikel",
    "status": "processed",
    "created_at": "2026-06-18T09:00:00Z"
  },
  {
    "id": "550e8400-e29b-41d4-a716-446655440098",
    "sender_id": "uuid-pengirim-lain",
    "sender_name": "Raditya Nathaniel Nugroho",
    "title": "Artikel Kedua",
    "status": "queued",
    "created_at": "2026-06-18T08:55:00Z"
  }
]
```

Maksimal 100 artikel, diurutkan `created_at DESC`.

**Response 404 Not Found**

```json
{"error": "user not found"}
```

---

## Rate Limiting

Semua endpoint dilindungi rate limiter berbasis Redis. Default: **60 request per menit** per IP.

**Response 429 Too Many Requests**

```json
{"error": "rate limit exceeded, max 60 requests per minute"}
```

Header `Retry-After: 60` akan disertakan.

---

## CORS

API mendukung CORS untuk frontend di `localhost:3000`:

- Allowed Origins: `*`
- Allowed Methods: `GET, POST, PUT, DELETE, OPTIONS`
- Allowed Headers: `Content-Type, Idempotency-Key, Authorization`

---

## Contoh Penggunaan (curl)

### Health check

```bash
curl http://localhost:8080/health
```

### Ambil daftar user

```bash
curl http://localhost:8080/users
```

### Kirim artikel

```bash
curl -X POST http://localhost:8080/articles \
  -H "Content-Type: application/json" \
  -d '{
    "sender_id": "ID_PENGIRIM",
    "receiver_id": "ID_PENERIMA",
    "title": "Artikel Percobaan",
    "content": "Ini adalah isi artikel percobaan untuk menguji pipeline.",
    "idempotency_key": "test-key-001"
  }'
```

### Lihat detail artikel

```bash
curl http://localhost:8080/articles/ID_ARTIKEL
```

### Lihat inbox user

```bash
curl http://localhost:8080/users/ID_PENERIMA/inbox
```
