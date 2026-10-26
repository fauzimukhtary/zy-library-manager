package book

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Deklarasi interface bookCreator
type bookCreator interface {
	Create(context.Context, CreateBookRequest) (Book, error)
	List(context.Context, int, int) ([]Book, int, error)
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

	// Membatasi ukuran body di bawah 1 MB (2^20 Bytes).
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	input := CreateBookRequest{TotalCopies: -1}
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

	// QUERY EXECUTION
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

// Method List milik struct Handler
func (handler *Handler) List(w http.ResponseWriter, r *http.Request) {
	// Mengambil nilai parameter page
	page, err := parsePageParameter(r, "page", 1, 0)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Mengambil nilai parameter page_size
	pageSize, err := parsePageParameter(r, "page_size", 20, 100)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	maxInt := int(^uint(0) >> 1)
	if page-1 > maxInt/pageSize {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "page is too large")
		return
	}
	offset := (page - 1) * pageSize

	// QUERY EXECUTION
	books, totalItems, err := handler.creator.List(r.Context(), pageSize, offset)
	if err != nil {
		// Error dari database HANYA dicatat di log server.
		log.Printf("list books: %v", err)
		writeError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"An unexpected error occured",
		)
		return
	}

	// Total halaman menurut pageSize
	totalPages := 0
	if totalItems > 0 {
		totalPages = (totalItems-1)/pageSize + 1
	}

	// Kirim response ke client
	response := ListBooksResponse{
		Data: books,
		Pagination: Pagination{
			Page:       page,
			PageSize:   pageSize,
			TotalItems: totalItems,
			TotalPages: totalPages,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("write list books response: %v", err)
	}
}

func normalizeISBN(isbn string) string {
	isbn = strings.TrimSpace(isbn)
	isbn = strings.ReplaceAll(isbn, "-", "")
	isbn = strings.ReplaceAll(isbn, " ", "")
	return isbn
}

func parsePageParameter(
	r *http.Request,
	name string,
	defaultValue int,
	maxValue int,
) (int, error) {
	values, exists := r.URL.Query()[name]
	if !exists {
		return defaultValue, nil
	}
	if len(values) != 1 || values[0] == "" {
		return 0, errors.New(name + " must be a positive integer")
	}

	value, err := strconv.Atoi(values[0])
	if err != nil || value < 1 {
		return 0, errors.New(name + " must be a positive integer")
	}
	if maxValue > 0 && value > maxValue {
		return 0, errors.New(name + " must not exceed 100")
	}

	return value, nil
}

func validateCreateBook(input CreateBookRequest) error {
	titleLength := utf8.RuneCountInString(input.Title)
	if titleLength < 1 || titleLength > 200 {
		return errors.New("title must be between 1 and 200 characters")
	}

	authorLength := utf8.RuneCountInString(input.Author)
	if authorLength < 1 || authorLength > 150 {
		return errors.New("author must be between 1 and 150 characters")
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
