package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/MikhailRaia/url-shortener/internal/generator"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
}

func NewStorage(dsn string) (*Storage, error) {
	if dsn == "" {
		return nil, errors.New("database connection string is empty")
	}

	ctx := context.Background()
	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}

	// Проверяем соединение
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	storage := &Storage{
		pool: pool,
	}

	// Создаем таблицу, если она не существует
	if err := storage.createTable(ctx); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *Storage) createTable(ctx context.Context) error {
	// Создаем таблицу, если она не существует
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS urls (
			id VARCHAR(12) PRIMARY KEY,
			original_url TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
	`

	if _, err := s.pool.Exec(ctx, createTableQuery); err != nil {
		return err
	}

	// Убедимся, что у нас есть индекс для быстрого поиска по id
	// (хотя PRIMARY KEY уже создает индекс, но явно указываем для полноты)
	createIndexQuery := `
		CREATE INDEX IF NOT EXISTS idx_urls_id ON urls(id);
	`

	_, err := s.pool.Exec(ctx, createIndexQuery)
	return err
}

func (s *Storage) Save(originalURL string) (string, error) {
	ctx := context.Background()

	// Генерируем ID для сокращенного URL
	id, err := generator.GenerateID(8)
	if err != nil {
		return "", fmt.Errorf("error generating ID: %w", err)
	}

	// Проверяем, существует ли URL с таким ID
	var exists bool
	for {
		err := s.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM urls WHERE id = $1)", id).Scan(&exists)
		if err != nil {
			return "", fmt.Errorf("error checking if ID exists: %w", err)
		}

		if !exists {
			break
		}

		// Если ID уже существует, генерируем новый
		id, err = generator.GenerateID(8)
		if err != nil {
			return "", fmt.Errorf("error generating new ID: %w", err)
		}
	}

	// Сохраняем URL в базу данных
	_, err = s.pool.Exec(ctx, "INSERT INTO urls (id, original_url) VALUES ($1, $2)", id, originalURL)
	if err != nil {
		return "", fmt.Errorf("error inserting URL into database: %w", err)
	}

	return id, nil
}

func (s *Storage) Get(id string) (string, bool) {
	ctx := context.Background()

	var originalURL string
	err := s.pool.QueryRow(ctx, "SELECT original_url FROM urls WHERE id = $1", id).Scan(&originalURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// URL не найден
			return "", false
		}
		// Произошла ошибка при выполнении запроса
		fmt.Printf("Error querying database: %v\n", err)
		return "", false
	}

	return originalURL, true
}

func (s *Storage) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

func (s *Storage) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}
