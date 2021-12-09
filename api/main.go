package main

import (
	"context"
	"sync"

	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	handler "github.com/papillon1102/Go-Tasks/api/tasksHandler"

	"github.com/phuslu/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Add session management system (FIXME)
// "url": "http://192.168.99.100/api",

var taskHandler *handler.TaskHandler
var authHandler *handler.AuthHandler

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
	mutex := &sync.Mutex{}
	ctx := context.Background()

	// Connect to Mongo via ENV var
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal().Err(err)
	} else {
		log.Info().Msg("Connected to MongoDB")
	}

	// "Tasks-Collection" of MongoDB
	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("tasks")

	// "Users-Collection" of MongoDB
	userCollection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("users")

	// Add redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "192.168.99.100:6379",
		Password: "",
		DB:       0,
	})

	// Make new task-handler
	taskHandler = handler.NewTasksHandler(ctx, collection, redisClient)
	authHandler = handler.NewAuthHandler(userCollection, ctx, redisClient, "", mutex)

	status := redisClient.Ping()
	log.Info().Msgf("Status: %v", status)
}

func NewRouter() *gin.Engine {

	router := gin.Default()

	// Setup CORS handle
	router.Use(cors.Default())

	// Call another router later (NOTE)
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Fuck world",
		})
	})

	router.GET("/task", taskHandler.ListTaskHandler)
	router.PUT("/task/:id", taskHandler.UpdateTaskHandler)

	// router.POST("/signin", authHandler.SignInHandler)
	router.POST("/signin", authHandler.SignInPWTHandler)
	router.POST("/refresh", authHandler.RefreshHandler)
	router.POST("/signup", authHandler.SignUpPWTHandler)

	// Create new router group
	auth := router.Group("/")

	// Add middleware to the group
	auth.Use(authHandler.AuthMiddleware())

	// Add middleware to router
	{
		auth.POST("/task", taskHandler.NewTaskHandler)

	}
	return router
}

func main() {
	r := NewRouter()
	r.Run(":8081")
}
