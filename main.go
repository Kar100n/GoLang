package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

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
	router.GET("/tasks/:id", getTask)
	router.PUT("/tasks/:id", updateTask)
	router.DELETE("/tasks/:id", deleteTask)
	router.GET("/tasks", listTasks)

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

func getTask(c *gin.Context) {
	id := c.Param("id")
	var task Task

	row := db.QueryRow("SELECT * FROM tasks WHERE ID = ?", id)
	err := row.Scan(&task.ID, &task.Title, &task.Description, &task.DueDate, &task.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

func updateTask(c *gin.Context) {
	id := c.Param("id")
	var task Task

	if err := c.BindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.Exec("UPDATE tasks SET Title = ?, Description = ?, DueDate = ? WHERE ID = ?", task.Title, task.Description, task.DueDate, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	task.ID, _ = strconv.Atoi(id)
	c.JSON(http.StatusOK, task)
}

func deleteTask(c *gin.Context) {
	id := c.Param("id")

	_, err := db.Exec("DELETE FROM tasks WHERE ID = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted"})
}

func listTasks(c *gin.Context) {
	var tasks []Task

	rows, err := db.Query("SELECT * FROM tasks")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.DueDate, &task.Status); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		tasks = append(tasks, task)
	}

	c.JSON(http.StatusOK, tasks)
}
