# Laporan Pengujian Penjaminan Mutu (QA Report) — ArticleSwap

*   **Penyusun / QA Lead**: Ajie Armansyah Sunaryo (NIM: 24/545286/PA/23170)
*   **Tanggal Pengujian**: 20 Juni 2026
*   **Status Akhir Proyek**: **PASSED** (Semua skenario pengujian sukses memenuhi spesifikasi minimal)

---

## 1. Metodologi Pengujian
Pengujian dilakukan menggunakan dua metode utama untuk memastikan sistem berjalan secara fungsional dan mampu menangani beban trafik secara skalabel serta resilien:
1.  **Pengujian Fungsional Manual (Manual E2E Testing)**: Memvalidasi alur pengguna mulai dari submit artikel, pengolahan asinkron worker, pengiriman ke inbox penerima, verifikasi idempotensi, serta simulasi kegagalan worker (*degraded mode*).
2.  **Pengujian Beban Asinkron (Stress Testing k6)**: Mengguyur sistem dengan user virtual (VU) untuk mengukur latensi, throughput, keandalan connection pool database, dan efisiensi caching Redis.

---

## 2. Hasil Pengujian Fungsional (Manual Testing)

| ID Test | Skenario Pengujian | Langkah Pengujian | Hasil yang Diharapkan | Hasil Aktual | Status |
| :--- | :--- | :--- | :--- | :--- | :---: |
| **TF-01** | Submit & Delivery Artikel | Mengirim artikel dari User A ke User B via UI Frontend. | Artikel tersimpan di PostgreSQL dan langsung muncul di Inbox User B. | Artikel langsung tersimpan dan diterima dengan sukses di inbox. | **PASSED** |
| **TF-02** | Asynchronous Pipeline | Mengirim artikel, lalu memantau antrean RabbitMQ dan status pemrosesan. | API Gateway merespon instan (non-blocking). Status pemrosesan di-update secara berkala oleh worker. | Respons HTTP 201 instan. Status artikel berubah dari `queued` -> `processing` -> `completed`. | **PASSED** |
| **TF-03** | Pencegahan Duplikasi (Idempotency) | Menekan tombol "Submit" berkali-kali secara cepat untuk artikel yang sama. | API Gateway mendeteksi token `idempotencyKey` yang sama dan tidak menduplikasi baris di database. | Hanya satu artikel yang terbuat di PostgreSQL. Request duplikat mengembalikan ID artikel yang sama. | **PASSED** |
| **TF-04** | Caching Hasil Pemrosesan | Mengirim ulang artikel dengan teks isi konten yang persis sama. | Worker tidak melakukan kalkulasi stemming/wordcloud ulang melainkan mengambil langsung dari Redis Cache. | Latensi pemrosesan turun drastis (sub-milidetik) karena data dibaca dari Redis Cache. | **PASSED** |
| **TF-05** | Fault Isolation (Degraded Mode) | Mematikan `wordcloud-worker` (`docker compose stop`), lalu mengirim artikel. | Pengiriman artikel dan proses stemming tetap berjalan sukses. Halaman detail artikel menampilkan status degraded tanpa merubuhkan sistem. | Artikel sukses terkirim dan hasil stemming muncul. Bagian word cloud menampilkan indikator degraded secara anggun. | **PASSED** |

---

## 3. Hasil Pengujian Beban & Skalabilitas (Stress Testing k6)

Berikut adalah data pengujian beban riil menggunakan k6 yang dijalankan pada infrastruktur lokal Docker Compose:

### A. Pengujian Kinerja Baseline (`baseline.js`)
*   **Beban**: 10 Virtual Users (VU) selama 1 menit.
*   **Total HTTP Requests**: 2901 request (Rata-rata: 48.24 request/detik).
*   **Latensi**: Rata-rata **6.81 ms** | P95: **16.51 ms** | P99: **33.32 ms**.
*   **Error Rate**: 0.00%.
*   *Analisis*: Sistem beroperasi sangat responsif dan efisien di bawah beban trafik normal.

### B. Pengujian Beban Ekstrim (`stress.js`)
*   **Beban**: Ramping bertahap dari 10 hingga 100 Virtual Users (VU) selama 5 menit.
*   **Total HTTP Requests**: 55337 request (Rata-rata: 184.32 request/detik).
*   **Latensi**: Rata-rata **9.93 ms** | P95: **33.44 ms**.
*   **Error Rate**: 0.00%.
*   *Analisis*: Throughput melonjak hingga 4x lipat dari baseline dengan kenaikan latensi yang sangat minim, membuktikan skalabilitas tinggi dari API Gateway berbasis Go.

### C. Pengujian Idempotensi Beban (`idempotency-test.js`)
*   **Beban**: 5 Virtual Users (VU) mengirimkan request dengan idempotency key ganda.
*   **Total HTTP Requests**: 301 request.
*   **Checks Succeeded**: 100% (600 checks).
*   **Error Rate**: 0.00%.
*   *Analisis*: Tidak ditemukan data duplikat pada database PostgreSQL meskipun request dikirim secara konkuren dan agresif.

### D. Pengujian Pembatasan Trafik (`ratelimit-test.js`)
*   **Beban**: 1 Virtual User (VU) membombardir request berlebih.
*   **Total HTTP Requests**: 288 request.
*   **Allowed (HTTP 200)**: 60 request.
*   **Blocked (HTTP 429)**: 228 request (**Rate Block: 79.2%**).
*   *Analisis*: Rate Limiter berbasis Redis bekerja sempurna dalam menolak request berlebih guna melindungi layanan dari serangan Denial of Service (DoS).

### E. Pengujian Saturasi Koneksi Database (`pool-saturation-test.js`)
*   **Beban**: Ramping bertahap hingga 80 Virtual Users (VU) untuk membanjiri pool database.
*   **Total HTTP Requests**: 35677 request.
*   **DB-related Errors**: **0 error (0.00% Error Rate)**.
*   *Analisis*: Connection pooling (`pgxpool`) terbukti tangguh meredam lonjakan koneksi database tanpa menghasilkan galat koneksi putus atau jenuh.

### F. Pengujian End-to-End Pipeline & Ketahanan Worker (`pipeline-test.js` & `degraded-test.js`)
*   **Skenario Degraded Mode (Wordcloud Off)**: Mematikan worker wordcloud.
    *   **Hasil**: 100% checks berhasil (450 checks sukses), rata-rata latensi HTTP sangat cepat di angka **9.47 ms**. Ini membuktikan kegagalan modul tambahan terisolasi dengan baik.
*   **Skenario Pipeline E2E (Semua Worker Off)**: Mengirim 40 artikel saat worker mati.
    *   **Hasil**: API Gateway sukses menampung 40 artikel (Error Rate 0% di Gateway). Namun, status akhir artikel tertahan di status antrean (Success Rate Pipeline 0% akibat timeout). Ini membuktikan sifat *loosely coupled* di mana matinya worker pemrosesan tidak memengaruhi ketersediaan API Gateway utama.

---

## 4. Kesimpulan Evaluasi Penjaminan Mutu (QA)
Berdasarkan hasil pengujian di atas, aplikasi **ArticleSwap** telah memenuhi seluruh prinsip arsitektur aplikasi scalable yang baik:
1.  **Loosely Coupled**: Terbukti dari pengujian *degraded* dan *pipeline* di mana kegagalan worker latar belakang terisolasi dan tidak merambat ke API Gateway.
2.  **Resilien & Fault Tolerance**: Sistem menangani kegagalan transien dengan retry mechanism dan meredam kebanjiran request lewat rate limiter & connection pool.
3.  **High Performance**: Latensi rata-rata di bawah 10ms pada kondisi stres dengan tingkat kegagalan transaksi data sebesar 0%.

QA merekomendasikan proyek ini untuk **siap dirilis/didemokan**.
