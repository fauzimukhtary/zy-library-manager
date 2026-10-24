-- +goose Up
ALTER TABLE books
    DROP CONSTRAINT books_isbn_check;

ALTER TABLE books
    ADD CONSTRAINT books_isbn_format_check
    CHECK (isbn ~ '^([0-9]{9}[0-9Xx]|[0-9]{13})$');

-- +goose Down
ALTER TABLE books
    DROP CONSTRAINT books_isbn_format_check;

ALTER TABLE books
    ADD CONSTRAINT books_isbn_check
    CHECK (isbn ~ '^([0-9]{9}[0-9Xx]|[0-9]{13}$)');
