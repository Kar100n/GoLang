package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Contact string `json:"contact"`
	Role    string `json:"role"`
	LibID   string `json:"libID"`
}

type Library struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type BookInventory struct {
	ISBN            string `json:"isbn"`
	LibID           string `json:"libID"`
	Title           string `json:"title"`
	Authors         string `json:"authors"`
	Publisher       string `json:"publisher"`
	Version         string `json:"version"`
	TotalCopies     int    `json:"totalCopies"`
	AvailableCopies int    `json:"availableCopies"`
}

var db *sql.DB
var err error

func main() {
	db, err = sql.Open("sqlite3", "./library.db")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	// Ensure the tables exist
	createUsersTable()
	createLibraryTable()
	defer deleteLibraryTable()
	createBookInventoryTable()
	purgeAllTables()
	createAllTables()

	router := gin.Default()
	router.POST("/users", createUser)
	router.GET("/users/:id", getUser)
	router.PUT("/users/:id", updateUser)
	router.DELETE("/users/:id", deleteUser)
	router.GET("/users", listUsers)

	router.POST("/books", createBook)
	router.GET("/books/:isbn", getBook)
	router.PUT("/books/:isbn", updateBook)
	router.DELETE("/books/:isbn", deleteBook)
	router.GET("/books", listBooks)

	router.Run(":8081")
}

func createUsersTable() {
	createUsersTableSQL := `CREATE TABLE IF NOT EXISTS users (
        "ID" INTEGER PRIMARY KEY AUTOINCREMENT,
        "Name" TEXT,
        "Email" TEXT,
        "Contact" TEXT,
        "Role" TEXT,
        "LibID" TEXT UNIQUE
    );`

	_, err := db.Exec(createUsersTableSQL)
	if err != nil {
		fmt.Println(err)
	}
}

func createBookInventoryTable() {
	createBookInventoryTableSQL := `CREATE TABLE IF NOT EXISTS book_inventory (
        "ISBN" TEXT PRIMARY KEY,
        "LibID" TEXT,
        "Title" TEXT,
        "Authors" TEXT,
        "Publisher" TEXT,
        "Version" TEXT,
        "TotalCopies" INTEGER,
        "AvailableCopies" INTEGER,
        FOREIGN KEY ("LibID") REFERENCES users("LibID")
    );`

	_, err := db.Exec(createBookInventoryTableSQL)
	if err != nil {
		fmt.Println(err)
	}
}

func createUser(c *gin.Context) {
	var newUser User

	if err := c.BindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if newUser.Role != "Admin" && newUser.Role != "Reader" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role must be either 'Admin' or 'Reader' (Case Sensitive)"})
		return
	}

	statement, _ := db.Prepare("INSERT INTO users (Name, Email, Contact, Role, LibID) VALUES (?,?,?,?,?)")
	result, _ := statement.Exec(newUser.Name, newUser.Email, newUser.Contact, newUser.Role, newUser.LibID)
	id, _ := result.LastInsertId()

	newUser.ID = int(id)
	c.JSON(http.StatusCreated, newUser)
}

func getUser(c *gin.Context) {
	id := c.Param("id")
	var user User

	row := db.QueryRow("SELECT * FROM users WHERE ID =?", id)
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Contact, &user.Role, &user.LibID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func updateUser(c *gin.Context) {
	id := c.Param("id")
	var user User

	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.Exec("UPDATE users SET Name =?, Email =?, Contact =?, Role =?, LibID =? WHERE ID =?", user.Name, user.Email, user.Contact, user.Role, user.LibID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user.ID, _ = strconv.Atoi(id)
	c.JSON(http.StatusOK, user)
}

func deleteUser(c *gin.Context) {
	id := c.Param("id")

	_, err := db.Exec("DELETE FROM users WHERE ID =?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}

func listUsers(c *gin.Context) {
	var users []User

	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Contact, &user.Role, &user.LibID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		users = append(users, user)
	}

	c.JSON(http.StatusOK, users)
}

// BookInventory Creation
func createBook(c *gin.Context) {
	var newBook BookInventory

	if err := c.BindJSON(&newBook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	statement, _ := db.Prepare("INSERT INTO book_inventory (ISBN, LibID, Title, Authors, Publisher, Version, TotalCopies, AvailableCopies) VALUES (?,?,?,?,?,?,?,?)")
	_, _ = statement.Exec(newBook.ISBN, newBook.LibID, newBook.Title, newBook.Authors, newBook.Publisher, newBook.Version, newBook.TotalCopies, newBook.AvailableCopies)
	c.JSON(http.StatusCreated, newBook)
}

func getBook(c *gin.Context) {
	isbn := c.Param("isbn")
	var book BookInventory

	row := db.QueryRow("SELECT * FROM book_inventory WHERE ISBN =?", isbn)
	err := row.Scan(&book.ISBN, &book.LibID, &book.Title, &book.Authors, &book.Publisher, &book.Version, &book.TotalCopies, &book.AvailableCopies)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, book)
}

func updateBook(c *gin.Context) {
	isbn := c.Param("isbn")
	var book BookInventory

	if err := c.BindJSON(&book); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.Exec("UPDATE book_inventory SET LibID =?, Title =?, Authors =?, Publisher =?, Version =?, TotalCopies =?, AvailableCopies =? WHERE ISBN =?", book.LibID, book.Title, book.Authors, book.Publisher, book.Version, book.TotalCopies, book.AvailableCopies, isbn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, book)
}

func deleteBook(c *gin.Context) {
	isbn := c.Param("isbn")

	_, err := db.Exec("DELETE FROM book_inventory WHERE ISBN =?", isbn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Book deleted"})
}

func listBooks(c *gin.Context) {
	var books []BookInventory

	rows, err := db.Query("SELECT * FROM book_inventory")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var book BookInventory
		if err := rows.Scan(&book.ISBN, &book.LibID, &book.Title, &book.Authors, &book.Publisher, &book.Version, &book.TotalCopies, &book.AvailableCopies); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		books = append(books, book)
	}

	c.JSON(http.StatusOK, books)
}

func createLibraryTable() {
	createLibraryTableSQL := `CREATE TABLE IF NOT EXISTS library (
        "ID" INTEGER PRIMARY KEY AUTOINCREMENT,
        "Name" TEXT
    );`

	_, err := db.Exec(createLibraryTableSQL)
	if err != nil {
		fmt.Println(err)
	}
}

func deleteLibraryTable() {
	deleteLibraryTableSQL := `DROP TABLE IF EXISTS library;`

	_, err := db.Exec(deleteLibraryTableSQL)
	if err != nil {
		fmt.Println(err)
	}
}

func purgeAllTables() {
	tables := []string{"users", "books", "library"}

	for _, table := range tables {
		deleteTableSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s;", table)

		_, err := db.Exec(deleteTableSQL)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func createAllTables() {
	createUsersTable()
	createBookInventoryTable()
	createLibraryTable()
}
