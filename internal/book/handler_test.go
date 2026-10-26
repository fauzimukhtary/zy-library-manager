package book_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"zy-library-manager/internal/book"
)

type fakeCreator struct {
	input         book.CreateBookRequest
	createResult  book.Book
	createErr     error
	creatorCalled bool

	listResult []book.Book
	totalItems int
	listErr    error
	listLimit  int
	listOffset int
	listCalled bool
}

func (fake *fakeCreator) Create(
	_ context.Context,
	input book.CreateBookRequest,
) (book.Book, error) {
	fake.creatorCalled = true
	fake.input = input
	return fake.createResult, fake.createErr
}

func (fake *fakeCreator) List(
	_ context.Context,
	limit int,
	offset int,
) ([]book.Book, int, error) {
	fake.listCalled = true
	fake.listLimit = limit
	fake.listOffset = offset
	return fake.listResult, fake.totalItems, fake.listErr
}

func TestListBookReturnsAppropriatePaginationProperties(t *testing.T) {
	creator := &fakeCreator{totalItems: 21}
	handler := book.NewHandler(creator)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/books",
		nil,
	)

	query := url.Values{}
	query.Add("page", "2")
	query.Add("page_size", "10")

	request.URL.RawQuery = query.Encode()

	response := httptest.NewRecorder()
	handler.List(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	if !creator.listCalled {
		t.Fatal("expected repository to be called")
	}

	var responseBody book.ListBooksResponse

	if err := json.Unmarshal(response.Body.Bytes(), &responseBody); err != nil {
		t.Fatalf("list decode response: %v", err)
	}

	if responseBody.Pagination.TotalItems != 21 {
		t.Errorf("expected total items = 21, got %v", responseBody.Pagination.TotalItems)
	}

	if responseBody.Pagination.Page != 2 {
		t.Errorf("expected page = 2, got %v", responseBody.Pagination.Page)
	}

	if responseBody.Pagination.PageSize != 10 {
		t.Errorf("expected page size = 10, got %v", responseBody.Pagination.PageSize)
	}

	if responseBody.Pagination.TotalPages != 3 {
		t.Errorf("expected total pages = 3, got %v", responseBody.Pagination.TotalPages)
	}

	if creator.listLimit != 10 {
		t.Errorf("expected limit = 10, got %v", creator.listLimit)
	}

	if creator.listOffset != 10 {
		t.Errorf("expected offset = 10, got %v", creator.listOffset)
	}
}

func TestListBookInvalidParameter(t *testing.T) {
	creator := &fakeCreator{}
	handler := book.NewHandler(creator)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/books",
		nil,
	)

	query := url.Values{}
	query.Add("page", "0")

	request.URL.RawQuery = query.Encode()

	response := httptest.NewRecorder()
	handler.List(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %v",
			http.StatusBadRequest,
			response.Code,
			response.Body.String(),
		)
	}

	if creator.listCalled {
		t.Fatal("expected repository not to be called")
	}
}

func TestListBookHidesUnexpectedRepositoryError(t *testing.T) {
	creator := &fakeCreator{listErr: errors.New("database details")}
	handler := book.NewHandler(creator)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/books",
		nil,
	)

	response := httptest.NewRecorder()
	handler.List(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected status code: %d, got %d", http.StatusInternalServerError, response.Code)
	}

	if strings.Contains(response.Body.String(), "database details") {
		t.Fatal("response must not expose internal error details")
	}
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

	creator := &fakeCreator{createResult: nowBook}
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
	if !creator.creatorCalled {
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
		t.Fatalf("create decode response: %v", err)
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
	if creator.creatorCalled {
		t.Fatal("repository should not be called for invalid input")
	}
}

func TestCreateBookRejectsInputWithMissingJSONField(t *testing.T) {
	creator := &fakeCreator{}
	handler := book.NewHandler(creator)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/books",
		strings.NewReader(`{
            "title": "Clean Code",
            "author": "Robert C. Martin",
            "isbn": "9780132350884"
        }`),
	)
	request.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()
	handler.Create(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
	if creator.creatorCalled {
		t.Fatal("repository should not be called for invalid input (JSON field missing)")
	}
}

func TestCreateBookRejectsInvalidAuthorInput(t *testing.T) {
	creator := &fakeCreator{}
	handler := book.NewHandler(creator)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/books",
		strings.NewReader(`{
            "title": "Clean Code",
            "author": "QWERTYUIOPASDFGHJKLZXCVBNMQWERTYUIOPASDFGHJKLZXCVBNMQWERTYUIOPASDFGHJKLZXCVBNMQWERTYUIOPASDFGHJKLZXCVBNMQWERTYUIOPASDFGHJKLZXCVBNMQWERTYUIOPASDFGHJKLZXCVBNM",
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
	if creator.creatorCalled {
		t.Fatal("repository should not be called for invalid input (JSON field missing)")
	}

	var responseBody struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(response.Body.Bytes(), &responseBody); err != nil {
		t.Fatalf("create rejects invalid author input decode response: %v", err)
	}

	if responseBody.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got %q", responseBody.Error.Code)
	}

	authorErr := "author must be between 1 and 150 characters"
	if responseBody.Error.Message != authorErr {
		t.Errorf("expected error message: %v, got %v", authorErr, responseBody.Error.Message)
	}
}

func TestCreateBookRejectsInvalidTitleInput(t *testing.T) {
	creator := &fakeCreator{}
	handler := book.NewHandler(creator)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/books",
		strings.NewReader(`{
            "title": "QWERTYUIOPASDFGHJKLZXCVBNMQWERTYUIOPASDFGHJKLZXCVBNMQWERTYUIOPASDFGHJKLZXCVBNMQWERTYUIOPASDFGHJKLZXCVBNMQWERTYUIOPASDFGHJKLZXCVBNMQWERTYUIOPASDFGHJKLZXCVBNMQWERTYUIOPASDFGHJKLZXCVBNMQWERTYUIOPASDFGHJKLZXCVBNM",
            "author": "Charles H. Brown",
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
	if creator.creatorCalled {
		t.Fatal("repository should not be called for invalid input (JSON field missing)")
	}

	var responseBody struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(response.Body.Bytes(), &responseBody); err != nil {
		t.Fatalf("create rejects invalid title input decode response: %v", err)
	}

	if responseBody.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got %q", responseBody.Error.Code)
	}

	titleErr := "title must be between 1 and 200 characters"
	if responseBody.Error.Message != titleErr {
		t.Errorf("expected error message: %q, got %q", titleErr, responseBody.Error.Message)
	}
}

func TestCreateBookReturnsConflictForDuplicateISBN(t *testing.T) {
	creator := &fakeCreator{createErr: book.ErrDuplicateISBN}
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
	creator := &fakeCreator{createErr: errors.New("database details")}
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
