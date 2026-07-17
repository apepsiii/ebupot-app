# Prompt: Build & Install Pattern untuk Aplikasi Go (Gin)

Dokumen ini berisi instruksi/prompt yang dapat diberikan kepada AI agent untuk mengimplementasikan pola **single binary + setup wizard** pada proyek Go web aplikasi apa pun. Salin blok prompt di bawah dan sesuaikan nama proyek, framework, dan struktur sesuai kebutuhan.

---

## PROMPT UNTUK AI AGENT

```
Saya punya proyek Go web aplikasi (Gin Framework) dan ingin mengimplementasikan
sistem build & deployment dengan pola berikut:

### Tujuan
1. SEMUA file runtime (HTML templates, static assets CSS/JS, config YAML) di-embed
   ke dalam binary tunggal via Go `embed` package — sehingga deployment hanya
   butuh 1 file binary, tanpa perlu upload folder terpisah.
2. Saat binary dijalankan pertama kali dan file `.env` belum ada, binary membuka
   WIZARD INSTALASI INTERAKTIF di terminal yang menanyakan konfigurasi (port,
   domain, scheme, dll), lalu otomatis:
   - Generate file .env
   - Buat folder runtime (database, uploads)
   - Optional: pasang systemd service
   - Optional: pasang Nginx reverse proxy
   - Optional: setup SSL Let's Encrypt
   - Cetak pesan sukses/gagal per langkah
   - Jalankan server
3. Jika .env sudah ada, lewati wizard dan langsung jalankan server.

### Yang perlu dibuat:

#### A. File `embed.go` (package main, di root proyek)
- Embed direktori templates: `//go:embed all:templates` → `var templateFS embed.FS`
- Embed direktori public (CSS/JS/images): `//go:embed all:public` → `var publicFS embed.FS`
- Embed config default: `//go:embed config.yaml` → `var configYAML []byte`

#### B. Update config loader (`config/config.go`)
- `Load(yamlData []byte)` — parse embedded YAML sebagai default, lalu load .env
  (godotenv), lalu override dari environment variables.
- `EnvExists() bool` — cek apakah .env sudah ada (untuk deteksi first run).
- Struktur Config harus mencakup minimal: app.env, server.host, server.port,
  server.domain, server.scheme, database.path, upload.max_size_mb,
  upload.dir, session.secret, session.max_age.
- `Server.Scheme`: kosong = auto-detect dari request (X-Forwarded-Proto/TLS),
  "http"/"https" = force.
- `Addr()` method → "host:port".
- `IsProduction()` → env == "production".

#### C. Update routes (`routes/routes.go`)
- `SetupRouter(templateFS embed.FS, publicFS embed.FS) *gin.Engine`
- Templates: parse semua .html dari embed.FS menggunakan `filepath.Base()` sebagai
  nama template (agar cocok dengan `c.HTML(200, "nama.html", data)`). Set FuncMap
  SEBELUM parse. Gunakan `r.SetHTMLTemplate(tmpl)`.
- Static assets: `fs.Sub(publicFS, "public")` lalu `r.StaticFS("/public", http.FS(sub))`.

#### D. Setup wizard (`setup/wizard.go`)
- `RunWizard(yamlData []byte) *config.Config`
- Flow:
  1. Cetak banner
  2. Tanya interaktif: port, domain, scheme, max upload, auto-generate session
     secret (openssl rand atau crypto/rand)
  3. Tulis file .env (chmod 0600)
  4. Buat folder data/, uploads/, dan logs/
  5. Jika Linux + root: tawarkan systemd, nginx, SSL (jalankan perintah via
      exec.Command dengan sudo)
  6. Systemd setup otomatis:
     - chown -R <service-user>:<service-user> untuk folder data/, uploads/, logs/
       (agar service yang jalan sebagai www-data bisa menulis)
     - copy service file, daemon-reload, enable, restart (auto-start)
  7. Cetak ringkasan + info akses (URL, login, lokasi .env & log) + jalankan server
- Helper functions: askInput (dengan default), askYesNo (dengan default Y/N),
  generateSecret, isRoot, runCmd (exec.Command dengan stdin/stdout/stderr).
- Systemd: generate service file dinamis (path dari os.Executable() + os.Getwd()),
  sertakan ProtectSystem=strict, ProtectHome=true, ReadWritePaths=data/uploads/logs,
  copy ke /etc/systemd/system/, daemon-reload, enable, restart.
- Nginx: generate config dengan client_max_body_size, proxy_pass, X-Forwarded-Proto,
  copy ke sites-available, symlink ke sites-enabled, nginx -t, reload.
- SSL: certbot --nginx -d domain --redirect.

#### E. Update `main.go`
```go
func main() {
    var cfg *config.Config
    firstRun := !config.EnvExists()
    if firstRun {
        cfg = setup.RunWizard(configYAML)
    } else {
        cfg = config.Load(configYAML)
    }
    // Init logging (slog → stdout + file logs/app.log) SEBELUM apa pun
    if err := config.InitLogging(cfg.Log.Dir); err != nil {
        log.Fatalf("Gagal init logging: %v", err)
    }
    if firstRun {
        printPostSetupInfo(cfg) // cetak URL akses, login, lokasi .env & log
    }
    config.ConnectDatabase()
    r := routes.SetupRouter(templateFS, publicFS)
    // Pesan foreground + petunjuk systemd
    r.Run(cfg.Addr())
}
```

#### E2. Logging (`config/logging.go`)
- `InitLogging(logDir string) error` — buat folder `logs/`, buka `app.log`
  (mode 0640), pakai `slog.NewJSONHandler(io.MultiWriter(os.Stdout, file), ...)`.
- `slog.SetDefault(Logger)` agar semua `slog.Info/Error/Warn` tulis ke stdout + file.
- Jangan log password/secret ke file.
- Pada error fatal: gunakan `slog.Error` dulu (masuk file) lalu `log.Fatal` (stderr).

#### F. Script `build.sh` (root proyek)
```bash
#!/bin/bash
set -e
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ebupot-app .
# tampilkan ukuran + instruksi upload
```

#### G. Script `deploy/deploy.ps1` (opsional, untuk Windows)
- Cross-compile, scp upload single binary ke VPS, tampilkan instruksi wizard.

### Catatan teknis:
- Gunakan driver SQLite pure-Go (github.com/glebarez/sqlite) agar CGO_ENABLED=0
  (tidak butuh C compiler di VPS).
- Template names harus konsisten: controllers panggil `c.HTML(200, "basename.html", data)`,
  dan embed parser gunakan `filepath.Base()` sebagai nama.
- Cross-template references ({{template "header.html" .}}) harus pakai basename.
- .env TIDAK di-embed (di-generate oleh wizard, di-gitignore).
- config.yaml di-embed sebagai default, .env override.
- Jika user ingin reconfigure: hapus .env, jalankan ulang binary → wizard muncul lagi.
- Logging via `log/slog` (built-in Go 1.21+) ke stdout + `logs/app.log`. Jangan
  gunakan `log.Printf` untuk log penting karena tidak masuk file.
- Systemd service WAJIB sertakan folder `logs/` di `ReadWritePaths` selain
  `data/` dan `uploads/` (karena `ProtectSystem=strict` membuat FS read-only).
- Setelah buat folder, wizard otomatis `chown -R <service-user>` agar service
  yang jalan sebagai www-data bisa menulis — tanpa ini, service crash dengan
  "permission denied" saat init logging / akses DB.
- Validasi upload PDF: SELALU cek magic bytes (`%PDF`) terlepas dari Content-Type
  header. Handle error pada `file.Open()` dan `Read()`, pastikan `n == 4`.
- Saat edit dokumen dengan re-upload file: simpan DB dulu, hapus file lama
  setelahnya; jangan regenerate UUID/share-link agar link lama tetap valid.
```

---

## Checklist Implementasi

Gunakan checklist ini untuk memverifikasi AI agent sudah mengimplementasi dengan benar:

- [ ] `embed.go` ada di root dengan directive `//go:embed` untuk templates, public, config.yaml
- [ ] `config.Load()` menerima `[]byte` (embedded YAML), bukan baca file dari disk
- [ ] `config.EnvExists()` tersedia dan dipakai di main.go untuk deteksi first run
- [ ] `config.InitLogging()` ada: slog → stdout + `logs/app.log` (mode 0640)
- [ ] `routes.SetupRouter()` menerima `embed.FS` untuk templates & public
- [ ] Templates di-parse dari embed.FS dengan `filepath.Base()` sebagai nama
- [ ] Static assets diserve dari embed.FS via `http.FS(fs.Sub(...))`
- [ ] `setup/wizard.go` menanyakan: port, domain, scheme, upload size, session secret
- [ ] Wizard generate .env, buat folder data/uploads/logs, optional systemd/nginx/SSL
- [ ] Wizard print pesan [OK]/[GAGAL] per langkah
- [ ] Wizard otomatis `chown -R <service-user>` untuk data/uploads/logs
- [ ] Systemd service file sertakan `ReadWritePaths=.../data .../uploads .../logs`
- [ ] Systemd setup auto-restart service (bukan hanya enable)
- [ ] `main.go` cek `EnvExists()` → wizard jika false, normal run jika true
- [ ] `main.go` cetak info post-setup (URL, login, lokasi .env & log) saat first run
- [ ] `build.sh` cross-compile linux/amd64 dengan CGO_ENABLED=0
- [ ] Binary berjalan tanpa file eksternal (coba pindah ke folder kosong lalu jalankan)
- [ ] Hapus .env → wizard muncul lagi
- [ ] `.env` ada di `.gitignore`
- [ ] `logs/` ada di `.gitignore` (dengan `.gitkeep`)
- [ ] Password/secret TIDAK di-log ke file
- [ ] Validasi upload PDF cek magic bytes terlepas dari Content-Type
- [ ] `go build` dan `go vet ./...` lolos tanpa error

---

## Struktur File yang Dihasilkan

```
proyek-anda/
├── embed.go              # Directive //go:embed
├── main.go               # Entry point + wizard flow + post-setup info
├── build.sh              # Cross-compile script
├── config.yaml           # Default config (di-embed)
├── .env.example          # Template env (untuk referensi)
├── .env                  # Generated oleh wizard (gitignore)
├── .gitignore
├── config/
│   ├── config.go         # Load embedded YAML + .env override + LOG_DIR
│   ├── database.go       # Koneksi DB dari config (log via slog)
│   └── logging.go        # InitLogging: slog → stdout + logs/app.log
├── setup/
│   └── wizard.go         # Setup wizard + auto-chown + systemd restart
├── routes/
│   └── routes.go         # SetupRouter(embed.FS, embed.FS)
├── controllers/
├── models/
├── middlewares/
├── templates/            # Di-embed ke binary
├── public/               # Di-embed ke binary
├── data/                 # Dibuat saat runtime (gitignore)
├── uploads/              # Dibuat saat runtime (gitignore)
├── logs/                 # Dibuat saat runtime (gitignore, +.gitkeep)
└── deploy/
    ├── deploy.ps1        # Upload otomatis (opsional)
    ├── ebupot.service    # Template referensi systemd
    └── nginx.conf        # Template referensi nginx
```

---

## Contoh Output Wizard

```
+======================================================+
|     [Nama Aplikasi] — Setup Wizard                   |
+======================================================+

File konfigurasi (.env) belum ditemukan.
Mari kita setup aplikasi ini.

[1] Port aplikasi [8080]: 8080
[2] Domain publik [localhost:8080]: app.domainanda.com
[3] Scheme (http/https/kosong=auto-detect): 
[4] Maksimal upload (MB) [100]: 
[5] Auto-generate session secret? [Y/n]: y

  [OK] File .env dibuat
  [OK] Folder data/, uploads/, logs/ dibuat

Setup systemd service? [Y/n]: y
  [OK] Systemd service terpasang & dijalankan (auto-start saat boot)
       Status: sudo systemctl status ebupot

Setup Nginx reverse proxy? [Y/n]: y
  [OK] Nginx terpasang untuk app.domainanda.com

Setup SSL (Let's Encrypt)? [y/N]: y
  [OK] SSL terpasang

+======================================================+
|              SETUP SELESAI!                          |
+======================================================+

+======================================================+
|              INFO APLIKASI                           |
+======================================================+
|  Akses      : https://app.domainanda.com
|  Login      : admin / admin123
|  .env       : /opt/app/.env
|  Log file   : /opt/app/logs/app.log
+======================================================+

Perintah berguna:
  Mulai service : sudo systemctl start ebupot
  Stop service  : sudo systemctl stop ebupot
  Status service: sudo systemctl status ebupot
  Lihat log     : sudo journalctl -u ebupot -f
                 tail -f logs/app.log

>>> Memulai server...

{"level":"INFO","msg":"Aplikasi started","app":"...","env":"production","listen":"0.0.0.0:8080"}
{"level":"INFO","msg":"Database Connected & Migrated!"}
{"level":"INFO","msg":"Server berjalan","url":"http://0.0.0.0:8080","domain_qr":"app.domainanda.com"}

>>> Server berjalan di foreground. Tekan Ctrl+C untuk berhenti.
>>> Untuk background: sudo systemctl start ebupot
```

---

## Tips untuk AI Agent

1. **Selalu test dengan menghapus .env** — wizard harus muncul kembali.
2. **Test binary di folder kosong** — tidak boleh ada dependency file eksternal
   (kecuali .env yang dibuat wizard dan folder data/uploads/logs yang dibuat wizard).
3. **CGO_ENABLED=0 wajib** — gunakan driver DB pure-Go (glebarez/sqlite untuk
   SQLite, atau driver lain yang tidak butuh CGO).
4. **Template basename** — pastikan semua `c.HTML()` dan `{{template}}`
   references menggunakan basename file, bukan path lengkap.
5. **FuncMap sebelum Parse** — set FuncMap pada `template.New("")` sebelum
   mem-parse template dari embed.FS.
6. **Permission .env = 0600** — file berisi secret, tidak boleh world-readable.
7. **Wizard skip di Windows** — systemd/nginx/SSL hanya untuk Linux. Deteksi via
   `runtime.GOOS == "linux"`.
8. **Logging ke stdout + file** — gunakan `io.MultiWriter`. Saat jalan manual,
   output tetap muncul di terminal; saat jalan via systemd, masuk journal + file.
9. **ReadWritePaths systemd** — wajib sertakan `data`, `uploads`, DAN `logs`
   karena `ProtectSystem=strict` membuat seluruh FS read-only.
10. **Auto-chown folder** — setelah buat folder, `chown -R <service-user>` agar
    service jalan sebagai www-data bisa menulis. Tanpa ini, service crash dengan
    "permission denied" saat init logging / akses DB.
11. **Auto-restart service** — wizard harus `systemctl restart` (bukan hanya
    enable) agar service langsung jalan setelah setup.
12. **Jangan log password** — default credential cukup dicetak di banner post-setup,
    jangan ditulis ke `logs/app.log`.
13. **Validasi upload aman** — cek magic bytes terlepas dari Content-Type; handle
    error `Open()`/`Read()`; pastikan `n == 4` sebelum validasi.
14. **Edit dengan re-upload** — simpan DB dulu, hapus file lama setelahnya; jangan
    regenerate UUID/share-link agar link lama tetap valid.
