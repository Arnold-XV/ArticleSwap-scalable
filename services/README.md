# Services ArticleSwap

Folder ini akan berisi implementasi service:

- `api-gateway`: HTTP API utama berbasis Go.
- `workers/stemmer`: worker stemming artikel.
- `workers/wordcloud`: worker penghitung frekuensi kata untuk word cloud.
- `workers/aggregator`: worker pengubah status akhir pipeline.

Dockerfile sudah disiapkan agar tahap implementasi berikutnya tinggal mengisi source code sesuai struktur masing-masing service.

