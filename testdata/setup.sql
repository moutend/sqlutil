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

INSERT INTO book (title, author, price, published_at) VALUES ('Game of Thresholds - Episode 1', 'Alice', 10.99, '2019-01-01');
INSERT INTO book (title, author, price, published_at) VALUES ('Game of Thresholds - Episode 2', 'Alice', 20.99, '2019-02-01');
INSERT INTO book (title, author, price, published_at) VALUES ('Game of Thresholds - Episode 3', 'Alice', 30.99, '2019-03-01');
INSERT INTO book (title, author, price, published_at) VALUES ('Game of Thresholds - Episode 4', 'Alice', 40.99, '2019-04-01');
INSERT INTO book (title, author, price, published_at) VALUES ('Game of Thresholds - Episode 5', 'Alice', 50.99, '2019-05-01');
