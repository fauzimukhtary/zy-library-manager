package book

import "time"

// Struct request dari client.
type CreateBookRequest struct {
	Title       string `json:"title"`
	Author      string `json:"author"`
	ISBN        string `json:"isbn"`
	TotalCopies int    `json:"total_copies"`
}

// Struct response dari aplikasi.
type Book struct {
	ID              int64     `json:"id"`
	Title           string    `json:"title"`
	Author          string    `json:"author"`
	ISBN            string    `json:"isbn"`
	TotalCopies     int       `json:"total_copies"`
	AvailableCopies int       `json:"available_copies"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

type ListBooksResponse struct {
	Data       []Book     `json:"data"`
	Pagination Pagination `json:"pagination"`
}
