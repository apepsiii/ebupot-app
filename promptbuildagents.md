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
  4. Buat folder data/ dan uploads/
  5. Jika Linux + root: tawarkan systemd, nginx, SSL (jalankan perintah via
     exec.Command dengan sudo)
  6. Cetak ringkasan + jalankan server
- Helper functions: askInput (dengan default), askYesNo (dengan default Y/N),
  generateSecret, isRoot, runCmd (exec.Command dengan stdin/stdout/stderr).
- Systemd: generate service file dinamis (path dari os.Executable()), copy ke
  /etc/systemd/system/, daemon-reload, enable.
- Nginx: generate config dengan client_max_body_size, proxy_pass, X-Forwarded-Proto,
  copy ke sites-available, symlink ke sites-enabled, nginx -t, reload.
- SSL: certbot --nginx -d domain --redirect.

#### E. Update `main.go`
```go
func main() {
    var cfg *config.Config
    if !config.EnvExists() {
        cfg = setup.RunWizard(configYAML)
    } else {
        cfg = config.Load(configYAML)
    }
    // init database, setup router, run server
    config.ConnectDatabase()
    r := routes.SetupRouter(templateFS, publicFS)
    r.Run(cfg.Addr())
}
```

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
```

---

## Checklist Implementasi

Gunakan checklist ini untuk memverifikasi AI agent sudah mengimplementasi dengan benar:

- [ ] `embed.go` ada di root dengan directive `//go:embed` untuk templates, public, config.yaml
- [ ] `config.Load()` menerima `[]byte` (embedded YAML), bukan baca file dari disk
- [ ] `config.EnvExists()` tersedia dan dipakai di main.go untuk deteksi first run
- [ ] `routes.SetupRouter()` menerima `embed.FS` untuk templates & public
- [ ] Templates di-parse dari embed.FS dengan `filepath.Base()` sebagai nama
- [ ] Static assets diserve dari embed.FS via `http.FS(fs.Sub(...))`
- [ ] `setup/wizard.go` menanyakan: port, domain, scheme, upload size, session secret
- [ ] Wizard generate .env, buat folder, optional systemd/nginx/SSL
- [ ] Wizard print pesan [OK]/[GAGAL] per langkah
- [ ] `main.go` cek `EnvExists()` → wizard jika false, normal run jika true
- [ ] `build.sh` cross-compile linux/amd64 dengan CGO_ENABLED=0
- [ ] Binary berjalan tanpa file eksternal (coba pindah ke folder kosong lalu jalankan)
- [ ] Hapus .env → wizard muncul lagi
- [ ] `.env` ada di `.gitignore`
- [ ] `go build` dan `go vet ./...` lolos tanpa error

---

## Struktur File yang Dihasilkan

```
proyek-anda/
├── embed.go              # Directive //go:embed
├── main.go               # Entry point + wizard flow
├── build.sh              # Cross-compile script
├── config.yaml           # Default config (di-embed)
├── .env.example          # Template env (untuk referensi)
├── .env                  # Generated oleh wizard (gitignore)
├── .gitignore
├── config/
│   ├── config.go         # Load embedded YAML + .env override
│   └── database.go       # Koneksi DB dari config
├── setup/
│   └── wizard.go         # Setup wizard interaktif
├── routes/
│   └── routes.go         # SetupRouter(embed.FS, embed.FS)
├── controllers/
├── models/
├── middlewares/
├── templates/            # Di-embed ke binary
├── public/               # Di-embed ke binary
├── data/                 # Dibuat saat runtime (gitignore)
├── uploads/              # Dibuat saat runtime (gitignore)
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
  [OK] Folder data/ dibuat
  [OK] Folder uploads/ dibuat

Setup systemd service? [Y/n]: y
  [OK] Systemd service terpasang

Setup Nginx reverse proxy? [Y/n]: y
  [OK] Nginx terpasang untuk app.domainanda.com

Setup SSL (Let's Encrypt)? [y/N]: y
  [OK] SSL terpasang

+======================================================+
|              SETUP SELESAI!                          |
+======================================================+

Server berjalan di http://localhost:8080
```

---

## Tips untuk AI Agent

1. **Selalu test dengan menghapus .env** — wizard harus muncul kembali.
2. **Test binary di folder kosong** — tidak boleh ada dependency file eksternal
   (kecuali .env yang dibuat wizard dan folder data/uploads yang dibuat wizard).
3. **CGO_ENABLED=0 wajib** — gunakan driver DB pure-Go (glebarez/sqlite untuk
   SQLite, atau driver lain yang tidak butuh CGO).
4. **Template basename** — pastikan semua `c.HTML()` dan `{{template}}`
   references menggunakan basename file, bukan path lengkap.
5. **FuncMap sebelum Parse** — set FuncMap pada `template.New("")` sebelum
   mem-parse template dari embed.FS.
6. **Permission .env = 0600** — file berisi secret, tidak boleh world-readable.
7. **Wizard skip di Windows** — systemd/nginx/SSL hanya untuk Linux. Deteksi via
   `runtime.GOOS == "linux"`.
