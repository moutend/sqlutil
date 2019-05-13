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
	"github.com/volatiletech/sqlboiler/types"
)

type Book struct {
	ID          int64
	Title       string
	Author      string
	Price       float64
	PublishedAt time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

func setupDB() (*sql.DB, error) {
	source := os.Getenv("DATABASE_URI")
	db, err := sql.Open("postgres", source)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open '%s'", source)
	}
	if err := db.Ping(); err != nil {
		return nil, errors.Wrapf(err, "failed to ping '%s'", source)
	}

	return db, nil
}

func createTables(db *sql.DB) error {
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

	return nil
}

func insertValues(db *sql.DB) error {
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

func setupData(db *sql.DB) error {
	if err := createTables(db); err != nil {
		return err
	}
	if err := insertValues(db); err != nil {
		return err
	}

	return nil
}

func teardownData(db *sql.DB) {
	query := `DROP TABLE book;`
	if _, err := db.Exec(query); err != nil {
		panic(err)
	}
}

func TestBind_invalid_builtin(t *testing.T) {
	db, err := setupDB()
	assert.Nil(t, err)

	if err != nil {
		return
	}
	defer db.Close()

	assert.Nil(t, setupData(db))
	defer teardownData(db)

	query := `SELECT price FROM book ORDER BY id LIMIT 1`
	rows, err := db.Query(query)
	assert.Nil(t, err)
	defer rows.Close()

	var f64 float64

	err = Bind(rows, f64)
	assert.NotNil(t, err, "Because not a pointer, it should fail.")
}

func TestBind_invalid_builtin_slice(t *testing.T) {
	db, err := setupDB()
	assert.Nil(t, err)

	if err != nil {
		return
	}
	defer db.Close()

	assert.Nil(t, setupData(db))
	defer teardownData(db)

	query := `SELECT price FROM book ORDER BY id LIMIT 10`
	rows, err := db.Query(query)
	assert.Nil(t, err)
	defer rows.Close()

	var f64s []float64

	err = Bind(rows, f64s)
	assert.NotNil(t, err, "Because not a pointer, it should fail.")
}

func TestBind_invalid_struct(t *testing.T) {
	db, err := setupDB()
	assert.Nil(t, err)

	if err != nil {
		return
	}
	defer db.Close()

	assert.Nil(t, setupData(db))
	defer teardownData(db)

	query := `SELECT * FROM book ORDER BY id LIMIT 1`
	rows, err := db.Query(query)
	assert.Nil(t, err)
	defer rows.Close()

	var book Book

	err = Bind(rows, book)
	assert.NotNil(t, err, "Because not a pointer, it should fail.")
}

func TestBind_invalid_struct_slice(t *testing.T) {
	db, err := setupDB()
	assert.Nil(t, err)

	if err != nil {
		return
	}
	defer db.Close()

	assert.Nil(t, setupData(db))
	defer teardownData(db)

	query := `SELECT * FROM book ORDER BY id LIMIT 10`
	rows, err := db.Query(query)
	assert.Nil(t, err)
	defer rows.Close()

	var books []Book

	err = Bind(rows, books)
	assert.NotNil(t, err)
}

func TestBind_builtin(t *testing.T) {
	db, err := setupDB()
	assert.Nil(t, err)

	if err != nil {
		return
	}
	defer db.Close()

	assert.Nil(t, setupData(db))
	defer teardownData(db)

	query := `SELECT title FROM book ORDER BY id LIMIT 1`
	rows, err := db.Query(query)
	assert.Nil(t, err)
	defer rows.Close()

	var title string

	assert.Nil(t, Bind(rows, &title))
	assert.Equal(t, "Game of Thresholds - Episode 1", title)
}

/*
func TestBind_builtin_slice(t *testing.T) {
	db, err := setupDB()
	assert.Nil(t, err)

	if err != nil {
		return
	}
	defer db.Close()

	assert.Nil(t, setupData(db))
	defer teardownData(db)

	query := `SELECT title FROM book ORDER BY id LIMIT 10`
	rows, err := db.Query(query)
	assert.Nil(t, err)
	defer rows.Close()

	var titles []string

	assert.Nil(t, Bind(rows, &titles))
	assert.Equal(t, 10, len(titles))

	for i, title := range titles {
		assert.Equal(t, fmt.Sprintf("Game of Thresholds - Episode %d", i+1), title)
	}
}
*/

func TestBind_struct(t *testing.T) {
	db, err := setupDB()
	assert.Nil(t, err)

	if err != nil {
		return
	}
	defer db.Close()

	assert.Nil(t, setupData(db))
	defer teardownData(db)

	query := `SELECT * FROM book ORDER BY id LIMIT 1`
	rows, err := db.Query(query)
	assert.Nil(t, err)
	defer rows.Close()

	var book Book

	assert.Nil(t, Bind(rows, &book))
	assert.Equal(t, "Game of Thresholds - Episode 1", book.Title)
	assert.Equal(t, "Alice", book.Author)
	assert.Equal(t, 10.99, book.Price)
	assert.True(t, !book.PublishedAt.IsZero())
	assert.True(t, !book.CreatedAt.IsZero())
	assert.True(t, !book.UpdatedAt.IsZero())
}

func TestBind_struct_slice(t *testing.T) {
	db, err := setupDB()
	assert.Nil(t, err)

	if err != nil {
		return
	}
	defer db.Close()

	assert.Nil(t, setupData(db))
	defer teardownData(db)

	query := `SELECT * FROM book ORDER BY id LIMIT 10`
	rows, err := db.Query(query)
	assert.Nil(t, err)
	defer rows.Close()

	var books []Book

	assert.Nil(t, Bind(rows, &books))
	assert.Equal(t, 10, len(books))

	for i, book := range books {
		assert.Equal(t, fmt.Sprintf("Game of Thresholds - Episode %d", i+1), book.Title)
		assert.Equal(t, "Alice", book.Author)
		assert.Equal(t, 0.99+10*float64(i+1), book.Price)
		assert.True(t, !book.PublishedAt.IsZero())
		assert.True(t, !book.CreatedAt.IsZero())
		assert.True(t, !book.UpdatedAt.IsZero())
	}
}

func TestBind_nil_slice(t *testing.T) {
	db, err := setupDB()
	assert.Nil(t, err)

	if err != nil {
		return
	}
	defer db.Close()

	assert.Nil(t, createTables(db))
	defer teardownData(db)

	var books []Book

	query := `SELECT * FROM book ORDER BY id`
	rows, err := db.Query(query)
	assert.Nil(t, err)
	defer rows.Close()

	assert.Nil(t, Bind(rows, &books))
	assert.Nil(t, books)
	assert.Equal(t, 0, len(books))
}

func TestBind_zero_length_slice(t *testing.T) {
	db, err := setupDB()
	assert.Nil(t, err)

	if err != nil {
		return
	}
	defer db.Close()

	assert.Nil(t, createTables(db))
	defer teardownData(db)

	books := make([]Book, 0)

	query := `SELECT * FROM book ORDER BY id`
	rows, err := db.Query(query)
	assert.Nil(t, err)
	defer rows.Close()

	assert.Nil(t, Bind(rows, &books))
	assert.NotNil(t, books)
	assert.Equal(t, 0, len(books))
}

func TestBind_count(t *testing.T) {
	db, err := setupDB()
	assert.Nil(t, err)

	if err != nil {
		return
	}
	defer db.Close()

	assert.Nil(t, setupData(db))
	defer teardownData(db)

	var books []struct {
		Author        string
		NumberOfBooks int
	}

	query := `
SELECT
  author
, COUNT(*) AS number_of_books
FROM book
GROUP BY author
`
	rows, err := db.Query(query)
	assert.Nil(t, err)
	defer rows.Close()

	assert.Nil(t, Bind(rows, &books))
	assert.Equal(t, 1, len(books))

	for _, book := range books {
		assert.Equal(t, "Alice", book.Author)
		assert.Equal(t, 10000, book.NumberOfBooks)
	}
}

func TestBind_scanner(t *testing.T) {
	db, err := setupDB()
	assert.Nil(t, err)

	if err != nil {
		return
	}
	defer db.Close()

	assert.Nil(t, setupData(db))
	defer teardownData(db)

	var books []struct {
		Price types.Decimal
	}

	query := `SELECT price FROM book ORDER BY id`
	rows, err := db.Query(query)
	assert.Nil(t, err)
	defer rows.Close()

	assert.Nil(t, Bind(rows, &books))
	assert.Equal(t, 10000, len(books))

	for i, book := range books {
		assert.NotNil(t, book.Price.Big)
		assert.Equal(t, 2, book.Price.Big.Scale())

		f64, ok := book.Price.Big.Float64()
		assert.True(t, ok)
		assert.Equal(t, float64(i+1)*10+0.99, f64)
	}
}
