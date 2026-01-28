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

// Storage implements URLStorage using PostgreSQL.
type Storage struct {
	pool *pgxpool.Pool
}

// NewStorage connects to PostgreSQL using DSN and prepares schema.
func NewStorage(dsn string) (*Storage, error) {
	if dsn == "" {
		return nil, errors.New("database connection string is empty")
	}

	ctx := context.Background()
	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storage := &Storage{
		pool: pool,
	}

	if err := storage.createTable(ctx); err != nil {
		return nil, fmt.Errorf("failed to create database tables: %w", err)
	}

	return storage, nil
}

func (s *Storage) createTable(ctx context.Context) error {
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS urls (
			id VARCHAR(12) PRIMARY KEY,
			original_url TEXT NOT NULL,
			user_id VARCHAR(32),
			is_deleted BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
	`

	if _, err := s.pool.Exec(ctx, createTableQuery); err != nil {
		return fmt.Errorf("failed to create urls table: %w", err)
	}

	alterTableQuery := `
		ALTER TABLE urls ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT FALSE;
	`

	if _, err := s.pool.Exec(ctx, alterTableQuery); err != nil {
		return fmt.Errorf("failed to alter urls table: %w", err)
	}

	createIndexQuery := `
		CREATE INDEX IF NOT EXISTS idx_urls_id ON urls(id);
	`

	if _, err := s.pool.Exec(ctx, createIndexQuery); err != nil {
		return fmt.Errorf("failed to create index on id: %w", err)
	}

	createUniqueIndexQuery := `
		CREATE UNIQUE INDEX IF NOT EXISTS idx_urls_original_url ON urls(original_url);
	`

	if _, err := s.pool.Exec(ctx, createUniqueIndexQuery); err != nil {
		return fmt.Errorf("failed to create unique index on original_url: %w", err)
	}

	return nil
}

// Save stores a new URL and returns its generated short ID.
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

// Get retrieves the original URL for a given short ID.
func (s *Storage) Get(id string) (string, bool) {
	ctx := context.Background()

	var originalURL string
	var isDeleted bool
	err := s.pool.QueryRow(ctx, "SELECT original_url, is_deleted FROM urls WHERE id = $1", id).Scan(&originalURL, &isDeleted)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", false
		}
		fmt.Printf("Error querying database: %v\n", err)
		return "", false
	}

	if isDeleted {
		return "", false
	}

	return originalURL, true
}

// GetWithDeletedStatus retrieves the original URL and checks if it has been deleted.
func (s *Storage) GetWithDeletedStatus(id string) (string, error) {
	ctx := context.Background()

	var originalURL string
	var isDeleted bool
	err := s.pool.QueryRow(ctx, "SELECT original_url, is_deleted FROM urls WHERE id = $1", id).Scan(&originalURL, &isDeleted)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("error querying database: %w", err)
	}

	if isDeleted {
		return "", storage.ErrURLDeleted
	}

	return originalURL, nil
}

// Ping verifies the database connection is alive.
func (s *Storage) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

// SaveBatch stores multiple URLs and returns a map of correlation IDs to short IDs.
func (s *Storage) SaveBatch(items []model.BatchRequestItem) (map[string]string, error) {
	ctx := context.Background()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	result := make(map[string]string)

	for _, item := range items {
		var existingID string
		err := tx.QueryRow(ctx, "SELECT id FROM urls WHERE original_url = $1 FOR UPDATE", item.OriginalURL).Scan(&existingID)
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
			err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM urls WHERE id = $1)", id).Scan(&exists)
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

		_, err = tx.Exec(ctx, "INSERT INTO urls (id, original_url) VALUES ($1, $2)",
			id, item.OriginalURL)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				if err := tx.QueryRow(ctx, "SELECT id FROM urls WHERE original_url = $1", item.OriginalURL).Scan(&existingID); err == nil {
					result[item.CorrelationID] = existingID
					continue
				}
			}
			return nil, fmt.Errorf("error inserting URL into database: %w", err)
		}

		result[item.CorrelationID] = id
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return result, nil
}

// Close releases the underlying connection pool.
func (s *Storage) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

// SaveWithUser stores a new URL associated with a user and returns its generated short ID.
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

// SaveBatchWithUser stores multiple URLs associated with a user and returns a map of correlation IDs to short IDs.
func (s *Storage) SaveBatchWithUser(items []model.BatchRequestItem, userID string) (map[string]string, error) {
	ctx := context.Background()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	result := make(map[string]string)

	for _, item := range items {
		var existingID string
		err := tx.QueryRow(ctx, "SELECT id FROM urls WHERE original_url = $1 FOR UPDATE", item.OriginalURL).Scan(&existingID)
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
			err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM urls WHERE id = $1)", id).Scan(&exists)
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

		_, err = tx.Exec(ctx, "INSERT INTO urls (id, original_url, user_id) VALUES ($1, $2, $3)",
			id, item.OriginalURL, userID)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				if err := tx.QueryRow(ctx, "SELECT id FROM urls WHERE original_url = $1", item.OriginalURL).Scan(&existingID); err == nil {
					result[item.CorrelationID] = existingID
					continue
				}
			}
			return nil, fmt.Errorf("error inserting URL into database: %w", err)
		}

		result[item.CorrelationID] = id
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return result, nil
}

// GetUserURLs retrieves all non-deleted URLs associated with a user.
func (s *Storage) GetUserURLs(userID string) ([]model.UserURL, error) {
	ctx := context.Background()

	rows, err := s.pool.Query(ctx, "SELECT id, original_url FROM urls WHERE user_id = $1 AND is_deleted = FALSE", userID)
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

// GetStats returns total number of URLs and users.
func (s *Storage) GetStats() (int, int, error) {
	ctx := context.Background()

	var urlsCount int
	err := s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM urls").Scan(&urlsCount)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get urls count: %w", err)
	}

	var usersCount int
	err = s.pool.QueryRow(ctx, "SELECT COUNT(DISTINCT user_id) FROM urls WHERE user_id IS NOT NULL AND user_id != ''").Scan(&usersCount)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get users count: %w", err)
	}

	return urlsCount, usersCount, nil
}

// DeleteUserURLs marks specified URLs as deleted for a user.
func (s *Storage) DeleteUserURLs(userID string, urlIDs []string) error {
	if len(urlIDs) == 0 {
		return nil
	}

	ctx := context.Background()

	query := `UPDATE urls SET is_deleted = TRUE WHERE user_id = $1 AND id = ANY($2) AND is_deleted = FALSE`

	_, err := s.pool.Exec(ctx, query, userID, urlIDs)
	if err != nil {
		return fmt.Errorf("error deleting URLs: %w", err)
	}

	return nil
}
