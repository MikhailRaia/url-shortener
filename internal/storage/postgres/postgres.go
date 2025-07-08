package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/MikhailRaia/url-shortener/internal/storage"

	"github.com/MikhailRaia/url-shortener/internal/generator"
	"github.com/MikhailRaia/url-shortener/internal/model"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
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

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	storage := &Storage{
		pool: pool,
	}

	if err := storage.createTable(ctx); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *Storage) createTable(ctx context.Context) error {
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS urls (
			id VARCHAR(12) PRIMARY KEY,
			original_url TEXT NOT NULL,
			user_id VARCHAR(32),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
	`

	if _, err := s.pool.Exec(ctx, createTableQuery); err != nil {
		return err
	}

	createIndexQuery := `
		CREATE INDEX IF NOT EXISTS idx_urls_id ON urls(id);
	`

	if _, err := s.pool.Exec(ctx, createIndexQuery); err != nil {
		return err
	}

	createUniqueIndexQuery := `
		CREATE UNIQUE INDEX IF NOT EXISTS idx_urls_original_url ON urls(original_url);
	`

	_, err := s.pool.Exec(ctx, createUniqueIndexQuery)
	return err
}

func (s *Storage) Save(originalURL string) (string, error) {
	ctx := context.Background()

	var existingID string
	err := s.pool.QueryRow(ctx, "SELECT id FROM urls WHERE original_url = $1", originalURL).Scan(&existingID)
	if err == nil {
		return existingID, storage.ErrURLExists
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("error checking if URL exists: %w", err)
	}

	id, err := generator.GenerateID(8)
	if err != nil {
		return "", fmt.Errorf("error generating ID: %w", err)
	}

	var exists bool
	for {
		err := s.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM urls WHERE id = $1)", id).Scan(&exists)
		if err != nil {
			return "", fmt.Errorf("error checking if ID exists: %w", err)
		}

		if !exists {
			break
		}

		id, err = generator.GenerateID(8)
		if err != nil {
			return "", fmt.Errorf("error generating new ID: %w", err)
		}
	}

	_, err = s.pool.Exec(ctx, "INSERT INTO urls (id, original_url) VALUES ($1, $2)", id, originalURL)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			if err := s.pool.QueryRow(ctx, "SELECT id FROM urls WHERE original_url = $1", originalURL).Scan(&existingID); err == nil {
				return existingID, storage.ErrURLExists
			}
		}
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
			return "", false
		}
		fmt.Printf("Error querying database: %v\n", err)
		return "", false
	}

	return originalURL, true
}

func (s *Storage) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

func (s *Storage) SaveBatch(items []model.BatchRequestItem) (map[string]string, error) {
	ctx := context.Background()
	result := make(map[string]string)

	for _, item := range items {
		var existingID string
		err := s.pool.QueryRow(ctx, "SELECT id FROM urls WHERE original_url = $1", item.OriginalURL).Scan(&existingID)
		if err == nil {
			result[item.CorrelationID] = existingID
			continue
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("error checking if URL exists: %w", err)
		}

		id, err := generator.GenerateID(8)
		if err != nil {
			return nil, fmt.Errorf("error generating ID: %w", err)
		}

		var exists bool
		for {
			err := s.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM urls WHERE id = $1)", id).Scan(&exists)
			if err != nil {
				return nil, fmt.Errorf("error checking if ID exists: %w", err)
			}

			if !exists {
				break
			}

			id, err = generator.GenerateID(8)
			if err != nil {
				return nil, fmt.Errorf("error generating new ID: %w", err)
			}
		}

		_, err = s.pool.Exec(ctx, "INSERT INTO urls (id, original_url) VALUES ($1, $2)",
			id, item.OriginalURL)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				if err := s.pool.QueryRow(ctx, "SELECT id FROM urls WHERE original_url = $1", item.OriginalURL).Scan(&existingID); err == nil {
					result[item.CorrelationID] = existingID
					continue
				}
			}
			return nil, fmt.Errorf("error inserting URL into database: %w", err)
		}

		result[item.CorrelationID] = id
	}

	return result, nil
}

func (s *Storage) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

func (s *Storage) SaveWithUser(originalURL, userID string) (string, error) {
	ctx := context.Background()

	var existingID string
	err := s.pool.QueryRow(ctx, "SELECT id FROM urls WHERE original_url = $1", originalURL).Scan(&existingID)
	if err == nil {
		return existingID, storage.ErrURLExists
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("error checking if URL exists: %w", err)
	}

	id, err := generator.GenerateID(8)
	if err != nil {
		return "", fmt.Errorf("error generating ID: %w", err)
	}

	var exists bool
	for {
		err := s.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM urls WHERE id = $1)", id).Scan(&exists)
		if err != nil {
			return "", fmt.Errorf("error checking if ID exists: %w", err)
		}

		if !exists {
			break
		}

		id, err = generator.GenerateID(8)
		if err != nil {
			return "", fmt.Errorf("error generating new ID: %w", err)
		}
	}

	_, err = s.pool.Exec(ctx, "INSERT INTO urls (id, original_url, user_id) VALUES ($1, $2, $3)", id, originalURL, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			if err := s.pool.QueryRow(ctx, "SELECT id FROM urls WHERE original_url = $1", originalURL).Scan(&existingID); err == nil {
				return existingID, storage.ErrURLExists
			}
		}
		return "", fmt.Errorf("error inserting URL into database: %w", err)
	}

	return id, nil
}

func (s *Storage) SaveBatchWithUser(items []model.BatchRequestItem, userID string) (map[string]string, error) {
	ctx := context.Background()
	result := make(map[string]string)

	for _, item := range items {
		var existingID string
		err := s.pool.QueryRow(ctx, "SELECT id FROM urls WHERE original_url = $1", item.OriginalURL).Scan(&existingID)
		if err == nil {
			result[item.CorrelationID] = existingID
			continue
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("error checking if URL exists: %w", err)
		}

		id, err := generator.GenerateID(8)
		if err != nil {
			return nil, fmt.Errorf("error generating ID: %w", err)
		}

		var exists bool
		for {
			err := s.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM urls WHERE id = $1)", id).Scan(&exists)
			if err != nil {
				return nil, fmt.Errorf("error checking if ID exists: %w", err)
			}

			if !exists {
				break
			}

			id, err = generator.GenerateID(8)
			if err != nil {
				return nil, fmt.Errorf("error generating new ID: %w", err)
			}
		}

		_, err = s.pool.Exec(ctx, "INSERT INTO urls (id, original_url, user_id) VALUES ($1, $2, $3)",
			id, item.OriginalURL, userID)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				if err := s.pool.QueryRow(ctx, "SELECT id FROM urls WHERE original_url = $1", item.OriginalURL).Scan(&existingID); err == nil {
					result[item.CorrelationID] = existingID
					continue
				}
			}
			return nil, fmt.Errorf("error inserting URL into database: %w", err)
		}

		result[item.CorrelationID] = id
	}

	return result, nil
}

func (s *Storage) GetUserURLs(userID string) ([]model.UserURL, error) {
	ctx := context.Background()

	rows, err := s.pool.Query(ctx, "SELECT id, original_url FROM urls WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("error querying user URLs: %w", err)
	}
	defer rows.Close()

	var result []model.UserURL
	for rows.Next() {
		var id, originalURL string
		if err := rows.Scan(&id, &originalURL); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		result = append(result, model.UserURL{
			ShortURL:    id,
			OriginalURL: originalURL,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return result, nil
}
