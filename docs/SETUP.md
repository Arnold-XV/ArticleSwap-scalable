# Setup Lokal ArticleSwap

> **Singkatnya:** Kamu **tidak perlu install PostgreSQL, Redis, atau RabbitMQ** secara manual. Semuanya jalan di dalam Docker container. Yang perlu diinstall hanya Docker Desktop dan Go.

---

## Yang Perlu Diinstall

| Software | Untuk apa | Wajib? |
| --- | --- | --- |
| Docker Desktop | Menjalankan PostgreSQL, Redis, RabbitMQ, dan semua service via container | **Wajib** |
| Go 1.23+ | Build API Gateway secara lokal (opsional, bisa lewat Docker saja) | Disarankan |
| k6 | Menjalankan stress test | Perlu saat stress test |

---

## 1. Install Docker Desktop

Docker Desktop sudah bundel Docker Engine + Docker Compose. Tidak perlu install terpisah.

### Windows

1. Download Docker Desktop di: https://www.docker.com/products/docker-desktop/

2. Jalankan installer → ikuti wizard → **Restart Windows** saat diminta.

3. Buka Docker Desktop setelah restart. Tunggu sampai status di pojok kiri bawah berubah menjadi **"Engine running"** (ikon hijau).

4. Verifikasi di PowerShell:

```powershell
docker --version
docker compose version
```

> **Catatan WSL 2:** Docker Desktop di Windows pakai WSL 2. Saat install, Docker akan otomatis minta enable WSL 2 jika belum aktif. Ikuti saja instruksinya.

### macOS

1. Download Docker Desktop di: https://www.docker.com/products/docker-desktop/

2. Pilih sesuai chip:
   - **Apple Silicon (M1/M2/M3/M4)** → pilih "Mac with Apple Silicon"
   - **Intel** → pilih "Mac with Intel chip"

3. Buka file `.dmg` → drag Docker ke Applications → buka Docker dari Applications.

4. Tunggu sampai ikon Docker di menu bar berhenti animasi (status: **Running**).

5. Verifikasi di Terminal:

```bash
docker --version
docker compose version
```

Contoh output yang benar (sama untuk Windows dan Mac):
```
Docker version 27.x.x, build ...
Docker Compose version v2.x.x
```

---

## 2. Install Go (Opsional tapi Disarankan)

Go dibutuhkan jika ingin build atau test code Go di luar Docker.

### Windows

1. Download Go 1.23 di: https://go.dev/dl/
2. Pilih file **Windows AMD64** (`.msi`).
3. Jalankan installer → ikuti wizard.
4. Buka PowerShell baru, verifikasi:

```powershell
go version
```

Output: `go version go1.23.x windows/amd64`

### macOS

Pakai Homebrew (cara paling mudah):

```bash
brew install go
```

Atau download manual di https://go.dev/dl/ → pilih **macOS ARM64** (Apple Silicon) atau **macOS AMD64** (Intel).

Verifikasi:

```bash
go version
```

Output: `go version go1.23.x darwin/arm64` (atau `darwin/amd64`)

---

## 3. Install k6 (Untuk Stress Test)

### Windows

Pakai Chocolatey:

```powershell
choco install k6
```

Atau download manual di: https://github.com/grafana/k6/releases/latest
→ pilih `k6-v*-windows-amd64.zip` → extract → copy `k6.exe` ke PATH.

### macOS

Pakai Homebrew:

```bash
brew install k6
```

Verifikasi (keduanya):

```bash
k6 version
```

---

## 4. Jalankan ArticleSwap

Setelah Docker Desktop jalan, langkah berikut **sama untuk Windows dan Mac**:

### Windows (PowerShell)

```powershell
# 1. Masuk ke folder project
cd path\ke\ArticleSwap-scalable

# 2. Salin file environment
copy .env.example .env

# 3. Build dan jalankan semua service
docker compose up --build
```

### macOS (Terminal)

```bash
# 1. Masuk ke folder project
cd /path/ke/ArticleSwap-scalable

# 2. Salin file environment
cp .env.example .env

# 3. Build dan jalankan semua service
docker compose up --build
```

Docker akan otomatis:
- Pull image `postgres:16-alpine`, `redis:7-alpine`, `rabbitmq:3.13-management-alpine`
- Build image Go untuk API Gateway dari source code
- Jalankan `001_schema.sql` → buat semua tabel + seed 5 user
- Start semua container

### URL setelah jalan

| Service | URL |
| --- | --- |
| API Gateway | http://localhost:8080 |
| API Health Check | http://localhost:8080/health |
| RabbitMQ Management | http://localhost:15672 (user: `articleswap`, pass: `articleswap`) |
| Frontend (jika sudah dibuat) | http://localhost:3000 |

---

## 5. Verifikasi Semua Berjalan

Buka terminal/PowerShell baru (biarkan `docker compose up` tetap jalan di terminal lain):

```bash
# Cek health API Gateway
curl http://localhost:8080/health

# Cek daftar user (harus return 5 user)
curl http://localhost:8080/users
```

Response `/health` yang benar:
```json
{
  "status": "ok",
  "postgres": "up",
  "redis": "up",
  "rabbitmq": "up",
  "time": "..."
}
```

---

## 6. Matikan Service

```bash
# Stop semua container (data tetap tersimpan di Docker volume)
docker compose down

# Stop + hapus semua data (reset database ke kondisi awal)
docker compose down -v
```

---

## Troubleshooting

### Docker Desktop tidak mau start

**Windows:** Pastikan Virtualization aktif di BIOS (Intel VT-x / AMD-V). Cek di Task Manager → Performance → CPU → Virtualization: **Enabled**.

**Mac:** Coba quit Docker Desktop dari menu bar → buka ulang. Jika masih gagal, cek di System Settings → Privacy & Security apakah ada permission yang perlu di-allow.

### Port sudah dipakai

Jika muncul error `bind: address already in use`, ada aplikasi lain yang pakai port 5432/6379/5672/8080. Matikan aplikasinya atau ubah port di `.env`.

### Database tidak terinisialisasi

Jika tabel tidak terbuat, kemungkinan volume Postgres sudah ada dari run sebelumnya. Reset dengan:

```bash
docker compose down -v
docker compose up --build
```

### `docker compose` tidak dikenal

Pastikan pakai Docker Desktop versi baru. Kalau masih error, coba versi lama:

```bash
docker-compose up --build
```

---

## Ringkasan Singkat

### Windows

```powershell
# Install Docker Desktop → restart Windows → buka Docker Desktop
copy .env.example .env
docker compose up --build
curl http://localhost:8080/health
```

### macOS

```bash
# Install Docker Desktop → buka Docker Desktop → tunggu running
cp .env.example .env
docker compose up --build
curl http://localhost:8080/health
```

Itu saja. PostgreSQL, Redis, RabbitMQ semuanya otomatis. 🚀
