package sqlutil

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
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

func migrationUp(db *sql.DB) error {
	data, err := ioutil.ReadFile(filepath.Join("testdata", "0001_unittest.up.sql"))
	if err != nil {
		return err
	}
	if _, err := db.Exec(string(data)); err != nil {
		return err
	}

	return nil
}

func migrationDown(db *sql.DB) error {
	data, err := ioutil.ReadFile(filepath.Join("testdata", "0001_unittest.down.sql"))
	if err != nil {
		return err
	}
	if _, err := db.Exec(string(data)); err != nil {
		return err
	}

	return nil
}

func insertValues(tx *sql.Tx) error {
	for i := 0; i < 10000; i++ {
		query := `INSERT INTO book (title, author, price, published_at) VALUES ($1, $2, $3, $4)`
		if _, err := tx.Exec(
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

func setupData(tx *sql.Tx) error {
	if err := insertValues(tx); err != nil {
		return err
	}

	return nil
}

func TestMain(m *testing.M) {
	code := 0
	defer func() {
		os.Exit(code)
	}()

	db, err := setupDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := migrationUp(db); err != nil {
		log.Fatal(err)
	}

	code = m.Run()

	if err := migrationDown(db); err != nil {
		log.Fatal(err)
	}
}

func TestBind_invalid_builtin(t *testing.T) {
	db, err := setupDB()
	assert.NoError(t, err)

	if err != nil {
		return
	}
	defer db.Close()

	tx, err := db.Begin()
	assert.NoError(t, err)

	if err != nil {
		return
	}
	defer tx.Rollback()

	assert.NoError(t, setupData(tx))

	query := `SELECT price FROM book ORDER BY id LIMIT 1`
	rows, err := tx.Query(query)
	assert.Nil(t, err)
	defer rows.Close()

	var f64 float64

	err = Bind(rows, f64)
	assert.NotNil(t, err, "Because not a pointer, it should fail.")
}

func TestBind_invalid_builtin_slice(t *testing.T) {
	db, err := setupDB()
	assert.NoError(t, err)

	if err != nil {
		return
	}
	defer db.Close()

	tx, err := db.Begin()
	assert.NoError(t, err)

	if err != nil {
		return
	}
	defer tx.Rollback()

	assert.NoError(t, setupData(tx))

	query := `SELECT price FROM book ORDER BY id LIMIT 10`
	rows, err := tx.Query(query)
	assert.NoError(t, err)
	defer rows.Close()

	var f64s []float64

	err = Bind(rows, f64s)
	assert.NotNil(t, err, "Because not a pointer, it should fail.")
}

func TestBind_invalid_struct(t *testing.T) {
	db, err := setupDB()
	assert.NoError(t, err)

	if err != nil {
		return
	}
	defer db.Close()

	tx, err := db.Begin()
	assert.NoError(t, err)

	if err != nil {
		return
	}
	defer tx.Rollback()

	assert.NoError(t, setupData(tx))

	query := `SELECT * FROM book ORDER BY id LIMIT 1`
	rows, err := tx.Query(query)
	assert.NoError(t, err)
	defer rows.Close()

	var book Book

	err = Bind(rows, book)
	assert.NotNil(t, err, "Because not a pointer, it should fail.")
}

func TestBind_invalid_struct_slice(t *testing.T) {
	db, err := setupDB()
	assert.NoError(t, err)

	if err != nil {
		return
	}
	defer db.Close()

	tx, err := db.Begin()
	assert.NoError(t, err)

	if err != nil {
		return
	}
	defer tx.Rollback()

	assert.NoError(t, setupData(tx))

	query := `SELECT * FROM book ORDER BY id LIMIT 10`
	rows, err := tx.Query(query)
	assert.NoError(t, err)
	defer rows.Close()

	var books []Book

	err = Bind(rows, books)
	assert.NotNil(t, err, "because not a pointer")
}

func TestBind_builtin(t *testing.T) {
	db, err := setupDB()
	assert.NoError(t, err)

	if err != nil {
		return
	}
	defer db.Close()

	tx, err := db.Begin()
	assert.NoError(t, err)

	if err != nil {
		return
	}
	defer tx.Rollback()

	assert.NoError(t, setupData(tx))

	query := `SELECT title FROM book ORDER BY id LIMIT 1`
	rows, err := tx.Query(query)
	assert.NoError(t, err)
	defer rows.Close()

	var title string

	assert.NoError(t, Bind(rows, &title))
	assert.Equal(t, "Game of Thresholds - Episode 1", title)
}

func TestBind_builtin_slice(t *testing.T) {
	db, err := setupDB()
	assert.NoError(t, err)

	if err != nil {
		return
	}
	defer db.Close()

	tx, err := db.Begin()
	assert.NoError(t, err)

	if err != nil {
		return
	}
	defer tx.Rollback()

	assert.NoError(t, setupData(tx))

	query := `SELECT title FROM book ORDER BY id LIMIT 10`
	rows, err := tx.Query(query)
	assert.NoError(t, err)
	defer rows.Close()

	var titles []string

	assert.NoError(t, Bind(rows, &titles))
	assert.Equal(t, 10, len(titles))

	for i, title := range titles {
		assert.Equal(t, fmt.Sprintf("Game of Thresholds - Episode %d", i+1), title)
	}
}

func TestBind_struct(t *testing.T) {
	db, err := setupDB()
	assert.NoError(t, err)

	if err != nil {
		return
	}
	defer db.Close()

	tx, err := db.Begin()
	assert.NoError(t, err)

	if err != nil {
		return
	}
	defer tx.Rollback()

	assert.NoError(t, setupData(tx))

	query := `SELECT * FROM book ORDER BY id LIMIT 1`
	rows, err := tx.Query(query)
	assert.NoError(t, err)
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
	assert.NoError(t, err)

	if err != nil {
		return
	}
	defer db.Close()

	tx, err := db.Begin()
	assert.NoError(t, err)

	if err != nil {
		return
	}
	defer tx.Rollback()

	assert.NoError(t, setupData(tx))

	query := `SELECT * FROM book ORDER BY id LIMIT 10`
	rows, err := tx.Query(query)
	assert.NoError(t, err)
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

func TestBind_scanner_slice(t *testing.T) {
	db, err := setupDB()
	assert.NoError(t, err)

	if err != nil {
		return
	}
	defer db.Close()

	tx, err := db.Begin()
	assert.NoError(t, err)

	if err != nil {
		return
	}
	defer tx.Rollback()

	assert.NoError(t, setupData(tx))

	query := `SELECT author, price FROM book ORDER BY id`
	rows, err := tx.Query(query)
	assert.NoError(t, err)
	defer rows.Close()

	var books []struct {
		Author string
		Price  types.Decimal
	}

	assert.NoError(t, Bind(rows, &books))
	assert.Equal(t, 10000, len(books))

	for i, book := range books {
		assert.Equal(t, "Alice", book.Author)
		assert.NotNil(t, book.Price.Big)
		assert.Equal(t, 2, book.Price.Big.Scale())

		f64, ok := book.Price.Big.Float64()
		assert.True(t, ok)
		assert.Equal(t, float64(i+1)*10+0.99, f64)
	}
}

func TestBind_nil_slice(t *testing.T) {
	db, err := setupDB()
	assert.NoError(t, err)

	if err != nil {
		return
	}
	defer db.Close()

	tx, err := db.Begin()
	assert.NoError(t, err)

	if err != nil {
		return
	}
	defer tx.Rollback()

	var books []Book

	query := `SELECT * FROM book ORDER BY id`
	rows, err := tx.Query(query)
	assert.NoError(t, err)
	defer rows.Close()

	assert.NoError(t, Bind(rows, &books))
	assert.Nil(t, books)
	assert.Equal(t, 0, len(books))
}

func TestBind_zero_length_slice(t *testing.T) {
	db, err := setupDB()
	assert.NoError(t, err)

	if err != nil {
		return
	}
	defer db.Close()

	tx, err := db.Begin()
	assert.NoError(t, err)

	if err != nil {
		return
	}
	defer tx.Rollback()

	books := make([]Book, 0)

	query := `SELECT * FROM book ORDER BY id`
	rows, err := tx.Query(query)
	assert.NoError(t, err)
	defer rows.Close()

	assert.NoError(t, Bind(rows, &books))
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

	tx, err := db.Begin()
	assert.NoError(t, err)

	if err != nil {
		return
	}
	defer tx.Rollback()

	assert.NoError(t, setupData(tx))

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
	rows, err := tx.Query(query)
	assert.NoError(t, err)
	defer rows.Close()

	assert.NoError(t, Bind(rows, &books))
	assert.Equal(t, 1, len(books))

	for _, book := range books {
		assert.Equal(t, "Alice", book.Author)
		assert.Equal(t, 10000, book.NumberOfBooks)
	}
}
