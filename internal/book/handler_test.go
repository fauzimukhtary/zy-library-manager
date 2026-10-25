package book_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"zy-library-manager/internal/book"
)

type fakeCreator struct {
	input  book.CreateBookRequest
	result book.Book
	err    error
	called bool
}

func (fake *fakeCreator) Create(
	_ context.Context,
	input book.CreateBookRequest,
) (book.Book, error) {
	fake.called = true
	fake.input = input
	return fake.result, fake.err
}

func TestCreateBookReturnsCreatedAndNormalizesInput(t *testing.T) {
	nowBook := book.Book{
		ID:              1,
		Title:           "Clean Code",
		Author:          "Robert C. Martin",
		ISBN:            "9780132350884",
		TotalCopies:     5,
		AvailableCopies: 5,
	}

	creator := &fakeCreator{result: nowBook}
	handler := book.NewHandler(creator)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/books",
		strings.NewReader(`{
            "title": "  Clean Code  ",
            "author": " Robert C. Martin ",
            "isbn": "978-0-132350-88-4",
            "total_copies": 5
        }`),
	)
	request.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()
	handler.Create(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
	if !creator.called {
		t.Fatal("expected repository to be called")
	}
	if creator.input.Title != "Clean Code" {
		t.Errorf("expected trimmed title, got %q", creator.input.Title)
	}
	if creator.input.Author != "Robert C. Martin" {
		t.Errorf("expected trimmed author, got %q", creator.input.Author)
	}
	if creator.input.ISBN != "9780132350884" {
		t.Errorf("expected normalized ISBN, got %q", creator.input.ISBN)
	}

	var responseBody struct {
		Data book.Book `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &responseBody); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if responseBody.Data.ID != 1 {
		t.Errorf("expected book ID 1, got %d", responseBody.Data.ID)
	}
}

func TestCreateBookRejectsInvalidInput(t *testing.T) {
	creator := &fakeCreator{}
	handler := book.NewHandler(creator)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/books",
		strings.NewReader(`{
            "title": "   ",
            "author": "Robert C. Martin",
            "isbn": "9780132350884",
            "total_copies": 5
        }`),
	)
	request.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()
	handler.Create(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
	if creator.called {
		t.Fatal("repository should not be called for invalid input")
	}
}

func TestCreateBookReturnsConflictForDuplicateISBN(t *testing.T) {
	creator := &fakeCreator{err: book.ErrDuplicateISBN}
	handler := book.NewHandler(creator)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/books",
		strings.NewReader(`{
            "title": "Clean Code",
            "author": "Robert C. Martin",
            "isbn": "9780132350884",
            "total_copies": 5
        }`),
	)
	request.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()
	handler.Create(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, response.Code)
	}

	var responseBody struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &responseBody); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if responseBody.Error.Code != "DUPLICATE_ISBN" {
		t.Errorf("expected DUPLICATE_ISBN, got %q", responseBody.Error.Code)
	}
}

func TestCreateBookHidesUnexpectedRepositoryError(t *testing.T) {
	creator := &fakeCreator{err: errors.New("database details")}
	handler := book.NewHandler(creator)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/books",
		strings.NewReader(`{
            "title": "Clean Code",
            "author": "Robert C. Martin",
            "isbn": "9780132350884",
            "total_copies": 5
        }`),
	)
	request.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()
	handler.Create(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, response.Code)
	}
	if strings.Contains(response.Body.String(), "database details") {
		t.Fatal("response must not expose internal error details")
	}
}
