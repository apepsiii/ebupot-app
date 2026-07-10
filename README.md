# e-Bupot Portal & QR Generator

Aplikasi manajemen dokumen e-Bupot (bukti potong pajak) terpusat dengan generator QR Code. Dibangun dengan **Golang (Gin Framework)**, **SQLite (GORM)**, dan **HTML/Bootstrap**.

Administrator dapat mengelola akun klien, mengunggah dokumen PDF bukti potong berdasarkan periode bulan/tahun, dan mendistribusikannya via QR Code. Wajib Pajak (User) dapat login untuk melihat dan mengunduh dokumen miliknya. Pemindai QR Code langsung mendapatkan file PDF tanpa perlu login.

---

## Fitur Utama

- **Autentikasi & Role-Based Access** — Admin & User dengan session-based middleware (bcrypt hashing).
- **Manajemen User (CRUD)** — Admin membuat/mengedit/menghapus akun klien.
- **Manajemen e-Bupot (CRUD)** — Upload PDF, pilih pemilik & periode, auto-generate UUID.
- **Nama File Otomatis** — Format kustom saat upload:
  ```
  M_01-DOC002_Ebupot_2126_BPA1_25087FU42.pdf
  ```
  | Bagian | Keterangan |
  |--------|------------|
  | `M_01` | ID User (2 digit) |
  | `DOC002` | Dokumen ke-N milik user (3 digit) |
  | `2126` | tanggal + bulan + tahun (2 digit) |
  | `BPA1` | kode tetap |
  | `25087FU42` | 9 karakter random alfanumerik |
- **QR Code Generator dengan Logo** — QR menyimpan URL format DJP Coretax, dengan logo yang dapat diunggah ditempel di tengah (error correction High).
- **Direct Download via QR** — Endpoint `/documentmanagementportal/api/DocumentExternalLink/:uuid` memaksa browser mengunduh PDF.
- **Isolasi Data** — User hanya melihat dokumen miliknya sendiri.
- **Halaman Error Elegan** — Tampilan 404 HTML untuk UUID tidak valid/kedaluwarsa.

### Format URL QR Code

```
https://[DOMAIN_ANDA]/documentmanagementportal/api/DocumentExternalLink/{UUID}
```

---

## Tech Stack

| Komponen | Teknologi |
|----------|-----------|
| Backend | Golang + Gin Framework |
| Database | SQLite + GORM |
| QR Code | skip2/go-qrcode |
| Logo Overlay | golang.org/x/image |
| Session | gin-contrib/sessions (cookie store) |
| Frontend | HTML + Bootstrap 5 + Bootstrap Icons |

---

## Struktur Proyek

```
ebupot-app/
├── config/              # Konfigurasi & koneksi database
│   ├── config.go        # Loader config.yaml + .env
│   └── database.go
├── controllers/         # Logika (auth, admin, user, download, helpers)
│   ├── auth_controller.go
│   ├── admin_controller.go
│   ├── user_controller.go
│   ├── download_controller.go
│   └── helpers.go
├── middlewares/         # Middleware autentikasi session
│   └── auth_middleware.go
├── models/              # Struct GORM (User, Ebupot)
│   ├── user.go
│   └── ebupot.go
├── routes/              # Definisi routing Gin + FuncMap template
│   └── routes.go
├── templates/           # View HTML (layout, auth, admin, user, errors)
├── public/              # Aset statis (CSS, JS)
├── uploads/             # Penyimpanan PDF & logo (TERPROTEKSI)
│   └── ebupots/
├── data/                # File database SQLite
├── config.yaml          # Pengaturan aplikasi (default)
├── .env.example         # Template environment variables
├── .env                 # Override & secrets (TIDAK di-commit)
├── main.go
├── go.mod
└── go.sum
```

---

## Persyaratan Sistem

- **Go 1.21+** (dites pada Go 1.26)
- Tidak membutuhkan CGO (menggunakan driver SQLite pure-Go `glebarez/sqlite`)

---

## Instalasi & Menjalankan

### 1. Clone repository

```bash
git clone https://github.com/apepsiii/ebupot-app.git
cd ebupot-app
```

### 2. Download dependencies

```bash
go mod download
```

### 3. Konfigurasi

Aplikasi membaca pengaturan dari dua sumber (`.env` meng-override `config.yaml`):

**a. `config.yaml`** (sudah ada, berisi default non-rahasia):

```yaml
app:
  name: "e-Bupot Portal"
  env: "development"          # development | production
server:
  host: "0.0.0.0"
  port: "8080"
  domain: "localhost:8080"    # domain publik untuk URL QR Code
database:
  path: "data/ebupot.db"
upload:
  max_size_mb: 5
  dir: "uploads/ebupots"
  logo_path: "uploads/logo.png"
qr:
  recovery_level: "high"      # low | medium | high | highest
  size: 512
session:
  secret: "change-this-in-env-to-a-long-random-string"
  max_age: 86400
```

**b. `.env`** (rahasia/override, **jangan di-commit**):

```bash
cp .env.example .env
```

Edit `.env` terutama `SESSION_SECRET` dan `SERVER_DOMAIN`:

| Env Var | Default | Keterangan |
|---------|---------|------------|
| `APP_ENV` | `development` | `production` untuk nonaktifkan debug |
| `SERVER_HOST` | `0.0.0.0` | `127.0.0.1` untuk lokal saja |
| `SERVER_PORT` | `8080` | Port aplikasi |
| `SERVER_DOMAIN` | `localhost:8080` | Domain publik untuk URL QR (tanpa http) |
| `DB_PATH` | `data/ebupot.db` | Lokasi file SQLite |
| `UPLOAD_MAX_SIZE_MB` | `100` | Batas ukuran upload PDF |
| `QR_RECOVERY_LEVEL` | `high` | Level koreksi error QR |
| `QR_SIZE` | `512` | Ukuran gambar QR (px) |
| `SESSION_SECRET` | *(default)* | **WAJIB ganti** di production |
| `SESSION_MAX_AGE` | `86400` | Umur sesi (detik) |

### 4. Jalankan aplikasi

```bash
go run main.go
```

Aplikasi berjalan di `http://localhost:8080` (sesuai `SERVER_PORT`).

### 4. Login default

Akun admin otomatis dibuat pada first run:

| Username | Password |
|----------|----------|
| `admin` | `admin123` |

> **Penting:** Segera ganti password admin setelah login pertama untuk keamanan.

---

## Penggunaan

### Admin

1. Login dengan akun admin.
2. **Manajemen User** → tambah akun klien (Nama, Username, Password).
3. **Manajemen e-Bupot** → upload PDF, pilih pemilik & periode (bulan/tahun). Nama file otomatis dibuat dengan format kustom.
4. Klik ikon **QR Code** pada baris dokumen untuk generate & unduh QR.
5. **Pengaturan** → upload logo yang akan tampil di tengah setiap QR Code.

### User (Klien)

1. Login dengan akun yang didaftarkan admin.
2. Filter dokumen berdasarkan bulan/tahun.
3. Klik **Download** untuk mengunduh bukti potong.

### Pemindai QR Code

- Pindai QR Code → browser otomatis mengunduh file PDF (tanpa login).

---

## Konfigurasi Produksi

Set di file `.env`:

```bash
APP_ENV=production
SERVER_DOMAIN=ebupot.domainanda.com
SESSION_SECRET=string-random-yang-sangat-panjang-dan-unik
```

---

## Deployment

Aplikasi membutuhkan **VPS** (bukan shared hosting) agar path URL DJP Coretax berfungsi penuh. Ada dua metode: **Docker** (termudah) atau **Binary + Systemd** (tanpa Docker).

### Metode A: Docker (Direkomendasikan)

#### Prasyarat di VPS
```bash
# Install Docker & Docker Compose
curl -fsSL https://get.docker.com | sh
```

#### Langkah部署

```bash
# 1. Clone repo ke VPS
git clone https://github.com/apepsiii/ebupot-app.git
cd ebupot-app

# 2. Buat .env production
cp .env.example .env
nano .env
#    Set: APP_ENV=production
#         SERVER_DOMAIN=ebupot.domainanda.com
#         SESSION_SECRET=<string-random-panjang>

# 3. Build & jalankan
docker compose up -d --build

# 4. Cek status
docker compose logs -f
```

Database & file upload tersimpan di folder `data/` dan `uploads/` (volume bind mount), aman saat container di-rebuild.

```bash
# Stop:      docker compose down
# Rebuild:   docker compose up -d --build
# Logs:      docker compose logs -f
```

---

### Metode B: Binary + Systemd (Tanpa Docker)

#### 1. Cross-compile dari komputer lokal (Windows)

```powershell
$env:GOOS="linux"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"
go build -ldflags="-s -w" -o ebupot-app .
```

Hasil: file binary `ebupot-app` (Linux amd64, tanpa CGO).

#### 2. Siapkan folder di VPS

```bash
sudo mkdir -p /opt/ebupot-app
sudo chown $USER:$USER /opt/ebupot-app
```

#### 3. Upload file ke VPS

File yang **wajib** di-upload (via SCP/rsync):

```
ebupot-app          # binary
templates/          # folder lengkap
public/             # folder lengkap
config.yaml         # konfigurasi default
.env                # konfigurasi production (secrets)
```

Contoh dengan SCP:
```bash
# Dari komputer lokal
scp -r ebupot-app templates public config.yaml .env user@IP_VPS:/opt/ebupot-app/
```

Atau clone langsung di VPS lalu build di sana (butuh Go terinstall):
```bash
git clone https://github.com/apepsiii/ebupot-app.git /opt/ebupot-app
cd /opt/ebupot-app
go build -ldflags="-s -w" -o ebupot-app .
```

#### 4. Buat folder data & uploads

```bash
cd /opt/ebupot-app
mkdir -p data uploads/ebupots
```

#### 5. Setup .env production

```bash
cp .env.example .env
nano .env
```

Isi:
```bash
APP_ENV=production
SERVER_HOST=127.0.0.1          # hanya lokal, akses publik via Nginx
SERVER_PORT=8080
SERVER_DOMAIN=ebupot.domainanda.com
SESSION_SECRET=ganti-dengan-secret-random-yang-panjang
UPLOAD_MAX_SIZE_MB=100
```

#### 6. Buat systemd service

```bash
sudo cp deploy/ebupot.service /etc/systemd/system/
sudo nano /etc/systemd/system/ebupot.service   # sesuaikan path jika perlu
sudo systemctl daemon-reload
sudo systemctl enable ebupot
sudo systemctl start ebupot

# Cek status
sudo systemctl status ebupot
sudo journalctl -u ebupot -f    # lihat log
```

#### 7. Setup Nginx + HTTPS (Let's Encrypt)

```bash
# Install Nginx & Certbot
sudo apt install -y nginx certbot python3-certbot-nginx

# Copy config Nginx
sudo cp deploy/nginx.conf /etc/nginx/sites-available/ebupot
sudo ln -s /etc/nginx/sites-available/ebupot /etc/nginx/sites-enabled/
sudo nano /etc/nginx/sites-available/ebupot   # ganti ebupot.domainanda.com

# Test & reload
sudo nginx -t
sudo systemctl reload nginx

# Setup SSL gratis (Let's Encrypt)
sudo certbot --nginx -d ebupot.domainanda.com
```

File template `deploy/nginx.conf` dan `deploy/ebupot.service` sudah disediakan di repo.

---

### Checklist Production

- [ ] `APP_ENV=production` di `.env`
- [ ] `SESSION_SECRET` diganti (bukan default)
- [ ] `SERVER_DOMAIN` diisi domain publik (untuk URL QR)
- [ ] `SERVER_HOST=127.0.0.1` (akses hanya via Nginx)
- [ ] Nginx reverse proxy aktif dengan `X-Forwarded-Proto` header
- [ ] SSL/HTTPS aktif (Let's Encrypt)
- [ ] `client_max_body_size 100M` di Nginx (untuk upload 100MB)
- [ ] Password admin default diganti
- [ ] Backup berkala folder `data/` (database SQLite)

### Backup & Restore

```bash
# Backup database & uploads
tar -czf ebupot-backup-$(date +%Y%m%d).tar.gz data/ uploads/

# Restore
tar -xzf ebupot-backup-YYYYMMDD.tar.gz
sudo systemctl restart ebupot
```

### Update Aplikasi

```bash
# Docker
cd /opt/ebupot-app
git pull
docker compose up -d --build

# Binary
cd /opt/ebupot-app
git pull
go build -ldflags="-s -w" -o ebupot-app .
sudo systemctl restart ebupot
```

---

## Endpoint / Routing

### Publik (tanpa auth)

| Method | Path | Deskripsi |
|--------|------|-----------|
| GET | `/` | Landing page |
| GET | `/login` | Halaman login |
| POST | `/login` | Proses login |
| GET | `/logout` | Logout |
| GET | `/documentmanagementportal/api/DocumentExternalLink/:uuid` | Download PDF via QR |

### Admin (middleware `RequireAdmin`)

| Method | Path | Deskripsi |
|--------|------|-----------|
| GET | `/admin/dashboard` | Dashboard admin |
| GET | `/admin/users` | Daftar user |
| POST | `/admin/users` | Tambah user |
| POST | `/admin/users/update/:id` | Edit user |
| POST | `/admin/users/delete/:id` | Hapus user |
| GET | `/admin/ebupots` | Daftar e-Bupot |
| POST | `/admin/ebupots` | Upload e-Bupot |
| POST | `/admin/ebupots/update/:id` | Edit e-Bupot |
| POST | `/admin/ebupots/delete/:id` | Hapus e-Bupot |
| GET | `/admin/ebupots/download/:id` | Download e-Bupot |
| GET | `/admin/ebupots/qr/:uuid` | Generate QR Code |
| GET | `/admin/settings` | Halaman pengaturan/logo |
| GET | `/admin/settings/logo/preview` | Pratinjau logo |
| POST | `/admin/settings/logo` | Upload logo |
| POST | `/admin/settings/logo/delete` | Hapus logo |

### User (middleware `RequireUser`)

| Method | Path | Deskripsi |
|--------|------|-----------|
| GET | `/user/dashboard` | Dashboard user + filter |
| GET | `/user/ebupots/download/:id` | Download e-Bupot |

---

## Struktur Database

### Tabel `users`

| Kolom | Tipe | Keterangan |
|-------|------|------------|
| id | INTEGER PK | Auto increment |
| username | TEXT | Unique |
| password | TEXT | Hash bcrypt |
| name | TEXT | Nama lengkap |
| role | TEXT | `admin` / `user` |
| created_at | DATETIME | Waktu dibuat |

### Tabel `ebupots`

| Kolom | Tipe | Keterangan |
|-------|------|------------|
| id | INTEGER PK | Auto increment |
| user_id | INTEGER FK | Relasi ke users |
| bulan | INTEGER | 1–12 |
| tahun | INTEGER | Contoh: 2026 |
| file_name | TEXT | Nama file kustom |
| file_path | TEXT | Lokasi fisik PDF |
| uuid_link | TEXT | UUID v4 (unique) |
| created_at | DATETIME | Waktu upload |

---

## Keamanan

- Password di-hash dengan **bcrypt**.
- Akses rute dilindungi middleware berbasis session.
- Upload PDF divalidasi (ekstensi + magic bytes `%PDF`) dengan batas **5MB**.
- Folder `uploads/` tidak diekspos sebagai static public.
- Isolasi data: query user difilter ketat berdasarkan `user_id`.
- Hapus dokumen juga menghapus file fisik (mencegah orphan files).

---

## License

MIT
