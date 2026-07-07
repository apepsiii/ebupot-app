Setelah dokumen BRD dan PRD selesai dan disepakati, Anda sudah memiliki fondasi yang sangat kuat. Tahap selanjutnya adalah bergeser dari fase perencanaan (planning) ke fase **Pengembangan (Development)**.

Berikut adalah peta jalan (roadmap) langkah-langkah yang perlu Anda lakukan selanjutnya untuk mulai membangun aplikasi ini:

### 1. Desain UI/UX (Wireframing & Mockup)

Sebelum mulai *coding*, ada baiknya Anda membuat coretan kasar atau desain halaman menggunakan alat seperti Figma, Balsamiq, atau sekadar di atas kertas. Halaman yang perlu didesain:

* Landing Page & Halaman Login.
* Tampilan Dashboard Admin (Tabel User, Tabel e-Bupot, Modal Upload, Modal QR Code).
* Tampilan Dashboard User.
* Halaman "Error/404" jika tautan QR Code kedaluwarsa.

### 2. Inisialisasi Proyek (Environment Setup)

Mulai menyiapkan lingkungan kerja (*workspace*) di komputer Anda:

* Install **Golang** (jika belum).
* Buat folder proyek baru dan jalankan `go mod init nama-proyek-anda` di terminal.
* Buat struktur folder yang sudah kita bahas di PRD (contoh: folder `controllers/`, `models/`, `templates/`, `uploads/`).
* Unduh *libraries* (dependencies) utama menggunakan perintah `go get`, seperti:
* `go get -u github.com/gin-gonic/gin`
* `go get -u gorm.io/gorm`
* `go get -u gorm.io/driver/sqlite`
* `go get -u github.com/google/uuid`
* `go get -u github.com/skip2/go-qrcode`



### 3. Pembuatan Model & Koneksi Database (GORM)

Langkah *coding* pertama biasanya dimulai dari data. Anda perlu menulis kode Go untuk melakukan koneksi ke SQLite dan mendefinisikan *Struct* (Model) untuk tabel `User` dan `Ebupot`. GORM akan menggunakan *Struct* ini untuk membuat tabel secara otomatis di *database* Anda (*Auto Migrate*).

### 4. Pengembangan Back-End (Controller & Routing)

Setelah *database* siap, Anda mulai membangun mesin utamanya menggunakan Gin Framework:

* Membuat sistem register (oleh admin) dan *login* (termasuk *hashing password*).
* Membangun fungsi CRUD untuk User dan e-Bupot.
* Membangun fungsi *Upload* PDF dan men-generate UUID.
* Membangun fungsi Generator QR Code dari UUID tersebut.
* Membangun *endpoint* khusus unduhan: `/documentmanagementportal/api/DocumentExternalLink/:uuid`.

### 5. Integrasi Front-End (HTML Templates)

Menyambungkan logika *Back-End* yang sudah dibuat ke antarmuka HTML. Anda akan menggunakan fitur *HTML Rendering* dari Gin untuk melempar data (misalnya data daftar dokumen) ke halaman web agar bisa dilihat oleh Admin dan User.

### 6. Testing (Pengujian)

Menguji aplikasi secara menyeluruh secara lokal (di komputer Anda):

* Apakah *User A* bisa melihat data *User B*? (Seharusnya tidak).
* Apakah format URL di QR Code sudah benar?
* Apakah saat QR Code di-*scan*, *file* PDF otomatis terunduh dan tidak rusak?

### 7. Deployment (Rilis ke Server/VPS)

Karena Anda membutuhkan domain khusus yang menyerupai *path* DJP Coretax, aplikasi ini tidak bisa sekadar di-hosting di *shared hosting* biasa. Anda perlu menyewa **VPS (Virtual Private Server)**, mengarahkan domain Anda ke IP VPS tersebut, dan menjalankan aplikasi Go Anda di sana agar bisa diakses oleh publik (Klien/Pemindai QR).