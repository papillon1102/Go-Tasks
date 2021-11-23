package main

import (
	"context"

	"os"

	"github.com/gin-gonic/gin"
	handler "github.com/papillon1102/go-tasks/tasksHandler"
	"github.com/phuslu/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var taskHandler *handler.TaskHandler

func init() {

	// Config for "phuslu log"
	if log.IsTerminal(os.Stderr.Fd()) {
		log.DefaultLogger = log.Logger{
			TimeFormat: "15:04:05",
			Caller:     1,
			Writer: &log.ConsoleWriter{
				ColorOutput:    true,
				QuoteString:    true,
				EndWithMessage: true,
			},
		}
	}

	ctx := context.Background()

	// Connect to Mongo via ENV var
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal().Err(err)
	} else {
		log.Info().Msg("Connected to MongoDB")
	}

	// Connect to "Go-Tasks" MongoDB
	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("tasks")

	// Make new task-handler
	taskHandler = handler.NewTasksHandler(ctx, collection)
}

func NewRouter() *gin.Engine {

	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Fuck world",
		})
	})

	router.POST("/task", taskHandler.NewTaskHandler)
	router.GET("/task", taskHandler.ListTaskHandler)
	router.PUT("/task/:id", taskHandler.UpdateTaskHandler)
	return router
}

func main() {
	r := NewRouter()
	r.Run()
}
