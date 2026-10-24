-- +goose Up
CREATE TABLE books (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title VARCHAR(200) NOT NULL
        CHECK (char_length(btrim(title)) BETWEEN 1 AND 200),
    author VARCHAR(150) NOT NULL
        CHECK (char_length(btrim(author)) BETWEEN 1 AND 150),
    isbn VARCHAR(13) NOT NULL
        CHECK (isbn ~ '^([0-9]{9}[0-9Xx]|[0-9]{13}$)'),
    total_copies INTEGER NOT NULL DEFAULT 0
        CHECK (total_copies >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT books_isbn_unique UNIQUE (isbn)
);

-- +goose Down
DROP TABLE books;
