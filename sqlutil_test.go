package sqlutil

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
)

func setupDB() (*sql.DB, error) {
	source := fmt.Sprintf(
		"user=%s dbname=%s password=%s sslmode=disable",
		os.Getenv("TEST_DATABASE_USERNAME"),
		os.Getenv("TEST_DATABASE_DBNAME"),
		os.Getenv("TEST_DATABASE_PASSWORD"),
	)
	db, err := sql.Open("postgres", source)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open '%s'", source)
	}
	if err := db.Ping(); err != nil {
		return nil, errors.Wrapf(err, "failed to ping '%s'", source)
	}

	return db, nil
}

func setupData(db *sql.DB) error {
	query := `
CREATE TABLE book (
  id BIGSERIAL NOT NULL,
  title VARCHAR(256) NOT NULL,
  author VARCHAR(128) NOT NULL,
  price DECIMAL NOT NULL,
  published_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY(id)
)
`
	if _, err := db.Exec(query); err != nil {
		return err
	}
	for i := 0; i < 10000; i++ {
		query := `INSERT INTO book (title, author, price, published_at) VALUES ($1, $2, $3, $4)`
		if _, err := db.Exec(
			query,
			fmt.Sprintf("Game of Thresholds - Episode %d", i+1),
			"Alice",
			float64(i+1)*10+0.99,
			time.Now().UTC().Format(time.RFC3339),
		); err != nil {
			return err
		}
	}

	return nil
}

func teardownData(db *sql.DB) {
	query := `DROP TABLE book;`
	if _, err := db.Exec(query); err != nil {
		panic(err)
	}
}

func TestBind(t *testing.T) {
	db, err := setupDB()
	assert.Nil(t, err)
	defer db.Close()

	assert.Nil(t, setupData(db))
	defer teardownData(db)
	type Book struct {
		ID          int64
		Title       string
		Author      string
		Price       float64
		PublishedAt time.Time

		CreatedAt time.Time
		UpdatedAt time.Time
	}

	var books []Book

	query := `SELECT * FROM book ORDER BY id LIMIT $1`
	rows, err := db.Query(query, 3)
	assert.Nil(t, err)
	defer rows.Close()

	assert.Nil(t, Bind(rows, &books))
	assert.Equal(t, 3, len(books))

	for i, book := range books {
		assert.Equal(t, fmt.Sprintf("Game of Thresholds - Episode %d", i+1), book.Title)
		assert.Equal(t, "Alice", book.Author)
		assert.Equal(t, 0.99+10*float64(i+1), book.Price)
		assert.True(t, !book.PublishedAt.IsZero())
		assert.True(t, !book.CreatedAt.IsZero())
		assert.True(t, !book.UpdatedAt.IsZero())
	}
}
