package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/papillon1102/go-tasks/models"
	"github.com/phuslu/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
)

type TaskHandler struct {
	collection *mongo.Collection
	ctx        context.Context
}

func NewTasksHandler(ctx context.Context, collection *mongo.Collection) *TaskHandler {
	return &TaskHandler{
		collection: collection,
		ctx:        ctx,
	}
}

// Handler to [ add new task ]
func (th *TaskHandler) NewTaskHandler(c *gin.Context) {
	var task models.Task

	// Decode request body in to "Task" struct
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Create ID and time for task
	task.ID = primitive.NewObjectID()
	task.CreatedAt = time.Now()

	// Insert it into DB
	_, err := th.collection.InsertOne(th.ctx, task)
	if err != nil {
		log.Error().Err(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while inserting new task"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// Handlers to [ list all tasks ]
func (th *TaskHandler) ListTaskHandler(c *gin.Context) {

	cur, err := th.collection.Find(th.ctx, bson.M{})
	if err != nil {
		log.Error().Err(err)
		return
	}
	defer cur.Close(th.ctx)

	var tasks []models.Task
	for cur.Next(th.ctx) {
		task := models.Task{}
		cur.Decode(&task)
		tasks = append(tasks, task)
	}

	// Encode "tasks" array into JSON
	c.JSON(http.StatusOK, tasks)
}

func (th *TaskHandler) UpdateTaskHandler(c *gin.Context) {
	id := c.Param("id")
	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		log.Error().Err(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	objectId, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{
		"$set": bson.M{
			"name": task.Name,
		},
	}
	_, err := th.collection.UpdateOne(th.ctx, bson.M{"_id": objectId}, filter)
	if err != nil {
		log.Error().Err(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task has been updated"})
}
