# Panduan Deployment — e-Bupot Portal

Deployment dengan **single binary** — semua file (templates, CSS, JS, config) ter-embed di dalam binary. Saat pertama dijalankan di server, binary membuka **wizard instalasi interaktif**.

---

## Alur Singkat

```
[Windows]                    [VPS]
bash build.sh  ──scp──>  upload binary
                           │
                           v
                       ./ebupot-app
                           │
                     wizard muncul ──> port, domain, dll
                           │
                     .env dibuat
                     folder dibuat
                     systemd (opsional)
                     nginx (opsional)
                     SSL (opsional)
                           │
                     server jalan ✓
```

---

## 1. Build Binary (Windows)

Buka terminal (Git Bash / PowerShell) di folder proyek:

```bash
bash build.sh
```

Hasil: **1 file binary** `ebupot-app` (~26MB, linux/amd64, semua aset ter-embed).

---

## 2. Upload ke VPS

```bash
# Buat folder di VPS
ssh root@IP_VPS "mkdir -p /opt/ebupot-app"

# Upload binary (hanya 1 file!)
scp ebupot-app root@IP_VPS:/opt/ebupot-app/

# Set permission
ssh root@IP_VPS "chmod +x /opt/ebupot-app/ebupot-app"
```

> Tidak perlu upload `templates/`, `public/`, `config.yaml` — semuanya sudah embed di binary.

---

## 3. Jalankan Wizard

SSH ke VPS dan jalankan binary:

```bash
ssh root@IP_VPS
cd /opt/ebupot-app
./ebupot-app
```

Wizard otomatis muncul:

```
+======================================================+
|     e-Bupot Portal — Setup Wizard                    |
+======================================================+

File konfigurasi (.env) belum ditemukan.
Mari kita setup aplikasi ini.

[1] Port aplikasi [8080]: 8080
[2] Domain publik (untuk QR & Share Link) [localhost:8080]: ebupot.domainanda.com
[3] Scheme (http/https/kosong=auto-detect): 
[4] Maksimal upload (MB) [100]: 
[5] Auto-generate session secret? [Y/n]: y
    Secret dibuat: 07975c34fbfcba1fa1c6e1b51fc141cf...

--- Membuat konfigurasi ---
  [OK] File .env dibuat
  [OK] Folder data/ dibuat
  [OK] Folder uploads/ dibuat

Setup systemd service (auto-start saat boot)? [Y/n]: y
  [OK] Systemd service terpasang

Setup Nginx reverse proxy? [Y/n]: y
  Domain untuk Nginx [ebupot.domainanda.com]: 
  [OK] Nginx reverse proxy terpasang

Setup SSL (Let's Encrypt)? [y/N]: y
  Domain untuk SSL [ebupot.domainanda.com]: 
  [OK] SSL Let's Encrypt terpasang

+======================================================+
|              SETUP SELESAI!                          |
+======================================================+
|  Aplikasi : e-Bupot Portal                          |
|  Port     : 8080                                     |
|  Domain   : ebupot.domainanda.com                   |
|  Scheme   : auto-detect                              |
+======================================================+

Server berjalan di http://localhost:8080
```

Server langsung jalan setelah wizard selesai.

---

## 4. Akses

```
https://ebupot.domainanda.com
```

Login default:

| Username | Password |
|----------|----------|
| `admin` | `admin123` |

> **Ganti password admin segera!**

---

## 5. Operasional

### Start / Stop / Restart

```bash
# Via systemd (jika dipasang saat wizard)
sudo systemctl start ebupot
sudo systemctl stop ebupot
sudo systemctl restart ebupot
sudo systemctl status ebupot

# Atau langsung
cd /opt/ebupot-app && ./ebupot-app
```

### Lihat log

```bash
# Via systemd
sudo journalctl -u ebupot -f

# Atau langsung (output ke terminal)
./ebupot-app
```

### Edit konfigurasi

```bash
nano /opt/ebupot-app/.env
sudo systemctl restart ebupot
```

### Update aplikasi

```bash
# Di Windows: build ulang
bash build.sh

# Upload binary baru (overwrite)
scp ebupot-app root@IP_VPS:/opt/ebupot-app/
```

```bash
# Di VPS: restart
ssh root@IP_VPS "chmod +x /opt/ebupot-app/ebupot-app && sudo systemctl restart ebupot"
```

> .env, database, dan uploads TIDAK terreplace — hanya binary yang diupdate.

### Backup

```bash
cd /opt/ebupot-app
sudo tar -czf /tmp/ebupot-backup-$(date +%Y%m%d).tar.gz data/ uploads/ .env
```

### Restore

```bash
cd /opt/ebupot-app
sudo tar -xzf /tmp/ebupot-backup-YYYYMMDD.tar.gz
sudo systemctl restart ebupot
```

---

## 6. Deployment Otomatis

### Opsi A: build.sh + scp (1 perintah build)

```bash
# Build binary
bash build.sh

# Upload + set permission
scp ebupot-app root@IP_VPS:/opt/ebupot-app/
ssh root@IP_VPS "chmod +x /opt/ebupot-app/ebupot-app"
```

Lalu SSH ke VPS dan jalankan `./ebupot-app` untuk memulai wizard.

### Opsi B: deploy.ps1 (build + upload otomatis)

```powershell
.\deploy\deploy.ps1 -VpsHost root@IP_VPS
```

Script ini otomatis build + upload binary, lalu menampilkan instruksi menjalankan wizard.

---

## 7. Troubleshooting

| Masalah | Solusi |
|---------|--------|
| Wizard tidak muncul | `.env` sudah ada — hapus dulu: `rm .env` lalu jalankan ulang |
| 502 Bad Gateway | Pastikan app jalan: `systemctl status ebupot` |
| Upload gagal (413) | Edit nginx config: `client_max_body_size 100M;` |
| QR URL pakai http | Set `SERVER_SCHEME=https` di `.env` atau pastikan Nginx kirim `X-Forwarded-Proto` |
| SSL gagal | Pastikan DNS domain sudah mengarah ke IP VPS sebelum wizard |
| Port bentrok | Ganti port di wizard atau edit `.env` `SERVER_PORT` |
| Permission denied | `sudo chown -R $USER:$USER /opt/ebupot-app` |

---

## 8. Menjalankan Ulang Wizard

Jika ingin reconfigure:

```bash
cd /opt/ebupot-app
rm .env          # hapus config lama
./ebupot-app     # wizard muncul lagi
```

---

## Checklist

- [ ] Binary ter-upload ke `/opt/ebupot-app/ebupot-app`
- [ ] `chmod +x ebupot-app`
- [ ] Wizard dijalankan, .env dibuat
- [ ] Domain benar di wizard (untuk QR & Share Link)
- [ ] Systemd service aktif (jika dipilih)
- [ ] Nginx reverse proxy aktif (jika dipilih)
- [ ] SSL Let's Encrypt aktif (jika dipilih)
- [ ] Password admin diganti
- [ ] QR & Share Link URL = `https://domain/...`
