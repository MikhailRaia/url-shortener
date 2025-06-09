package postgres

import (
	"context"
	"errors"

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
	query := `
		CREATE TABLE IF NOT EXISTS urls (
			id VARCHAR(10) PRIMARY KEY,
			original_url TEXT NOT NULL
		);
	`

	_, err := s.pool.Exec(ctx, query)
	return err
}

func (s *Storage) Save(originalURL string) (string, error) {
	ctx := context.Background()

	// Генерируем ID для сокращенного URL
	id, _ := generator.GenerateID(6)

	// Проверяем, существует ли URL с таким ID
	var exists bool
	for {
		err := s.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM urls WHERE id = $1)", id).Scan(&exists)
		if err != nil {
			return "", err
		}

		if !exists {
			break
		}

		// Если ID уже существует, генерируем новый
		id, _ = generator.GenerateID(6)
	}

	// Сохраняем URL в базу данных
	_, err := s.pool.Exec(ctx, "INSERT INTO urls (id, original_url) VALUES ($1, $2)", id, originalURL)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (s *Storage) Get(id string) (string, bool) {
	ctx := context.Background()

	var originalURL string
	err := s.pool.QueryRow(ctx, "SELECT original_url FROM urls WHERE id = $1", id).Scan(&originalURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", false
		}
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
