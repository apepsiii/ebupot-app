# Panduan Deployment — e-Bupot Portal

Tutorial deploy aplikasi e-Bupot Portal ke VPS menggunakan **binary + systemd** (tanpa Docker).

---

## Daftar Isi

1. [Prasyarat](#1-prasyarat)
2. [Cross-compile Binary (Windows)](#2-cross-compile-binary-windows)
3. [Setup VPS](#3-setup-vps)
4. [Konfigurasi Production](#4-konfigurasi-production)
5. [Systemd Service](#5-systemd-service)
6. [Nginx Reverse Proxy](#6-nginx-reverse-proxy)
7. [SSL / HTTPS (Let's Encrypt)](#7-ssl--https-lets-encrypt)
8. [Verifikasi](#8-verifikasi)
9. [Deployment Otomatis (Script)](#9-deployment-otomatis-script)
10. [Operasional](#10-operasional)

---

## 1. Prasyarat

| Kebutuhan | Keterangan |
|-----------|------------|
| VPS | Linux (Ubuntu 20.04+/Debian 12+), minimal 1GB RAM |
| Domain | Sudah diarahkan (A record) ke IP VPS |
| Port | 80 (HTTP) & 443 (HTTPS) terbuka |
| Akses | SSH dengan sudo/root |
| Komputer lokal | Go 1.21+ terinstall (untuk cross-compile) |

---

## 2. Cross-compile Binary (Windows)

Buka PowerShell di folder proyek:

```powershell
$env:GOOS="linux"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"
go build -ldflags="-s -w" -o ebupot-app .
```

> `CGO_ENABLED=0` wajib karena aplikasi menggunakan driver SQLite pure-Go (`glebarez/sqlite`), bukan `mattn/go-sqlite3` yang butuh C compiler.

Hasil: file `ebupot-app` (binary Linux amd64, ukuran ~15-20MB).

Reset environment setelah build:

```powershell
$env:GOOS=""; $env:GOARCH=""; $env:CGO_ENABLED=""
```

---

## 3. Setup VPS

### 3.1 Buat folder instalasi

SSH ke VPS:

```bash
ssh user@IP_VPS
```

```bash
sudo mkdir -p /opt/ebupot-app
sudo chown -R $USER:$USER /opt/ebupot-app
```

### 3.2 Upload file dari komputer lokal

Buka terminal baru di komputer lokal (Windows), upload file berikut:

```powershell
# Upload binary
scp ebupot-app user@IP_VPS:/opt/ebupot-app/

# Upload folder templates & public (wajib, dibaca dari disk saat runtime)
scp -r templates user@IP_VPS:/opt/ebupot-app/
scp -r public user@IP_VPS:/opt/ebupot-app/

# Upload config
scp config.yaml user@IP_VPS:/opt/ebupot-app/
scp .env.example user@IP_VPS:/opt/ebupot-app/
```

### 3.3 Buat folder data & uploads

Kembali ke SSH VPS:

```bash
cd /opt/ebupot-app
mkdir -p data uploads/ebupots
chmod +x ebupot-app
```

### 3.4 Buat user service (opsional, jika belum ada www-data)

```bash
id -u www-data &>/dev/null || sudo useradd -r -s /usr/sbin/nologin www-data
sudo chown -R www-data:www-data /opt/ebupot-app/data /opt/ebupot-app/uploads
```

---

## 4. Konfigurasi Production

### 4.1 Buat file .env

```bash
cd /opt/ebupot-app
cp .env.example .env
nano .env
```

### 4.2 Isi .env production

```bash
# Mode
APP_ENV=production

# Server
SERVER_HOST=127.0.0.1          # lokal saja, akses publik via Nginx
SERVER_PORT=8080
SERVER_DOMAIN=ebupot.domainanda.com   # domain publik ANDA (tanpa http)
SERVER_SCHEME=                  # kosong = auto-detect https dari Nginx

# Database
DB_PATH=data/ebupot.db

# Upload
UPLOAD_MAX_SIZE_MB=100

# QR Code
QR_RECOVERY_LEVEL=high
QR_SIZE=512

# Session (WAJIB ganti!)
SESSION_SECRET=hasilkan-dengan-openssl-di-bawah
SESSION_MAX_AGE=86400
```

### 4.3 Generate SESSION_SECRET yang aman

```bash
openssl rand -hex 32
```

Copy hasilnya ke `SESSION_SECRET=` di `.env`.

### 4.4 Penjelasan SERVER_DOMAIN & SERVER_SCHEME

| Config | Fungsi |
|--------|--------|
| `SERVER_DOMAIN` | Host yang dipakai di URL QR Code & Share Link. Wajib diisi domain publik Anda. |
| `SERVER_SCHEME` | `http` / `https` / kosong. Kosong = auto-detect dari header `X-Forwarded-Proto` Nginx. Di production, biarkan kosong. |

**Contoh URL yang dihasilkan:**
```
https://ebupot.domainanda.com/documentmanagementportal/api/DocumentExternalLink/3a57cbff-ba46-47f0-8400-5b7be5c78aba
```

---

## 5. Systemd Service

### 5.1 Buat service file

```bash
sudo nano /etc/systemd/system/ebupot.service
```

Isi (sesuaikan path jika berbeda):

```ini
[Unit]
Description=e-Bupot Portal & QR Generator
After=network.target

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=/opt/ebupot-app
ExecStart=/opt/ebupot-app/ebupot-app
EnvironmentFile=/opt/ebupot-app/.env
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=ebupot
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/ebupot-app/data /opt/ebupot-app/uploads

[Install]
WantedBy=multi-user.target
```

> Template ini juga tersedia di `deploy/ebupot.service` di repo.

### 5.2 Aktifkan & jalankan

```bash
sudo systemctl daemon-reload
sudo systemctl enable ebupot
sudo systemctl start ebupot
```

### 5.3 Cek status

```bash
sudo systemctl status ebupot
```

Output harus menampilkan `active (running)`.

### 5.4 Lihat log

```bash
sudo journalctl -u ebupot -f
```

Tekan `Ctrl+C` untuk keluar.

---

## 6. Nginx Reverse Proxy

### 6.1 Install Nginx

```bash
sudo apt update
sudo apt install -y nginx
```

### 6.2 Buat config

```bash
sudo nano /etc/nginx/sites-available/ebupot
```

Isi (ganti `ebupot.domainanda.com` dengan domain Anda):

```nginx
server {
    listen 80;
    server_name ebupot.domainanda.com;

    # Upload file besar (100MB)
    client_max_body_size 100M;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Timeout untuk upload file besar
        proxy_read_timeout 120s;
        proxy_send_timeout 120s;
    }
}
```

> Template ini juga tersedia di `deploy/nginx.conf` di repo.

### 6.3 Aktifkan config

```bash
sudo ln -sf /etc/nginx/sites-available/ebupot /etc/nginx/sites-enabled/ebupot
sudo rm -f /etc/nginx/sites-enabled/default
```

### 6.4 Test & reload

```bash
sudo nginx -t
sudo systemctl reload nginx
```

Saat ini aplikasi sudah bisa diakses via `http://ebupot.domainanda.com`.

---

## 7. SSL / HTTPS (Let's Encrypt)

### 7.1 Install Certbot

```bash
sudo apt install -y certbot python3-certbot-nginx
```

### 7.2 Generate sertifikat SSL

```bash
sudo certbot --nginx -d ebupot.domainanda.com
```

Jawab prompt:
- Email: masukkan email Anda
- Agree terms: `Y`
- Redirect HTTP → HTTPS: pilih `2` (redirect otomatis)

Certbot otomatis memodifikasi config Nginx untuk menambahkan SSL dan redirect.

### 7.3 Verifikasi auto-renewal

```bash
sudo certbot renew --dry-run
```

Sertifikat akan auto-renew setiap 60 hari.

---

## 8. Verifikasi

### 8.1 Cek semua service

```bash
sudo systemctl status ebupot    # aplikasi (active)
sudo systemctl status nginx     # web server (active)
```

### 8.2 Test akses

Buka browser:

```
https://ebupot.domainanda.com
```

Halaman login DJP-style harus tampil. Login dengan:

| Username | Password |
|----------|----------|
| `admin` | `admin123` |

> **Ganti password admin segera setelah login pertama!**

### 8.3 Test QR & Share Link

1. Login sebagai admin → Manajemen e-Bupot.
2. Upload dokumen PDF.
3. Klik tombol **QR Code** → QR image harus berisi URL `https://ebupot.domainanda.com/...`
4. Klik tombol **Share Link** → URL yang ditampilkan harus sama dengan QR.
5. Buka URL di browser lain → file PDF harus terunduh otomatis.

---

## 9. Deployment Otomatis (Script)

Alih-alih manual, gunakan script otomatis dari komputer Windows:

### deploy.ps1 (Windows → VPS)

```powershell
.\deploy\deploy.ps1 -VpsHost user@IP_VPS -Domain ebupot.domainanda.com
```

Script ini otomatis:
1. Cross-compile binary linux/amd64
2. Upload binary + templates + public + config ke VPS
3. Jalankan `install.sh` di VPS

### install.sh (di VPS)

Script `install.sh` otomatis:
1. Buat folder data & uploads
2. Generate `.env` production dengan SESSION_SECRET random
3. Setup systemd service + start
4. Install & konfigurasi Nginx
5. Setup SSL Let's Encrypt

Opsi:
```powershell
# Dengan SSH key tertentu
.\deploy\deploy.ps1 -VpsHost root@IP -Domain domain.com -SshKey ~/.ssh/id_rsa

# Skip SSL (testing dulu)
.\deploy\deploy.ps1 -VpsHost root@IP -Domain domain.com -SkipSSL
```

---

## 10. Operasional

### Restart aplikasi

```bash
sudo systemctl restart ebupot
```

### Update aplikasi (versi baru)

Dari komputer lokal:

```powershell
# Cross-compile ulang
$env:GOOS="linux"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"
go build -ldflags="-s -w" -o ebupot-app .

# Upload binary baru
scp ebupot-app user@IP_VPS:/opt/ebupot-app/

# Upload templates/public jika ada perubahan
scp -r templates user@IP_VPS:/opt/ebupot-app/
scp -r public user@IP_VPS:/opt/ebupot-app/
```

Di VPS:

```bash
sudo systemctl restart ebupot
```

### Backup

```bash
# Backup database + file upload
cd /opt/ebupot-app
sudo tar -czf /tmp/ebupot-backup-$(date +%Y%m%d).tar.gz data/ uploads/

# Pindahkan ke lokasi aman
scp user@IP_VPS:/tmp/ebupot-backup-*.tar.gz ./
```

### Restore

```bash
cd /opt/ebupot-app
sudo tar -xzf /tmp/ebupot-backup-YYYYMMDD.tar.gz
sudo chown -R www-data:www-data data/ uploads/
sudo systemctl restart ebupot
```

### Troubleshooting

| Masalah | Solusi |
|---------|--------|
| Service tidak start | `sudo journalctl -u ebupot -e` — cek error log |
| 502 Bad Gateway | Pastikan aplikasi jalan di 127.0.0.1:8080, cek `SERVER_PORT` di .env |
| Upload gagal (413) | Tambah `client_max_body_size 100M;` di Nginx config |
| QR URL pakai http, bukan https | Pastikan Nginx mengirim `X-Forwarded-Proto`, atau set `SERVER_SCHEME=https` di .env |
| SSL gagal | Pastikan DNS domain sudah mengarah ke IP VPS sebelum jalankan certbot |
| Permission denied | `sudo chown -R www-data:www-data /opt/ebupot-app/data /opt/ebupot-app/uploads` |

---

## Checklist Deployment

- [ ] DNS domain mengarah ke IP VPS
- [ ] Binary `ebupot-app` ter-upload ke `/opt/ebupot-app/`
- [ ] Folder `templates/` & `public/` ter-upload
- [ ] `.env` production dibuat dengan:
  - [ ] `APP_ENV=production`
  - [ ] `SERVER_HOST=127.0.0.1`
  - [ ] `SERVER_DOMAIN=domain-anda.com`
  - [ ] `SESSION_SECRET` diganti (bukan default)
- [ ] Folder `data/` & `uploads/` dibuat
- [ ] systemd service aktif (`systemctl status ebupot`)
- [ ] Nginx config aktif dengan `client_max_body_size 100M`
- [ ] SSL Let's Encrypt aktif
- [ ] Password admin diganti
- [ ] QR & Share Link URL menggunakan `https://domain-anda.com/...`
