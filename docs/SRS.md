# Software Requirements Specification (SRS)

## ZY Library Manager API

**Versi:** 1.0\
**Status:** Draft untuk implementasi MVP\
**Jenis sistem:** RESTful backend API\
**Nama proyek:** `zy-library-manager-api`

------------------------------------------------------------------------

## 1. Pendahuluan

### 1.1 Tujuan Dokumen

Dokumen ini menetapkan kebutuhan fungsional dan nonfungsional untuk ZY
Library Manager API. Dokumen menjadi acuan implementasi, pengujian, dan
evaluasi agar pengembangan dapat dilakukan tanpa perlu merancang ulang
kebutuhan utama selama proses coding.

### 1.2 Tujuan Produk

ZY Library Manager API adalah layanan backend untuk membantu pustakawan
mengelola katalog buku, data anggota, serta proses peminjaman dan
pengembalian buku. API menyediakan antarmuka HTTP berbasis JSON dan
menyimpan data secara persisten di PostgreSQL.

### 1.3 Ruang Lingkup

Versi MVP mencakup: - Pengelolaan data buku. - Pengelolaan data anggota
perpustakaan. - Pencatatan peminjaman buku. - Pencatatan pengembalian
buku. - Pencarian, filter sederhana, dan pagination untuk daftar buku. -
Validasi input dan penanganan error yang konsisten.

Versi MVP **tidak** mencakup: - Frontend atau aplikasi mobile. - Login,
autentikasi, dan role-based authorization. - Denda keterlambatan,
reservasi, perpanjangan pinjaman, atau notifikasi. - Multi-cabang
perpustakaan. - Barcode/RFID, unggah gambar, dan rekomendasi buku. -
Deployment cloud dan observability tingkat lanjut.

### 1.4 Istilah

  | Istilah | Definisi |
  | ------- | -------- |
  | API | Antarmuka HTTP yang digunakan client untuk berinteraksi dengan sistem |
  | Book | Data judul buku dalam katalog, bukan satu eksemplar fisik individual |
  | Member | Orang yang terdaftar sebagai anggota perpustakaan |
  | Loan | Catatan peminjaman satu judul buku oleh satu anggota |
  | Available copies | Jumlah eksemplar yang saat ini dapat dipinjam |
  | MVP | Versi minimum yang memenuhi kebutuhan inti |

### 1.5 Asumsi dan Keputusan MVP

1.  API digunakan oleh satu operator tepercaya, yaitu pustakawan. Karena
    autentikasi belum termasuk MVP, API hanya ditujukan untuk lingkungan
    pengembangan/lokal dan **tidak boleh dipublikasikan ke internet
    tanpa perlindungan akses**.
2.  Satu record `Book` mewakili satu judul/edisi dengan jumlah eksemplar
    yang tersedia, bukan setiap eksemplar individual.
3.  Setiap transaksi `Loan` meminjam satu unit dari satu judul buku.
    Untuk meminjam beberapa unit, client membuat beberapa transaksi.
4.  Setiap peminjaman memiliki tanggal pinjam dan tanggal jatuh tempo.
5.  Durasi pinjaman default adalah 14 hari kalender dan dihitung oleh
    server.
6.  Buku dan anggota yang pernah memiliki transaksi tidak boleh dihapus
    secara permanen. Penghapusan hanya diperbolehkan jika belum pernah
    dipakai dalam transaksi.
7.  Waktu disimpan di database dengan tipe yang sesuai dan API mengirim
    waktu dalam format ISO 8601/RFC 3339.
8.  Semua endpoint menggunakan prefix `/api/v1`.

------------------------------------------------------------------------

## 2. Gambaran Umum

### 2.1 Kelas Pengguna

  | Pengguna | Kebutuhan |
  | -------- | --------- | 
  | Pustakawan/operator | Mengelola buku, anggota, peminjaman, dan pengembalian |
  | Client API (misalnya Postman atau Mengirim request dan membaca aplikasi lain) | response JSON |

Autentikasi dan pembatasan akses berdasarkan identitas pengguna tidak
termasuk MVP.

### 2.2 Lingkungan Operasi

-   Bahasa: Go.
-   HTTP server: standard library `net/http`.
-   Database: PostgreSQL.
-   Driver PostgreSQL: `pgx` melalui `database/sql`.
-   Format komunikasi: JSON melalui HTTP.
-   Konfigurasi sensitif seperti kredensial database dibaca dari
    environment variable, bukan ditulis langsung di source code.

### 2.3 Batas Sistem

Sistem bertanggung jawab atas validasi, aturan bisnis, konsistensi
transaksi, dan penyimpanan data. Client bertanggung jawab mengirim
request sesuai kontrak API dan menampilkan informasi kepada penggunanya.

------------------------------------------------------------------------

## 3. Kebutuhan Fungsional

### 3.1 Pengelolaan Buku

**FR-BOOK-01 --- Membuat buku**\
Sistem harus menerima data buku baru dan menyimpannya jika seluruh
validasi berhasil.

**FR-BOOK-02 --- Melihat daftar buku**\
Sistem harus menyediakan daftar buku dengan pagination. Client dapat
mengirim parameter `page` dan `page_size`.

**FR-BOOK-03 --- Melihat detail buku**\
Sistem harus menyediakan detail satu buku berdasarkan ID.

**FR-BOOK-04 --- Memperbarui buku**\
Sistem harus mengizinkan pembaruan metadata buku. Field yang tidak
dikirim pada operasi PATCH tidak berubah.

**FR-BOOK-05 --- Menghapus buku**\
Sistem harus menghapus buku hanya jika buku belum pernah memiliki
transaksi peminjaman. Jika memiliki riwayat transaksi, sistem menolak
penghapusan.

**FR-BOOK-06 --- Mencari buku**\
Sistem harus mendukung pencarian berdasarkan judul atau penulis melalui
parameter `q`, tanpa membedakan kapitalisasi huruf.

**FR-BOOK-07 --- Memfilter buku**\
Sistem harus mendukung filter ketersediaan melalui parameter
`available=true|false`.

### 3.2 Pengelolaan Anggota

**FR-MEMBER-01 --- Membuat anggota**\
Sistem harus menyimpan anggota baru setelah validasi berhasil.

**FR-MEMBER-02 --- Melihat daftar anggota**\
Sistem harus menyediakan daftar anggota dengan pagination.

**FR-MEMBER-03 --- Melihat detail anggota**\
Sistem harus menyediakan detail anggota berdasarkan ID.

**FR-MEMBER-04 --- Memperbarui anggota**\
Sistem harus mengizinkan pembaruan nama, email, dan nomor anggota sesuai
aturan validasi.

**FR-MEMBER-05 --- Menghapus anggota**\
Sistem hanya boleh menghapus anggota yang belum pernah memiliki
transaksi peminjaman.

### 3.3 Peminjaman

**FR-LOAN-01 --- Membuat peminjaman**\
Sistem harus membuat transaksi peminjaman jika anggota aktif/terdaftar
dan buku memiliki eksemplar tersedia.

**FR-LOAN-02 --- Menetapkan tanggal**\
Sistem harus menetapkan `borrowed_at` berdasarkan waktu server dan
`due_at` 14 hari kalender setelah waktu peminjaman. Client tidak dapat
menentukan atau memanipulasi kedua nilai tersebut.

**FR-LOAN-03 --- Mengurangi ketersediaan**\
Sistem harus mengurangi jumlah eksemplar tersedia secara konsisten
ketika peminjaman berhasil.

**FR-LOAN-04 --- Menolak peminjaman tidak valid**\
Sistem harus menolak peminjaman jika anggota tidak ditemukan, buku tidak
ditemukan, atau tidak ada eksemplar tersedia.

**FR-LOAN-05 --- Melihat daftar peminjaman**\
Sistem harus menyediakan daftar transaksi dengan pagination. Filter
berdasarkan `member_id` dan `status` harus didukung.

**FR-LOAN-06 --- Melihat detail peminjaman**\
Sistem harus menyediakan detail transaksi berdasarkan ID.

### 3.4 Pengembalian

**FR-RETURN-01 --- Mengembalikan buku**\
Sistem harus mencatat waktu pengembalian berdasarkan waktu server untuk
transaksi yang masih dipinjam.

**FR-RETURN-02 --- Menambah ketersediaan**\
Sistem harus mengembalikan satu unit ke jumlah eksemplar tersedia
setelah pengembalian berhasil.

**FR-RETURN-03 --- Mencegah pengembalian ganda**\
Sistem harus menolak pengembalian untuk transaksi yang sudah
dikembalikan.

### 3.5 Respons dan Error

**FR-API-01 --- Format JSON**\
Semua request body dan response body yang berisi data harus menggunakan
JSON, kecuali response tanpa body.

**FR-API-02 --- Error konsisten**\
Error harus menggunakan format yang konsisten dan tidak mengekspos stack
trace, query SQL, atau kredensial.

**FR-API-03 --- Validasi request**\
Sistem harus menolak JSON malformed, field wajib yang hilang, tipe data
yang salah, dan nilai yang melanggar aturan validasi.

**FR-API-04 --- Health check**\
Sistem harus menyediakan endpoint untuk memeriksa bahwa proses API
berjalan. Status koneksi database dapat diperiksa melalui endpoint
readiness terpisah.

------------------------------------------------------------------------

## 4. Aturan Bisnis dan Validasi

### 4.1 Buku

-   `title` wajib diisi, setelah spasi di awal/akhir dihapus panjangnya
    harus 1--200 karakter.
-   `author` wajib diisi dengan panjang 1--150 karakter.
-   `isbn` wajib diisi, dinormalisasi dengan menghapus tanda hubung dan
    spasi, serta harus unik.
-   Format ISBN-10 atau ISBN-13 sebaiknya divalidasi; jika validasi
    checksum belum dibuat pada MVP awal, format dan keunikannya minimal
    harus divalidasi.
-   `total_copies` harus bilangan bulat lebih besar atau sama dengan
    nol.
-   Jumlah yang sedang dipinjam tidak boleh melebihi `total_copies`.
-   `available_copies` dihitung sebagai
    `total_copies - jumlah transaksi yang belum dikembalikan`, atau
    dijaga dengan mekanisme database yang konsisten. Implementasi tidak
    boleh membiarkan nilainya menjadi negatif.
-   Jika jumlah total eksemplar diturunkan, nilai baru tidak boleh lebih
    kecil daripada jumlah yang sedang dipinjam.

### 4.2 Anggota

-   `name` wajib diisi dan panjangnya 1--150 karakter setelah trim.
-   `email` wajib diisi, harus memiliki format email yang masuk akal,
    dan unik tanpa membedakan kapitalisasi.
-   `member_number` wajib diisi dan unik.
-   Anggota yang sudah mempunyai riwayat transaksi tidak boleh dihapus
    secara permanen.

### 4.3 Peminjaman dan Pengembalian

-   Satu transaksi meminjam tepat satu eksemplar.
-   Peminjaman hanya berhasil jika buku dan anggota ada serta stok
    tersedia.
-   Pembuatan transaksi dan perubahan stok harus bersifat atomik:
    keduanya berhasil bersama atau gagal bersama.
-   Operasi peminjaman harus aman terhadap request bersamaan agar stok
    tidak menjadi negatif atau terjadi peminjaman melebihi stok.
-   Pengembalian hanya dapat dilakukan satu kali untuk transaksi yang
    masih aktif.
-   Pengembalian dan penyesuaian ketersediaan harus bersifat atomik.
-   Transaksi yang sudah dikembalikan tetap disimpan sebagai riwayat.
-   Status peminjaman dihitung dari `returned_at`: `borrowed` jika
    nilainya null dan `returned` jika nilainya terisi.
-   Denda keterlambatan tidak dihitung dalam MVP.

------------------------------------------------------------------------

## 5. Desain Data

Rancangan berikut adalah model konseptual untuk diturunkan menjadi
migration SQL. Nama dan tipe data final dapat disesuaikan selama seluruh
kebutuhan dan constraint tetap terpenuhi.

### 5.1 Tabel `books`

  Kolom            Tipe yang disarankan                           Aturan
  ---------------- ---------------------------------------------- --------------------------------------
  `id`             `BIGINT GENERATED ... AS IDENTITY` atau UUID   Primary key
  `title`          `VARCHAR(200)`                                 Not null
  `author`         `VARCHAR(150)`                                 Not null
  `isbn`           `VARCHAR(13)`                                  Not null, unique setelah normalisasi
  `total_copies`   `INTEGER`                                      Not null, default 0, check \>= 0
  `created_at`     `TIMESTAMPTZ`                                  Not null
  `updated_at`     `TIMESTAMPTZ`                                  Not null

Ketersediaan dapat dihitung dari jumlah eksemplar total dikurangi
transaksi aktif. Jika kolom `available_copies` disimpan untuk memudahkan
locking dan update, nilainya harus dijaga secara atomik dan diuji
terhadap riwayat transaksi.

### 5.2 Tabel `members`

  Kolom             Tipe yang disarankan                           Aturan
  ----------------- ---------------------------------------------- -----------------------------------
  `id`              `BIGINT GENERATED ... AS IDENTITY` atau UUID   Primary key
  `member_number`   `VARCHAR(30)`                                  Not null, unique
  `name`            `VARCHAR(150)`                                 Not null
  `email`           `VARCHAR(254)`                                 Not null, unique case-insensitive
  `created_at`      `TIMESTAMPTZ`                                  Not null
  `updated_at`      `TIMESTAMPTZ`                                  Not null

Keunikan email sebaiknya ditegakkan di database, misalnya menggunakan
unique index pada `LOWER(email)`.

### 5.3 Tabel `loans`

  Kolom           Tipe yang disarankan                           Aturan
  --------------- ---------------------------------------------- -----------------------
  `id`            `BIGINT GENERATED ... AS IDENTITY` atau UUID   Primary key
  `book_id`       Tipe sama dengan `books.id`                    Foreign key, not null
  `member_id`     Tipe sama dengan `members.id`                  Foreign key, not null
  `borrowed_at`   `TIMESTAMPTZ`                                  Not null
  `due_at`        `TIMESTAMPTZ`                                  Not null
  `returned_at`   `TIMESTAMPTZ`                                  Nullable

Gunakan foreign key untuk menjaga referensi. Penghapusan parent
sebaiknya menggunakan perilaku `RESTRICT`/`NO ACTION`, bukan cascade,
agar riwayat tidak terhapus secara tidak sengaja.

### 5.4 Indeks

Indeks minimum yang dipertimbangkan: - Unique index untuk ISBN. - Unique
index untuk nomor anggota. - Unique index case-insensitive untuk
email. - Index pada `loans(book_id)` dan `loans(member_id)`. - Index
yang mendukung pencarian transaksi aktif dan filter daftar.

Indeks pencarian judul/penulis dapat ditambahkan setelah pola query dan
kebutuhan pencarian ditentukan.

------------------------------------------------------------------------

## 6. Spesifikasi Endpoint

Semua endpoint berada di bawah `/api/v1`. `{id}` adalah ID resource.
Bentuk response di bawah bersifat kontrak konseptual; implementasi boleh
menambahkan metadata yang konsisten.

### 6.1 Health

  Method   Endpoint    Sukses
  -------- ----------- ---------------------------------------------------
  `GET`    `/health`   `200 OK`
  `GET`    `/ready`    `200 OK` jika API siap dan database dapat diakses

### 6.2 Books

  -----------------------------------------------------------------------------
  Method            Endpoint               Sukses             Catatan
  ----------------- ---------------------- ------------------ -----------------
  `POST`            `/api/v1/books`        `201 Created`      Membuat buku

  `GET`             `/api/v1/books`        `200 OK`           Query: `page`,
                                                              `page_size`, `q`,
                                                              `available`

  `GET`             `/api/v1/books/{id}`   `200 OK`           Detail buku

  `PATCH`           `/api/v1/books/{id}`   `200 OK`           Update parsial

  `DELETE`          `/api/v1/books/{id}`   `204 No Content`   Hanya jika belum
                                                              ada riwayat loan
  -----------------------------------------------------------------------------

### 6.3 Members

  -------------------------------------------------------------------------------
  Method            Endpoint                 Sukses             Catatan
  ----------------- ------------------------ ------------------ -----------------
  `POST`            `/api/v1/members`        `201 Created`      Membuat anggota

  `GET`             `/api/v1/members`        `200 OK`           Query: `page`,
                                                                `page_size`, `q`

  `GET`             `/api/v1/members/{id}`   `200 OK`           Detail anggota

  `PATCH`           `/api/v1/members/{id}`   `200 OK`           Update parsial

  `DELETE`          `/api/v1/members/{id}`   `204 No Content`   Hanya jika belum
                                                                ada riwayat loan
  -------------------------------------------------------------------------------

### 6.4 Loans

  -----------------------------------------------------------------------------------
  Method            Endpoint                      Sukses            Catatan
  ----------------- ----------------------------- ----------------- -----------------
  `POST`            `/api/v1/loans`               `201 Created`     Body berisi
                                                                    `book_id` dan
                                                                    `member_id`

  `GET`             `/api/v1/loans`               `200 OK`          Query: `page`,
                                                                    `page_size`,
                                                                    `member_id`,
                                                                    `status`

  `GET`             `/api/v1/loans/{id}`          `200 OK`          Detail peminjaman

  `POST`            `/api/v1/loans/{id}/return`   `200 OK`          Mengembalikan
                                                                    buku
  -----------------------------------------------------------------------------------

### 6.5 Contoh Request

Membuat buku:

``` json
{
  "title": "Clean Code",
  "author": "Robert C. Martin",
  "isbn": "9780132350884",
  "total_copies": 5
}
```

Membuat anggota:

``` json
{
  "member_number": "MBR-0001",
  "name": "Alya Putri",
  "email": "alya@example.com"
}
```

Membuat peminjaman:

``` json
{
  "book_id": 1,
  "member_id": 1
}
```

Client tidak mengirim `borrowed_at`, `due_at`, atau `returned_at` saat
membuat peminjaman. Nilai-nilai tersebut dikendalikan server.

### 6.6 Contoh Response

Response buku:

``` json
{
  "data": {
    "id": 1,
    "title": "Clean Code",
    "author": "Robert C. Martin",
    "isbn": "9780132350884",
    "total_copies": 5,
    "available_copies": 4,
    "created_at": "2026-10-21T10:00:00Z",
    "updated_at": "2026-10-21T10:00:00Z"
  }
}
```

Response daftar:

``` json
{
  "data": [],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total_items": 0,
    "total_pages": 0
  }
}
```

Response error:

``` json
{
  "error": {
    "code": "BOOK_NOT_FOUND",
    "message": "Book was not found"
  }
}
```

Pesan error boleh ditulis dalam bahasa Indonesia jika dipilih secara
konsisten. `code` harus stabil agar client tidak bergantung pada teks
pesan.

------------------------------------------------------------------------

## 7. HTTP Status Code dan Error Contract

  -----------------------------------------------------------------------
  Kondisi                                                          Status
  ------------------------------ ----------------------------------------
  Request berhasil                                               `200 OK`

  Resource berhasil dibuat                                  `201 Created`

  Berhasil tanpa response body                           `204 No Content`

  JSON malformed atau parameter                         `400 Bad Request`
  tidak valid                    

  Resource tidak ditemukan                                `404 Not Found`

  Duplikasi ISBN/email/nomor                               `409 Conflict`
  anggota atau konflik keadaan   
  resource                       

  Content-Type tidak didukung                `415 Unsupported Media Type`

  Payload terlalu besar                           `413 Content Too Large`

  Kesalahan tak terduga                       `500 Internal Server Error`

  Database sementara tidak siap                 `503 Service Unavailable`
  -----------------------------------------------------------------------

Ketentuan: - Error tidak boleh mengandung SQL mentah, stack trace, atau
secret. - Error database yang diketahui harus dipetakan ke error API
yang sesuai. - `400` digunakan untuk input yang tidak valid; `409`
digunakan untuk konflik terhadap keadaan resource atau constraint
unik. - Endpoint `DELETE` yang berhasil mengembalikan `204` tanpa body.

Kode error aplikasi minimum: - `VALIDATION_ERROR` - `BOOK_NOT_FOUND` -
`MEMBER_NOT_FOUND` - `LOAN_NOT_FOUND` - `DUPLICATE_ISBN` -
`DUPLICATE_EMAIL` - `DUPLICATE_MEMBER_NUMBER` - `BOOK_UNAVAILABLE` -
`BOOK_HAS_LOAN_HISTORY` - `MEMBER_HAS_LOAN_HISTORY` -
`LOAN_ALREADY_RETURNED` - `INTERNAL_ERROR`

------------------------------------------------------------------------

## 8. Kebutuhan Nonfungsional

**NFR-01 --- Konsistensi data**\
Peminjaman dan pengembalian harus memakai transaksi database untuk
menjaga konsistensi antara riwayat transaksi dan ketersediaan buku.

**NFR-02 --- Keamanan konfigurasi**\
Kredensial database tidak boleh di-hardcode atau dimasukkan ke
repository Git.

**NFR-03 --- Keamanan error**\
Response error tidak boleh membocorkan detail internal database atau
konfigurasi.

**NFR-04 --- Maintainability**\
Kode harus memiliki pembagian tanggung jawab yang jelas antara HTTP
handler, business logic/service, dan akses data/repository. Abstraksi
tambahan hanya dibuat ketika memberi manfaat yang jelas.

**NFR-05 --- Testability**\
Aturan bisnis dan alur penting harus dapat diuji otomatis. Operasi
database kritis harus memiliki integration test.

**NFR-06 --- Database migration**\
Perubahan skema harus dikelola melalui migration yang dapat diterapkan
secara berulang dan dapat dilacak, bukan hanya perubahan manual di GUI
database.

**NFR-07 --- Pagination aman**\
`page` minimal 1. `page_size` memiliki default 20 dan maksimum 100.
Nilai di luar aturan harus ditangani secara konsisten.

**NFR-08 --- Batas input**\
Request body memiliki batas ukuran yang wajar. API menolak payload yang
terlalu besar dan JSON dengan tipe data yang salah.

**NFR-09 --- Waktu dan zona waktu**\
Waktu transaksi ditetapkan oleh server dan disimpan dalam `TIMESTAMPTZ`.
Response waktu menggunakan format RFC 3339/ISO 8601.

**NFR-10 --- Observabilitas dasar**\
Server mencatat method, path, status code, dan durasi request tanpa
mencatat password, token, atau data sensitif yang tidak diperlukan.

**NFR-11 --- Konkurensi**\
Peminjaman bersamaan tidak boleh menyebabkan jumlah peminjaman aktif
melebihi jumlah eksemplar. Mekanisme locking atau conditional update
harus diuji.

**NFR-12 --- Performa awal**\
Untuk lingkungan pengembangan dan dataset kecil hingga menengah,
endpoint daftar harus menggunakan pagination dan query yang efisien.
Target angka performa spesifik belum ditetapkan pada MVP.

------------------------------------------------------------------------

## 9. Arsitektur Implementasi yang Disarankan

Implementasi menggunakan monolit sederhana, bukan microservices.

Pembagian tanggung jawab yang disarankan:

-   **Handler:** parsing request, validasi bentuk input, memanggil
    service, menulis status dan JSON response.
-   **Service:** aturan bisnis dan pengaturan transaksi use case.
-   **Repository:** query database dan pemetaan hasil database.
-   **Model/domain:** representasi data dan konsep domain.
-   **Middleware:** logging, recovery, dan middleware umum lain jika
    diperlukan.
-   **Configuration:** membaca konfigurasi dari environment.
-   **Migration:** menyimpan versi skema database.
-   **Tests:** unit test dan integration test.

Struktur direktori awal yang mungkin:

``` text
zy-library-manager-api/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── book/
│   ├── member/
│   ├── loan/
│   ├── httpapi/
│   └── database/
├── migrations/
├── tests/
├── .env.example
├── .gitignore
├── go.mod
└── README.md
```

Struktur ini adalah usulan awal, bukan kewajiban mutlak. Jangan membuat
file, interface, atau package yang belum memiliki tanggung jawab nyata.

------------------------------------------------------------------------

## 10. Rencana Pengujian dan Kriteria Penerimaan

### 10.1 Pengujian Buku

-   Buku valid dapat dibuat dan muncul pada daftar.
-   Detail buku yang ada dapat diambil.
-   ID yang tidak ada menghasilkan `404`.
-   Input kosong, ISBN duplikat, dan jumlah eksemplar negatif ditolak.
-   Pencarian, filter, dan pagination bekerja sesuai kontrak.
-   Buku dengan riwayat peminjaman tidak dapat dihapus.

### 10.2 Pengujian Anggota

-   Anggota valid dapat dibuat dan diambil.
-   Email dan nomor anggota duplikat ditolak.
-   Email tidak valid ditolak.
-   Anggota dengan riwayat peminjaman tidak dapat dihapus.

### 10.3 Pengujian Peminjaman

-   Peminjaman berhasil jika anggota ada dan buku tersedia.
-   Peminjaman gagal jika anggota atau buku tidak ada.
-   Peminjaman gagal jika stok habis.
-   Tanggal peminjaman dan jatuh tempo dibuat oleh server.
-   Kegagalan di tengah proses tidak meninggalkan perubahan data
    parsial.
-   Request bersamaan tidak dapat meminjam melebihi stok.

### 10.4 Pengujian Pengembalian

-   Pengembalian transaksi aktif berhasil.
-   Pengembalian mengubah status menjadi `returned`.
-   Pengembalian kedua kali ditolak dengan `409 Conflict`.
-   Kegagalan proses tidak menyebabkan stok berubah tanpa perubahan
    transaksi.

### 10.5 Pengujian API Umum

-   JSON malformed ditolak dengan error yang konsisten.
-   Status code sesuai kontrak.
-   Error tidak membocorkan detail internal.
-   Pagination menolak nilai yang tidak valid atau menerapkan aturan
    normalisasi yang telah didokumentasikan.
-   Health check bekerja dan readiness mendeteksi masalah koneksi
    database.

------------------------------------------------------------------------

## 11. Urutan Implementasi

Pengerjaan dilakukan dalam irisan kecil yang bisa diuji, bukan
sekaligus.

1.  **Fondasi proyek:** inisialisasi Go module, konfigurasi, koneksi
    PostgreSQL, health check, dan migration awal.
2.  **Books --- vertikal pertama:** model, repository, service bila
    diperlukan, handler, validasi, CRUD, dan pengujian.
3.  **Penyempurnaan books:** pencarian, pagination, constraint, serta
    konsistensi error.
4.  **Members:** skema, CRUD, validasi, unique constraint, dan
    pengujian.
5.  **Loans:** foreign key, business rules, transaksi database, serta
    pengujian stok dan konkurensi.
6.  **Returns:** pengembalian atomik, penolakan pengembalian ganda, dan
    pengujian.
7.  **Hardening:** logging, penanganan konfigurasi, pengujian integrasi,
    dokumentasi penggunaan, dan review keamanan.
8.  **Ekstensi setelah MVP stabil:** autentikasi/otorisasi, denda,
    reservasi, dan deployment.

Setiap tahap harus memiliki kriteria selesai sebelum lanjut ke tahap
berikutnya. Jika bug ditemukan, perbaiki dan tambahkan regression test
sebelum menambah fitur baru.

------------------------------------------------------------------------

## 12. Hal yang Sengaja Ditunda

Hal berikut tidak menjadi syarat penyelesaian MVP: - JWT dan refresh
token. - Role/permission yang kompleks. - Denda keterlambatan. -
Reservasi dan antrean buku. - Notifikasi. - Docker dan deployment
cloud. - Caching. - Microservices. - Observability tingkat lanjut.

Hal-hal tersebut dapat menjadi versi lanjutan setelah kebutuhan inti
berjalan stabil.

------------------------------------------------------------------------

## 13. Definisi Selesai (Definition of Done)

MVP dianggap selesai ketika: 1. Seluruh endpoint yang tercantum telah
diimplementasikan atau perubahan kontraknya telah didokumentasikan. 2.
Skema database dibuat melalui migration. 3. Constraint dan business
rules utama ditegakkan. 4. Semua skenario penerimaan kritis lulus. 5.
Peminjaman dan pengembalian konsisten saat terjadi error atau request
bersamaan. 6. Konfigurasi sensitif tidak tersimpan dalam repository. 7.
README menjelaskan cara konfigurasi, migration, menjalankan server, dan
menjalankan test. 8. API dapat diuji dari client HTTP tanpa frontend.

------------------------------------------------------------------------

## Catatan Penutup

SRS ini sengaja membatasi MVP pada satu API monolitik dengan tiga
resource utama: buku, anggota, dan peminjaman. Dokumen ini adalah
baseline implementasi, bukan alasan untuk membangun semua fitur
sekaligus. Jika ditemukan kebutuhan yang bertentangan atau keputusan
yang belum jelas saat implementasi, catat perubahan pada dokumen sebelum
mengubah kontrak API atau skema data.

JANGAN MODIFIKASI KODE SENDIRI KECUALI PENGGUNA MEMINTANYA.
JELASKAN LANGKAH-LANGKAH TERPERINCI APA YANG HARUS DILAKUKAN PENGGUNA
PADA KODE PROGRAM DI SETIAP FILE.