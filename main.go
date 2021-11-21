package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var tasks []Task

type User struct {
	Name  string `json:"name"`
	GGID  string `json:"ggid"`
	Email string `json:"email"`
}

type Task struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Tags      []string  `json:"tags"`
	User      User      `json:"user"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:createdAt`
}

func NewTaskHandler(c *gin.Context) {
	var task Task

	// Decode request body in to "Task" struct
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	task.ID = uuid.NewString()
	task.CreatedAt = time.Now()

	tasks = append(tasks, task)
	c.JSON(http.StatusOK, task)
}

func ListTaskHandler(c *gin.Context) {

	// Encode "tasks" array into JSON
	c.JSON(http.StatusOK, tasks)
}

func NewRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Fuck world",
		})
	})

	router.POST("/task", NewTaskHandler)
	router.GET("/task", ListTaskHandler)
	return router
}

func main() {
	r := NewRouter()
	r.Run()
}
