# Frontend ArticleSwap

Frontend ArticleSwap dibuat dengan **Next.js + TypeScript** dan seluruh tampilan utama memakai Bahasa Indonesia. Bagian ini adalah fondasi kerja frontend untuk Maulana Faris dan Raditya, sekaligus siap disambungkan ke backend Go ketika API sudah berjalan.

## Status Saat Ini

Sudah tersedia:

- form kirim artikel,
- pilihan pengirim dan penerima,
- inbox penerima,
- detail artikel,
- status pipeline,
- placeholder hasil stemming,
- placeholder word cloud,
- helper API,
- mock fallback jika backend belum hidup.

Belum menjadi scope frontend saat ini:

- visual word cloud final,
- polling otomatis berkala,
- autentikasi user,
- fitur edit/hapus artikel,
- logic backend, RabbitMQ, Redis, dan PostgreSQL.

## Cara Menjalankan Lokal

Masuk ke folder frontend:

```bash
cd frontend
```

Install dependency:

```bash
npm install
```

Jalankan development server:

```bash
npm run dev
```

Frontend berjalan di:

```text
http://localhost:3000
```

Build production:

```bash
npm run build
```

Cek TypeScript:

```bash
npm run typecheck
```

## Environment

Frontend membaca API backend dari:

```text
NEXT_PUBLIC_API_BASE_URL
```

Default jika env tidak diisi:

```text
http://localhost:8080
```

Contoh `.env.local` untuk frontend:

```env
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
```

## Mock Fallback

Frontend akan mencoba memanggil API asli terlebih dahulu. Jika API backend belum hidup atau request gagal, data otomatis memakai mock fallback dari `lib/mock-data.ts`.

Artinya:

- Faris/Raditya bisa mengerjakan UI tanpa menunggu backend selesai.
- Saat backend hidup, frontend otomatis memakai API asli.
- Struktur mock dibuat mirip kontrak API agar transisi mudah.

## Struktur File Penting

```text
frontend/
  app/
    layout.tsx       Layout utama Next.js
    page.tsx         Halaman utama ArticleSwap
    globals.css      Styling global frontend
  lib/
    api.ts           Helper fetch API + mock fallback
    mock-data.ts     Data mock untuk development awal
  types/
    articles.ts      TypeScript type untuk user, artikel, status, event
  package.json       Script dan dependency frontend
  next.config.js     Konfigurasi Next.js
```

## Kontrak API Yang Dipakai Frontend

Endpoint yang sudah disiapkan di `lib/api.ts`:

### `GET /users`

Mengambil daftar user.

Response:

```ts
type User = {
  id: string;
  name: string;
  email: string;
};
```

### `POST /articles`

Mengirim artikel baru.

Request:

```ts
type SubmitArticleInput = {
  senderId: string;
  receiverId: string;
  title: string;
  content: string;
  idempotencyKey: string;
};
```

Response:

```ts
type SubmitArticleResult = {
  articleId: string;
  status: string;
};
```

### `GET /users/:id/inbox`

Mengambil daftar artikel yang diterima user.

### `GET /articles/:id`

Mengambil detail artikel, isi mentah, hasil stemming, word frequencies, dan timeline pipeline.

## Status Pipeline

Status yang dipakai frontend:

| Status API | Label UI |
| --- | --- |
| `queued` | Menunggu |
| `processing` | Diproses |
| `processed` | Selesai |
| `degraded` | Sebagian Gagal |
| `failed` | Gagal |

Jika backend menambah status baru, update type `PipelineStatus` di `types/articles.ts` dan label di `app/page.tsx`.

## Tugas Lanjutan Untuk Raditya

Raditya bisa melanjutkan dari bagian berikut:

- memperbaiki visual word cloud agar lebih menarik,
- membuat komponen khusus untuk hasil stemming,
- menambahkan loading skeleton,
- menambahkan empty state yang lebih rapi,
- menambahkan polling otomatis detail artikel,
- membuat tampilan cache hit/degraded jika backend sudah mengirim datanya,
- menyiapkan screenshot UI untuk poster.

File utama yang kemungkinan disentuh:

- `app/page.tsx`
- `app/globals.css`
- `types/articles.ts`

## Tugas Lanjutan Untuk Backend

Backend perlu memastikan response API sesuai type di `types/articles.ts`.

Hal penting:

- `GET /users` harus mengembalikan array user.
- `POST /articles` harus menerima `idempotencyKey`.
- `GET /users/:id/inbox` harus mengembalikan artikel ringkas.
- `GET /articles/:id` harus mengembalikan detail artikel lengkap.
- Status backend sebaiknya memakai `queued`, `processing`, `processed`, `degraded`, atau `failed`.

Jika nama field backend berubah, update mapping di `lib/api.ts`, bukan langsung menyebar perubahan ke seluruh UI.

## Checklist Sebelum Commit Frontend

Jalankan:

```bash
npm run typecheck
npm run build
```

Cek manual:

- halaman bisa dibuka,
- daftar user muncul,
- artikel bisa dikirim dengan mock fallback,
- artikel masuk ke inbox,
- detail artikel bisa dipilih,
- status pipeline tampil dalam Bahasa Indonesia,
- tidak ada teks yang saling menumpuk di layar kecil.

## Catatan Untuk Tim

Frontend ini sengaja dibuat sebagai aplikasi operasional, bukan landing page. Fokusnya adalah demo alur ArticleSwap: kirim artikel, masuk inbox, lalu menampilkan status pipeline dan hasil pengolahan.
