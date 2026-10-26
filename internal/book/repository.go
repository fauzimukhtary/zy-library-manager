package book

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var ErrDuplicateISBN = errors.New("book ISBN already exists")

// Deklarasi struct Repository.
type Repository struct {
	db *sql.DB
}

// Constructor untuk struct Repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Method Create dimiliki struct Repository.
func (repo *Repository) Create(
	ctx context.Context, // Untuk pengendalian lifecycle.
	input CreateBookRequest, // Masukan data dari pengguna diambil dari JSON.
) (Book, error) {
	// Query digunakan untuk memasukkan data KECUALI ISBN-nya ada duplikat,
	// maka tidak akan dilakukan penambahan data dan return value kosong.
	const query = `
		INSERT INTO books (title, author, isbn, total_copies)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (isbn) DO NOTHING
		RETURNING id, title, author, isbn, total_copies, created_at, updated_at;
	`

	// Eksekusi query sekaligus mengembalikan data id, title, author, dll,
	// dimasukkan ke struct Book (created) yang baru dibuat.
	var created Book
	err := repo.db.QueryRowContext(
		ctx,
		query,
		input.Title,
		input.Author,
		input.ISBN,
		input.TotalCopies,
	).Scan(
		&created.ID,
		&created.Title,
		&created.Author,
		&created.ISBN,
		&created.TotalCopies,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	// Error jika ada ISBN yang sama.
	if errors.Is(err, sql.ErrNoRows) {
		return Book{}, ErrDuplicateISBN
	}
	// Error lain.
	if err != nil {
		return Book{}, fmt.Errorf("book repository: create book error: %w", err)
	}

	// Buku baru memiliki peminjaman 0.
	created.AvailableCopies = created.TotalCopies

	// Created berisi data buku yang baru saja ditambahkan.
	return created, nil
}

func (repo *Repository) List(
	ctx context.Context,
	limit int,
	offset int,
) ([]Book, int, error) {
	const countQuery = `SELECT COUNT(*) FROM books;`

	var totalItems int
	if err := repo.db.QueryRowContext(ctx, countQuery).Scan(&totalItems); err != nil {
		return nil, 0, fmt.Errorf("count books: %w", err)
	}

	const listQuery = `
		SELECT
			id,
			title,
			author,
			isbn,
			total_copies,
			total_copies AS available_copies,
			created_at,
			updated_at
		FROM books
		ORDER BY id
		LIMIT $1 OFFSET $2;
	`

	rows, err := repo.db.QueryContext(ctx, listQuery, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list books: %w", err)
	}
	defer rows.Close()

	books := make([]Book, 0)
	for rows.Next() {
		var item Book
		if err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Author,
			&item.ISBN,
			&item.TotalCopies,
			&item.AvailableCopies,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan book: %w", err)
		}

		books = append(books, item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate books: %w", err)
	}

	return books, totalItems, nil
}
