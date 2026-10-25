package book

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
	"strings"
	"unicode/utf8"
)

// Deklarasi interface bookCreator
type bookCreator interface {
	Create(context.Context, CreateBookRequest) (Book, error)
}

// Method Create milik struct Handler
type Handler struct {
	creator bookCreator
}

// Constructor untuk struct Handler.
func NewHandler(creator bookCreator) *Handler {
	return &Handler{creator: creator}
}

// Method Create milik struct Handler
func (handler *Handler) Create(w http.ResponseWriter, r *http.Request) {
	// Memastikan format mediaType adalah application/json.
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		// Throw error ke client tanpa dicatat di log server (client error).
		writeError(
			w,
			http.StatusUnsupportedMediaType,
			"UNSUPPORTED_MEDIA_TYPE",
			"Content-Type must be application/json",
		)
		return
	}

	// Membatasi ukuran body harus di bawah 1 MB (2^20 Bytes).
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var input CreateBookRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		var maxBytesError *http.MaxBytesError
		// Jika JSON melebihi batas 1 MB, throw error ke client.
		if errors.As(err, &maxBytesError) {
			writeError(
				w,
				http.StatusRequestEntityTooLarge,
				"PAYLOAD_TOO_LARGE",
				"Request body is too large",
			)
			return
		}

		// Error lain (field JSON tidak valid, dll).
		writeError(
			w,
			http.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request body must contain valid JSON",
		)
		return
	}

	// Memeriksa apakah JSON body punya lebih dari 1 objek
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		writeError(
			w,
			http.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request body must contain exactly one JSON object",
		)
		return
	}

	// Normalisasi title, author, dan ISBN.
	input.Title = strings.TrimSpace(input.Title)
	input.Author = strings.TrimSpace(input.Author)
	input.ISBN = normalizeISBN(input.ISBN)

	// Validasi gagal (ada kesalahan dalam input data)
	if err := validateCreateBook(input); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Jalankan query
	created, err := handler.creator.Create(r.Context(), input)
	if errors.Is(err, ErrDuplicateISBN) {
		// Jika ISBN sudah ada, throw error ke client
		writeError(
			w,
			http.StatusConflict,
			"DUPLICATE_ISBN",
			"A book with this ISBN already exists",
		)
		return
	}
	if err != nil {
		// Error dari database HANYA dicatat di log server
		log.Printf("create book: %v", err)
		// Throw error 500 ke client
		writeError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"An unexpected error occured",
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]any{
		"data": created,
	}); err != nil {
		log.Printf("write create book response: %v", err)
	}
}

func normalizeISBN(isbn string) string {
	isbn = strings.TrimSpace(isbn)
	isbn = strings.ReplaceAll(isbn, "-", "")
	isbn = strings.ReplaceAll(isbn, " ", "")
	return isbn
}

func validateCreateBook(input CreateBookRequest) error {
	titleLength := utf8.RuneCountInString(input.Title)
	if titleLength < 1 || titleLength > 200 {
		return errors.New("title must be between 1 and 200 characters")
	}

	authorLength := utf8.RuneCountInString(input.Author)
	if authorLength < 1 || authorLength > 150 {
		return errors.New("author must be between 1 and 200 characters")
	}

	if !validISBNFormat(input.ISBN) {
		return errors.New("ISBN must be a 10-character ISBN-10 or a 13-digit ISBN-13")
	}

	if input.TotalCopies < 0 {
		return errors.New("total copies must be zero or greater")
	}

	return nil
}

func validISBNFormat(isbn string) bool {
	if len(isbn) == 10 {
		for i := 0; i < 9; i++ {
			if isbn[i] < '0' || isbn[i] > '9' {
				return false
			}
		}
		last := isbn[9]
		return (last >= '0' && last <= '9') || last == 'X' || last == 'x'
	}

	if len(isbn) == 13 {
		for i := 0; i < 13; i++ {
			if isbn[i] < '0' || isbn[i] > '9' {
				return false
			}
		}
		return true
	}

	return false
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	}); err != nil {
		log.Printf("write error response: %v", err)
	}
}
