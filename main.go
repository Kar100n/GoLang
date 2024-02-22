package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	DueDate     string `json:"dueDate"`
	Status      string `json:"status"`
}

var db *sql.DB
var err error

func main() {
	db, err = sql.Open("sqlite3", "./tasks.db")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	// Ensure the table exists
	createTable()

	router := gin.Default()
	router.POST("/tasks", createTask)

	router.Run(":8080")
}

func createTable() {
	createTasksTableSQL := `CREATE TABLE IF NOT EXISTS tasks (
        "ID" INTEGER PRIMARY KEY AUTOINCREMENT,
        "Title" TEXT,
        "Description" TEXT,
        "DueDate" TEXT,
        "Status" TEXT
    );`

	_, err := db.Exec(createTasksTableSQL)
	if err != nil {
		fmt.Println(err)
	}
}

func createTask(c *gin.Context) {
	var newTask Task

	if err := c.BindJSON(&newTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	statement, _ := db.Prepare("INSERT INTO tasks (Title, Description, DueDate, Status) VALUES (?, ?, ?, ?)")
	result, _ := statement.Exec(newTask.Title, newTask.Description, newTask.DueDate, "pending")
	id, _ := result.LastInsertId()

	newTask.ID = int(id)
	c.JSON(http.StatusCreated, newTask)
}
