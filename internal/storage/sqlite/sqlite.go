package sqlite

import (
	"database/sql"
	"fmt"
	"strings"

	"go-short-url/internal/storage"

	_ "modernc.org/sqlite"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Создаём таблицу, если не существует
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS url (
            id INTEGER PRIMARY KEY,
            alias TEXT NOT NULL UNIQUE,
            url TEXT NOT NULL
        );
    `)
	if err != nil {
		return nil, fmt.Errorf("%s (create table): %w", op, err)
	}

	// Индекс на alias
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);`)
	if err != nil {
		return nil, fmt.Errorf("%s (create index): %w", op, err)
	}

	return &Storage{db: db}, nil
}

// SaveURL — сохраняет URL и alias
func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const op = "storage.sqlite.SaveURL"

	stmt, err := s.db.Prepare("INSERT INTO url(url, alias) VALUES (?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(urlToSave, alias)
	if err != nil {
		// alias уже существует
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", op, err)
	}

	return id, nil
}

// GetURL — получить URL по alias
func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.sqlite.GetURL"

	var result string
	err := s.db.QueryRow("SELECT url FROM url WHERE alias = ?", alias).Scan(&result)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("%s: %w", op, storage.ErrURLNotFound)
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return result, nil
}

// DeleteURL — удалить alias (на будущее)
func (s *Storage) DeleteURL(alias string) error {
	const op = "storage.sqlite.DeleteURL"

	res, err := s.db.Exec("DELETE FROM url WHERE alias = ?", alias)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rows, err := res.RowsAffected()
	if err == nil && rows == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrURLNotFound)
	}

	return nil
}

// Close — корректное закрытие базы
func (s *Storage) Close() error {
	return s.db.Close()
}
