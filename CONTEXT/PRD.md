Berikut adalah **Product Requirements Document (PRD)** versi terbaru dan terlengkap yang telah disesuaikan dengan semua revisi Anda, khususnya mengenai struktur *prefix* URL untuk sistem pengunduhan via QR Code.

---

# PRODUCT REQUIREMENTS DOCUMENT (PRD)

**Nama Produk:** Aplikasi Pengelola e-Bupot & QR Generator
**Platform:** Web Application
**Tumpukan Teknologi (Tech Stack):** Golang (Gin Framework), HTML/CSS (Bootstrap/Tailwind), SQLite

## 1. Tujuan Produk

Membangun portal manajemen dokumen e-Bupot yang memfasilitasi Administrator dalam mengelola dan mendistribusikan bukti potong pajak secara aman dan terstruktur kepada Wajib Pajak (User). Sistem ini dirancang untuk memberikan pengalaman mengunduh dokumen instan melalui pemindaian QR Code. Tautan di dalam QR Code secara khusus meniru *path* dari DJP Coretax, yaitu `https://[DOMAIN_SAYA]/documentmanagementportal/api/DocumentExternalLink/{UUID}`, untuk memberikan standar URL yang familiar namun tetap berada di bawah kendali *server* internal (domain Anda sendiri).

## 2. Persona Pengguna

1. **Admin:** Pengelola sistem yang memiliki wewenang penuh untuk membuat akun klien, mengunggah *file* fisik PDF, menetapkan periode pajak, dan menerbitkan QR Code.
2. **User (Klien / Wajib Pajak):** Pemilik dokumen e-Bupot yang hanya memiliki wewenang untuk masuk ke akun mereka sendiri, melihat daftar dokumen yang ditujukan untuk mereka, dan mengunduhnya.
3. **Guest (Scanner):** Pihak eksternal (atau Klien itu sendiri) yang menggunakan kamera HP/aplikasi *scanner* untuk memindai QR Code dan langsung mendapatkan *file* PDF tanpa harus *login*.

## 3. User Stories (Skenario Pengguna)

**Sebagai Admin:**

* *Saya ingin* masuk ke *dashboard* melalui halaman *login* yang aman *sehingga* hanya saya yang bisa mengelola *database* e-Bupot.
* *Saya ingin* melakukan CRUD (Create, Read, Update, Delete) pada data User *sehingga* klien saya memiliki kredensial untuk masuk ke portal.
* *Saya ingin* mengunggah dokumen PDF, memilih *User* pemiliknya, dan mengatur bulan & tahun *sehingga* dokumen didistribusikan dengan tepat sasaran.
* *Saya ingin* sistem otomatis men-generate UUID untuk setiap dokumen yang diunggah *sehingga* kerahasiaan tautan tetap terjaga.
* *Saya ingin* melihat *pop-up* QR Code yang memuat tautan dengan format *path* khusus (`/documentmanagementportal/api/DocumentExternalLink/...`) *sehingga* saya bisa menyimpannya dan membagikannya ke klien.

**Sebagai User (Klien):**

* *Saya ingin* *login* menggunakan akun yang didaftarkan Admin *sehingga* saya bisa mengakses *dashboard* pribadi saya.
* *Saya ingin* memfilter daftar e-Bupot berdasarkan bulan dan tahun *sehingga* saya dapat mencari bukti potong untuk periode spesifik.
* *Saya ingin* menekan tombol "Download" di tabel *dashboard* *sehingga* *file* e-Bupot langsung tersimpan di perangkat saya.

**Sebagai Pemindai (Scanner):**

* *Saya ingin* memindai QR Code menggunakan HP *sehingga* *browser* saya langsung mengunduh *file* PDF e-Bupot tersebut secara otomatis tanpa rintangan *login*.

## 4. Spesifikasi Fungsional & UI/UX

### 4.1. Halaman Publik (Front-End)

* **Landing Page (`/`):** Halaman statis satu halaman (*single page*) yang berisi nama produk, fungsi aplikasi, dan tombol utama "Login Portal".
* **Auth Page (`/login`):** *Form login* dengan kolom `Username` dan `Password`. Pesan validasi akan muncul jika kredensial salah atau *session* telah kedaluwarsa.

### 4.2. Dashboard Admin

* **Layout:** Menggunakan struktur *Sidebar* (Menu: Dashboard, Manajemen User, Manajemen e-Bupot, Logout) dan *Main Content Area* (untuk tabel dan form).
* **Manajemen User (CRUD):**
* Tabel menampilkan: `No`, `Nama`, `Username`, `Aksi` (Edit, Hapus).
* Modal/Form Tambah: Input `Nama Lengkap`, `Username`, `Password`. Password wajib di-hash sebelum disimpan.


* **Manajemen e-Bupot (CRUD):**
* Tabel menampilkan: `Bulan`, `Tahun`, `Nama User`, `Nama File`, `Aksi` (Generate QR, Edit, Hapus).
* Modal/Form Upload:
* Dropdown *Select User*.
* Dropdown Bulan (1-12) & Input Tahun.
* Input File (Validasi: Harus berekstensi `.pdf`).




* **QR Generator Modal:**
* Tombol "Generate QR" pada tabel memicu munculnya gambar QR Code.
* **Logika QR:** Sistem menyandikan URL lengkap `https://[DOMAIN_SAYA]/documentmanagementportal/api/DocumentExternalLink/{UUID}` ke dalam *barcode*.
* Terdapat tombol "Unduh QR" untuk menyimpan gambar tersebut dalam format `.png`.



### 4.3. Dashboard User

* **Layout:** Antarmuka sederhana yang terisolasi. *Header* menyapa nama *User*.
* **Manajemen Tampilan Dokumen:**
* Tabel hanya menampilkan data dokumen di mana `user_id` sama dengan ID *user* yang sedang *login*.
* Terdapat Filter *Dropdown* Bulan dan Tahun untuk menyortir data tabel.
* Terdapat tombol "Download File" di setiap baris.



### 4.4. Mesin Pengunduhan Otomatis (Direct Download Handler)

* **Endpoint Target:** `GET /documentmanagementportal/api/DocumentExternalLink/:uuid`
* **Logika Sistem:** Gin Framework akan menangkap rute ini, mengekstrak UUID dari parameter URL, mencari lokasi fisik *file* di *database* SQLite, dan mengirimkannya kembali ke *browser* pemindai.
* **Header HTTP Wajib:** Harus menyertakan instruksi agar *browser* mengunduh *file*:
* `Content-Type: application/pdf`
* `Content-Disposition: attachment; filename="{nama_file}.pdf"`



## 5. Kriteria Penerimaan (Acceptance Criteria)

Sistem siap di-*deploy* jika kondisi berikut terpenuhi:

1. **Keamanan Sesi:** Jika *User* atau *Guest* mencoba mengakses rute `/admin/...`, sistem wajib memblokir dan mengembalikannya ke halaman `/login`.
2. **Isolasi Data Klien:** *User A* tidak dapat melihat, memanipulasi, atau menemukan daftar e-Bupot milik *User B* melalui *dashboard*.
3. **Akurasi URL QR Code:** Tautan yang di-*scan* dari QR Code harus benar-benar cocok dengan format yang diminta: `https://[DOMAIN]/documentmanagementportal/api/DocumentExternalLink/{UUID}`.
4. **Keberhasilan Unduhan Eksternal:** Seseorang yang memindai QR Code dari HP akan langsung mendapatkan *prompt* *download* atau otomatis mengunduh *file* PDF tanpa error *corrupted file*.
5. **Integritas Hapus Data:** Ketika Admin menghapus e-Bupot melalui antarmuka, *file* `.pdf` fisik di dalam server (folder `uploads/`) harus ikut terhapus agar server tidak penuh oleh *file* yatim (*orphan files*).

## 6. Kebutuhan Teknis (Technical Constraints)

* **Penyimpanan (*Storage*):** *File* PDF disimpan di direktori server lokal (misalnya di root folder `./uploads/ebupots/`). Folder ini **wajib** diproteksi dan tidak boleh di-set sebagai *static folder* publik di Gin (misal menggunakan `router.StaticFS`) agar orang tidak bisa menebak direktori *file*.
* **Database:** Menggunakan SQLite3. Relasi *database* harus dijaga (contoh: jika data User dihapus, maka secara kaskade (*cascade*) data e-Bupot miliknya idealnya ikut terhapus atau dialihkan).
* **Arsitektur MVC:** Kode Go harus dipisah secara rapi (Controllers, Models, Middleware, Routers).

## 7. Penanganan Error (Error Handling)

* **Tautan Tidak Valid (404):** Jika pemindai mengakses URL `/documentmanagementportal/api/DocumentExternalLink/{UUID}` tetapi UUID tersebut salah, kedaluwarsa, atau dokumen telah dihapus Admin, sistem harus menampilkan halaman HTML elegan bertuliskan: *"Dokumen tidak ditemukan atau sudah tidak berlaku."* (Bukan sekadar pesan teks *error* server).
* **Format File Salah:** Saat Admin mencoba *upload* file non-PDF (misal: `.jpg`, `.docx`), sistem menolak dengan pesan *"Hanya file berformat PDF yang diizinkan."*
* **Ukuran File Maksimal:** Gin Router harus membatasi ukuran maksimal *upload file* (misalnya maksimal 5MB) untuk mencegah *spam memory*.