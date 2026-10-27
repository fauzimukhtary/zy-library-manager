# ZY Library Manager API

Backend REST API untuk mengelola katalog buku, anggota, dan transaksi
peminjaman perpustakaan.

## Status Implementasi

Saat ini sudah tersedia:

- Health check
- Koneksi PostgreSQL
- Migration tabel books
- Membuat buku
- Melihat daftar buku
- Pagination
- Pencarian judul dan penulis
- Filter ketersediaan
- Validasi JSON dan format ISBN
- Penanganan error konsisten

Fitur berikutnya:

- Melihat detail buku
- Memperbarui buku
- Menghapus buku
- CRUD anggota
- Peminjaman dan pengembalian

## Teknologi

- Go
- PostgreSQL
- pgx melalui database/sql
- Goose untuk migration

## Persyaratan

- Go 1.25 atau lebih baru
- PostgreSQL 14 atau lebih baru
- Goose CLI
- Windows PowerShell atau terminal lain

## Struktur Direktori

```text
zy-library-manager-api/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── book/
│   ├── database/
│   ├── httpapi/
│   ├── loan/
│   └── member/
├── migrations/
├── tests/
└── docs/
    └── SRS.md
```