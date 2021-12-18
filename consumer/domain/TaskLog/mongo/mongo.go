package mongoTaskLog

import (
	"context"
	"os"
	"time"

	domain "github.com/papillon1102/Go-Tasks-Re/domain/TaskLog"
	"github.com/phuslu/log"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
)

// MongoRepo -> mongoTaskLog -> domain TaskLog

type MongoRepo struct {
	taskLog *mongo.Collection
}

// mongoTaskLog : represent TaskLog collection in MongoDB
type mongoTaskLog struct {
	UserID     string      `json:"userid"`
	TaskID     string      `json:"taskid"`
	EventName  string      `json:"eventname"`
	RoutingKey string      `json:"routingkey"`
	Content    interface{} `json:"content"`
	Time       string      `json:"time"`
}

func NewTaskLog(dt domain.TaskLog) mongoTaskLog {
	return mongoTaskLog{
		UserID:     dt.GetUserID(),
		TaskID:     dt.GetTaskID(),
		EventName:  dt.GetEventName(),
		RoutingKey: dt.GetRoutingKey(),
		Content:    dt.GetContent(),
		Time:       dt.GetTime(),
	}
}

// New: return TaskLog-Collection of MongoDB
func New(collectionName string, mongoClient *mongo.Client) (*MongoRepo, error) {
	collection := mongoClient.Database(os.Getenv("MONGO_DATABASE")).Collection(collectionName)

	return &MongoRepo{
		taskLog: collection,
	}, nil
}

// ToAggregate : convert data got from Mongo to TaskLog Domain
func (m mongoTaskLog) ToAggregate() domain.TaskLog {
	dt := domain.TaskLog{}

	dt.SetContent(m.Content)
	dt.SetTaskId(m.TaskID)
	dt.SetUserId(m.UserID)
	dt.SetEventName(m.EventName)
	dt.SetRoutingKey(m.RoutingKey)
	dt.SetTime(m.Time)
	return dt
}

func (mr *MongoRepo) Get(userId string, taskId string) ([]domain.TaskLog, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	cur, err := mr.taskLog.Find(ctx, bson.M{"userid": userId, "taskid": taskId})

	if err != nil {
		log.Error().Err(err).Msg("Err in finding User Task Logs")
		return []domain.TaskLog{}, err
	}

	defer cur.Close(ctx)

	var tasklogs []domain.TaskLog
	for cur.Next(ctx) {
		var m mongoTaskLog
		err := cur.Decode(&m)
		if err != nil {
			log.Error().Err(err).Msg("Err in decoding Task-Log")
			return []domain.TaskLog{}, err
		}
		tasklog := m.ToAggregate()
		tasklogs = append(tasklogs, tasklog)
	}

	return tasklogs, nil
}

// MakeReport return all events created by user
func (mr *MongoRepo) MakeReport(userId string) ([]domain.TaskLog, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	var dls []domain.TaskLog
	cur, err := mr.taskLog.Find(ctx, bson.M{"userid": userId})

	if err != nil {
		log.Error().Err(err).Msg("Err in finding User Task Logs")
		return []domain.TaskLog{}, err
	}

	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var m mongoTaskLog
		err = cur.Decode(&m)
		if err != nil {
			log.Error().Err(err).Msg("Err in decoding Task-Log-Report")
			return []domain.TaskLog{}, err
		} else {
			log.Debug().Msgf("Domain Task %v\n", m)
		}

		domainTaskLog := m.ToAggregate()

		dls = append(dls, domainTaskLog)

	}

	return dls, nil
}
