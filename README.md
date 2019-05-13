sqlutil
=======

[![CircleCI](https://circleci.com/gh/moutend/sqlutil/tree/master.svg?style=svg)](https://circleci.com/gh/moutend/sqlutil/tree/master)

Package `sqlutil` provides handy way to bind the `*database/sql.Rows` to the slice or struct.

## Usage

For example, suppose you created a table defined like this:

```sql
CREATE TABLE book (
  id BIGSERIAL NOT NULL,
  title VARCHAR(256) NOT NULL,
  author VARCHAR(128) NOT NULL,
  price DECIMAL NOT NULL,
  published_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY(id)
);
```

You could bind the values into the slice of the struct:

```go
package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/moutend/sqlutil"
)

type Book struct {
	ID          int64
	Title       string
	Author      string
	Price       float64
	PublishedAt time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func main() {
	source := "user=testuser dbname=testdb sslmode=disable password=xxxxxxxx"
	db, err := sql.Open("postgres", source)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	query := `SELECT * FROM book ORDER BY published_at LIMIT 10`
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var books []Book

	if err := sqlutil.Bind(rows, &books); err != nil {
		log.Fatal(err)
	}

	for _, book := range books {
		fmt.Printf("%+v\n", book)
	}
}
```

## LICENSE

MIT

## Author

[Yoshiyuki Koyanagi <moutend@gmail.com>](https://github.com/moutend)
