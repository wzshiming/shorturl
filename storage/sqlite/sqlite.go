package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"

	"github.com/wzshiming/shorturl"
	"github.com/wzshiming/shorturl/storage"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(index uint64, dataSourceName string) (storage.Storage, error) {
	db, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	s := &Storage{
		db: db,
	}

	err = s.initTable(context.Background(), index)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Storage) initTable(ctx context.Context, index uint64) error {
	// Create the table if it does not exist
	_, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS shorturl (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    origin TEXT UNIQUE
);
`)
	if err != nil {
		return fmt.Errorf("create table: %w", err)
	}

	// Sed up the index as AUTO_INCREMENT with creating the table
	if index != 0 {
		row := s.db.QueryRowContext(ctx, `
SELECT seq FROM sqlite_sequence WHERE name = 'shorturl';
`)
		var seq int64
		err = row.Scan(&seq)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("get sqlite_sequence: %w", err)
			}

			_, err = s.db.ExecContext(ctx, `
INSERT OR IGNORE INTO sqlite_sequence(name, seq) VALUES ('shorturl', ?);
`, index)
			if err != nil {
				return fmt.Errorf("insert sqlite_sequence: %w", err)
			}
		}

		if seq < int64(index) {
			_, err = s.db.ExecContext(ctx, `
UPDATE sqlite_sequence SET seq = ? WHERE name = 'shorturl';
`, index)
			if err != nil {
				return fmt.Errorf("update sqlite_sequence: %w", err)
			}
		}
	}

	return nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) Encode(ctx context.Context, origin string) (index string, err error) {
	var id int64
	// Avoid that an unsuccessful insertion will also increment the id
	err = s.db.QueryRowContext(ctx, `
SELECT id FROM shorturl WHERE origin = ?;
`, origin).Scan(&id)
	if err == nil {
		return shorturl.Encode(uint64(id)), nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}

	result, err := s.db.ExecContext(ctx, `
INSERT OR IGNORE INTO shorturl (origin) VALUES (?);
`, origin)
	if err != nil {
		return "", err
	}

	id, err = result.LastInsertId()
	if err != nil {
		return "", err
	}

	return shorturl.Encode(uint64(id)), nil

}

func (s *Storage) Decode(ctx context.Context, index string) (origin string, err error) {
	id, err := shorturl.Decode(index)
	if err != nil {
		return "", err
	}

	row := s.db.QueryRowContext(ctx, `
SELECT origin FROM shorturl WHERE id = ?;
`, id)
	err = row.Scan(&origin)
	if err != nil {
		return "", err
	}

	return origin, nil
}
