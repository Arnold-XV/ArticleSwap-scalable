# Checklist Poster ArticleSwap

Project meminta laporan akhir berupa poster/infografis 2 halaman.

## Halaman 1 - Aplikasi ArticleSwap

Isi yang wajib ada:

- Judul: ArticleSwap - Platform Pertukaran Artikel Scalable dan Resilien.
- Deskripsi singkat aplikasi.
- Diagram arsitektur.
- Alur submit artikel sampai diterima user tujuan.
- Penjelasan RabbitMQ sebagai message broker.
- Penjelasan PostgreSQL connection pooling.
- Penjelasan Redis cache.
- Penjelasan worker scaling.
- Screenshot frontend.
- Screenshot RabbitMQ management atau Docker Compose.
- Tabel hasil stress testing.
- Grafik/angka before-after optimasi.

Contoh metrik:

| Skenario | Avg Latency | P95 | P99 | Throughput | Error Rate |
| --- | ---: | ---: | ---: | ---: | ---: |
| 1 worker tanpa cache | diisi hasil test | diisi | diisi | diisi | diisi |
| 3 worker + cache | diisi hasil test | diisi | diisi | diisi | diisi |

## Halaman 2 - Peran Anggota

Isi yang wajib ada:

- Nama dan NIM semua anggota.
- Peran setiap anggota.
- Komponen yang dikerjakan.
- Bukti kontribusi, misalnya screenshot fitur atau ringkasan commit.
- Bagian quality check bersama.
- Ringkasan pembelajaran dari Bab 10 dan Bab 11.

## Pembagian Peran

| Nama | NIM | Peran | Kontribusi Utama |
| --- | --- | --- | --- |
| Maulana Faris Al Ghifari | 24/544029/PA/23119 | Frontend Developer | Submit artikel, inbox, detail artikel, integrasi API |
| Raditya Nathaniel Nugroho | 24/543188/PA/23069 | Frontend Developer | Word cloud UI, hasil proses, loading/error state |
| Ajie Armansyah Sunaryo | 24/545286/PA/23170 | UI/UX Developer dan QA | Desain UI, testing manual, checklist demo |
| Arnoldus Dharma Wasesa M. | 24/545535/PA/23182 | Backend Developer | API Gateway, schema PostgreSQL, pooling, metrics |
| Aliya Khairun Nisa | 24/543832/PA/23111 | Backend Developer | RabbitMQ, worker, Redis cache, retry, circuit breaker |

## Quality Check Bersama

Checklist final:

- Semua endpoint berhasil dipakai frontend.
- Tidak ada artikel duplikat saat retry.
- Queue RabbitMQ berjalan.
- Redis cache terbukti dipakai.
- Worker scaling dapat dijalankan.
- Stress test selesai dan hasilnya dicatat.
- Poster memakai Bahasa Indonesia yang jelas.
- Demo dapat dijalankan ulang dari nol.

