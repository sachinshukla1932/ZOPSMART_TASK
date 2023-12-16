package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
	"gofr.dev/pkg/gofr"
)

var db *sql.DB

type Book struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Author   string `json:"author"`
	Quantity int    `json:"quantity"`
}

func createDatabase() {
	var err error
	db, err = sql.Open("sqlite3", "books.db")
	if err != nil {
		fmt.Println("error opening the database")
	}
	createTableQuery := `CREATE TABLE books(id INTEGER PRIMARY KEY AUTOINCREMENT, title VARCHAR(255), author VARCHAR(255), quantity INTEGER);`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		fmt.Println("Error creating database")
	}
}

func addBook(book Book) error {
	_, err := db.Exec("INSERT INTO books(title, author, quantity) VALUES (?, ?, ?)", book.Title, book.Author, book.Quantity)
	return err
}

func viewBooks() ([]Book, error) {
	rows, err := db.Query("SELECT * FROM books")
	if err != nil {
		fmt.Println("Error while executing query:", err)
		return nil, err
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var b Book
		err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Quantity)
		if err != nil {
			fmt.Println("Error while scanning row:", err)
			return nil, err
		}
		books = append(books, b)
	}

	return books, nil
}

func deleteBook(id int) error {
	_, err := db.Exec("DELETE FROM books WHERE id=?", id)
	return err
}

func updateBook(id int, book Book) error {
	_, err := db.Exec("UPDATE books SET title=?, author=?, quantity=? WHERE id=?", book.Title, book.Author, book.Quantity, id)
	return err
}

func main() {
	app := gofr.New()
	createDatabase()

	app.GET("/", func(ctx *gofr.Context) (interface{}, error) {
		return "Welcome to Book Management API", nil
	})

	app.POST("/add", func(ctx *gofr.Context) (interface{}, error) {
		var book Book
		if err := json.NewDecoder(ctx.Request().Body).Decode(&book); err != nil {
			return nil, err
		}
		err := addBook(book)
		if err != nil {
			return nil, err
		}
		return book, nil
	})

	app.GET("/view", func(ctx *gofr.Context) (interface{}, error) {
		books, err := viewBooks()
		if err != nil {
			fmt.Println("Could not view books")
			return nil, err
		}
		return books, nil
	})

	app.GET("/delete/:id", func(ctx *gofr.Context) (interface{}, error) {
		idParam := ctx.Param("id")
		if idParam == "" {
			return nil, fmt.Errorf("ID not provided")
		}

		id, err := strconv.Atoi(idParam)
		if err != nil {
			return nil, fmt.Errorf("invalid format")
		}

		deletedBook, err := viewBooks()
		if err != nil {
			return nil, err
		}

		err = deleteBook(id)
		if err != nil {
			fmt.Println("Couldn't delete book:", err)
			return nil, err
		}

		return deletedBook, nil
	})

	app.PUT("/update/:id", func(ctx *gofr.Context) (interface{}, error) {
		idParam := ctx.Param("id")
		if idParam == "" {
			return nil, fmt.Errorf("ID not provided")
		}

		id, err := strconv.Atoi(idParam)
		if err != nil {
			return nil, fmt.Errorf("invalid format")
		}

		var updatedBook Book
		if err := json.NewDecoder(ctx.Request().Body).Decode(&updatedBook); err != nil {
			return nil, err
		}

		err = updateBook(id, updatedBook)
		if err != nil {
			fmt.Println("Couldn't update book:", err)
			return nil, err
		}

		return updatedBook, nil
	})

	app.Start()
}
