Berikut adalah versi lengkap dan terbaru dari **Business Requirements Document (BRD)** yang telah mengintegrasikan semua detail fitur, alur kerja, spesifikasi teknis, serta revisi format *prefix* URL QR Code yang menyesuaikan dengan jalur (*path*) DJP Coretax, namun tetap berjalan di atas domain Anda sendiri.

---

# BUSINESS REQUIREMENTS DOCUMENT (BRD)

**Nama Proyek:** Aplikasi Manajemen e-Bupot & QR Generator
**Platform:** Web Application
**Tumpukan Teknologi (Tech Stack):** Golang (Gin Framework), HTML/CSS/JS, SQLite

## 1. Ringkasan Eksekutif & Tujuan

Proyek ini bertujuan untuk membangun sebuah sistem manajemen dokumen (e-Bupot) terpusat. Sistem ini memungkinkan Administrator untuk mengelola akun klien, mengunggah dokumen PDF e-Bupot secara spesifik berdasarkan bulan/tahun, dan mendistribusikan bukti potong pajak secara aman. Sistem dilengkapi dengan generator QR Code yang menghasilkan tautan unduhan dengan *prefix* URL khusus (meniru struktur DJP Coretax) pada domain mandiri, sehingga memudahkan Wajib Pajak untuk mengunduh dokumen secara langsung saat dipindai.

## 2. Aktor & Hak Akses (User Roles)

Sistem ini dirancang untuk dua jenis pengguna:

1. **Administrator (Admin):** Memiliki kontrol penuh (CRUD) atas manajemen *user*, pengelolaan dokumen e-Bupot, penentuan periode pajak, dan pembuatan (generate) QR Code.
2. **User (Klien/Wajib Pajak):** Pengguna dengan akses terbatas yang hanya dapat *login* untuk melihat dan mengunduh dokumen e-Bupot milik mereka sendiri berdasarkan filter bulan/tahun.

## 3. Kebutuhan Fungsional (Functional Requirements)

### 3.1. Modul Halaman Publik & Autentikasi

* **REQ-F01:** Sistem harus menyediakan *Landing Page* statis berbasis HTML sebagai halaman informasi utama.
* **REQ-F02:** Sistem harus memiliki halaman *Login* (Auth Page) dengan validasi *Username* dan *Password*.
* **REQ-F03:** Sistem harus mengarahkan pengguna secara otomatis ke *dashboard* yang tepat (Dashboard Admin atau Dashboard User) setelah proses *login* berhasil.

### 3.2. Modul Manajemen Pengguna (Dashboard Admin)

* **REQ-F04:** Admin dapat melihat daftar seluruh *user* yang terdaftar di dalam sistem.
* **REQ-F05:** Admin dapat membuat akun *user* baru dengan atribut minimum: Nama, Username, dan Password (disimpan dalam bentuk *hash*).
* **REQ-F06:** Admin dapat mengedit detail *user* dan menghapus akun *user*.

### 3.3. Modul Manajemen e-Bupot (Dashboard Admin)

* **REQ-F07:** Admin dapat mengunggah *file* dokumen (berformat PDF).
* **REQ-F08:** Saat proses *upload*, Admin diwajibkan untuk memilih **User Pemilik** dokumen dan menentukan **Periode** (Bulan dan Tahun).
* **REQ-F09:** Sistem secara otomatis akan men-generate ID unik (UUID v4) untuk setiap dokumen yang diunggah.
* **REQ-F10:** Admin dapat melihat daftar e-Bupot, memperbarui informasi (edit), dan menghapus data (menghapus *record* di *database* sekaligus *file* fisik di server).

### 3.4. Modul Generator QR Code & Tautan (Dashboard Admin)

* **REQ-F11:** Admin dapat menekan tombol "Generate QR Code" pada setiap baris data e-Bupot.
* **REQ-F12:** Sistem akan menghasilkan gambar QR Code yang menyimpan tautan (*link*) dengan *prefix* khusus yang menggabungkan domain utama aplikasi dengan *path* eksternal Coretax, dengan format:
`https://[DOMAIN_SAYA]/documentmanagementportal/api/DocumentExternalLink/{UUID}`
* **REQ-F13:** Admin dapat mengunduh (*download*) gambar QR Code tersebut.

### 3.5. Modul Distribusi & Unduhan Dokumen

* **REQ-F14:** Jika URL unik dari QR Code dipindai atau diakses melalui *browser*, sistem harus memicu proses *Download File* (memaksa *browser* mengunduh PDF, bukan membukanya di *tab* baru).
* **REQ-F15:** *User* yang masuk ke *dashboard* dapat melihat daftar e-Bupot milik mereka sendiri, menggunakan filter periode (bulan/tahun), dan mengunduh dokumen secara langsung tanpa perlu memindai QR Code.

## 4. Kebutuhan Non-Fungsional (Non-Functional Requirements)

* **Keamanan:** Kredensial *password* dilindungi menggunakan algoritma *hashing* (contoh: bcrypt). Manajemen akses *(Role-Based Access)* diatur menggunakan *Session* atau JWT yang divalidasi via *Middleware* di Gin.
* **Isolasi Data:** Kueri *database* pada level *User* harus disaring (di-filter) secara ketat berdasarkan `user_id` yang sedang *login*.
* **Penyimpanan:** *File* PDF disimpan di dalam penyimpanan lokal (folder khusus, misal `uploads/`), dan tidak diekspos sebagai direktori publik. Pengunduhan wajib melewati *Controller* aplikasi.
* **Antarmuka (UI):** Menggunakan HTML dan CSS *Framework* untuk memastikan tata letak responsif, baik diakses melalui desktop maupun ponsel.

## 5. Rancangan Struktur Database (SQLite)

### Tabel `users`

| Kolom | Tipe Data | Keterangan |
| --- | --- | --- |
| `id` | INTEGER | Primary Key, Auto Increment |
| `username` | TEXT | Unique, kredensial login |
| `password` | TEXT | Hash password |
| `name` | TEXT | Nama lengkap klien/wajib pajak |
| `role` | TEXT | 'admin' atau 'user' |
| `created_at` | DATETIME | Waktu pendaftaran akun |

### Tabel `ebupots`

| Kolom | Tipe Data | Keterangan |
| --- | --- | --- |
| `id` | INTEGER | Primary Key, Auto Increment |
| `user_id` | INTEGER | Foreign Key ke tabel `users` |
| `bulan` | INTEGER | Bulan berlakunya dokumen (1-12) |
| `tahun` | INTEGER | Tahun berlakunya dokumen (contoh: 2026) |
| `file_path` | TEXT | Lokasi fisik file PDF di server |
| `uuid_link` | TEXT | UUID v4 (cth: `3a57cbff-ba46...`) |
| `created_at` | DATETIME | Waktu dokumen diunggah |

## 6. Alur Kerja Sistem (System Flow)

1. **Pembuatan Data Base:** Admin *login* ke portal, menambahkan akun untuk "User A".
2. **Manajemen Dokumen:** Admin menuju menu e-Bupot, mengunggah *file* PDF bukti potong pajak, memilih nama "User A", dan mengatur periode bulan/tahun.
3. **Pembuatan UUID & Tautan:** Sistem menyimpan PDF, menghasilkan UUID `3a57cbff...`, dan mengaitkannya dengan struktur URL yang telah disepakati.
4. **Distribusi QR Code:** Admin men-generate QR Code untuk e-Bupot tersebut dan membagikannya ke Klien.
5. **Pengunduhan via Scan:** Klien/pihak terkait memindai QR Code. Gin *router* menangkap *path* `/documentmanagementportal/api/DocumentExternalLink/{UUID}`, mencocokkan UUID di SQLite, menemukan *file path*, lalu mengirim respons HTTP untuk mengunduh PDF.
6. **Pengunduhan via Dashboard:** Klien *login* ke akunnya sendiri, melihat tabel daftar dokumen khusus miliknya, dan mengklik tombol unduh langsung.

## 7. Rancangan Endpoint / Routing (Gin Framework)

Struktur perutean diatur agar sesuai dengan *prefix* URL yang diinginkan:

**Rute Publik & Mesin Unduh (Tanpa Middleware Auth)**

* `GET /` : Menampilkan Landing Page statis.
* `GET /login` : Menampilkan halaman form Login.
* `POST /login` : Memproses validasi Login pengguna.
* `GET /documentmanagementportal/api/DocumentExternalLink/:uuid` : Menangani proses *download file* langsung melalui *link* panjang yang disematkan dalam QR Code.

**Rute Dasbor Admin (Dilindungi Middleware 'Admin')**

* `GET /admin/dashboard` : Halaman utama admin.
* `GET /admin/users` : Menampilkan manajemen pengguna (CRUD).
* `POST /admin/users` : Menyimpan data pengguna baru/edit/hapus.
* `GET /admin/ebupots` : Menampilkan daftar e-Bupot dan form *upload*.
* `POST /admin/ebupots` : Menangani *upload file* PDF dan penyimpanan ke *database*.
* `POST /admin/ebupots/delete/:id` : Menghapus data dan *file* fisik.
* `GET /admin/ebupots/qr/:uuid` : Mengembalikan *response* gambar QR Code.

**Rute Dasbor User (Dilindungi Middleware 'User')**

* `GET /user/dashboard` : Menampilkan daftar e-Bupot terfilter berdasarkan *user_id* yang masuk, beserta opsi filter bulan.

## 8. Struktur Folder Proyek (Rekomendasi Go)

```text
/ebupot-app
├── /controllers       # Logika fungsi (AuthController, AdminController, UserController, DownloadController)
├── /database          # Konfigurasi koneksi SQLite (gorm)
├── /middleware        # Pengecekan autentikasi session
├── /models            # Definisi Struct Go (User, Ebupot)
├── /public            # Aset statis (CSS, JS, Images untuk Landing Page)
├── /templates         # File HTML terpisah (Layout, Login, Admin Panel, User Panel)
├── /uploads           # Folder penyimpanan file fisik PDF (Terproteksi/Bukan Public)
├── main.go            # File utama aplikasi (Inisialisasi Gin dan Routing)
└── go.mod / go.sum    # Manajemen dependency Go

```

## 9. Kebutuhan *Libraries* Utama (Go Dependencies)

* **`[github.com/gin-gonic/gin](https://github.com/gin-gonic/gin)`** : Framework HTTP web utama.
* **`gorm.io/gorm` & `gorm.io/driver/sqlite**` : ORM untuk mengelola operasi CRUD ke *database* SQLite.
* **`golang.org/x/crypto/bcrypt`** : Mengamankan *password user*.
* **`[github.com/google/uuid](https://github.com/google/uuid)`** : Menghasilkan UUID (seperti `3a57cbff-ba46-47f0-8400-5b7be5c78aba`).
* **`[github.com/skip2/go-qrcode](https://github.com/skip2/go-qrcode)`** : Men-generate gambar QR Code secara dinamis dari URL panjang.
* **`[github.com/gin-contrib/sessions](https://github.com/gin-contrib/sessions)`** : Mengelola sesi *login* agar *dashboard* aman.