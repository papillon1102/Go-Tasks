package list

import (
	"context"
	"errors"
	"strings"
	"time"

	domain "github.com/papillon1102/Go-Tasks-Re/domain/TaskLog"
	mongoTaskLog "github.com/papillon1102/Go-Tasks-Re/domain/TaskLog/mongo"
	"github.com/phuslu/log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type ListService struct {
	taskLogs domain.TaskLogRepository
}

func NewListService(lcfs ...ListConfig) (*ListService, error) {

	// Create the listservice
	ls := &ListService{}
	// Apply all Configurations passed in
	for _, lcf := range lcfs {
		// Pass the service into the configuration function
		err := lcf(ls)
		if err != nil {
			return nil, err
		}
	}
	return ls, nil

}

type ListConfig func(ls *ListService) error

func WithMongoTaskLogRepo(uriString string, collectionName string) ListConfig {
	return func(ls *ListService) error {
		ctx := context.Background()
		mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(uriString))
		if err = mongoClient.Ping(context.TODO(), readpref.Primary()); err != nil {
			log.Fatal().Err(err)
		} else {
			log.Info().Msg("Connected to MongoDB")
		}
		taskLogMongo, err := mongoTaskLog.New(collectionName, mongoClient)
		if err != nil {
			return err
		}

		// taskLogs interface now has been replaced with Mongo Repository of Task Log
		ls.taskLogs = taskLogMongo
		return nil
	}

}

type TaskDetailReport struct {
	Name string `json:"name"`
	Key  string `json:"key"`
	Data string `json:"data"`
	Time string `json:"time"`
}

// LisTaskLog returns all the info related to a specific task
func (ls *ListService) ListTaskLog(userid, taskid string) ([]TaskDetailReport, error) {
	tls, err := ls.taskLogs.Get(userid, taskid)
	if err != nil {
		return []TaskDetailReport{}, err
	}
	var tdrs []TaskDetailReport
	for _, tl := range tls {
		taskDetail := TaskDetailReport{
			Name: tl.GetEventName(),
			Key:  tl.GetRoutingKey(),
			Data: tl.GetContent(),
			Time: tl.GetTime(),
		}
		tdrs = append(tdrs, taskDetail)
	}
	return tdrs, nil
}

// ReturnTaskLogReport return report for all user's activities with their tasks today.
func (ls *ListService) ReturnTaskLogReport(userid string) ([]TaskDetailReport, error) {
	taskLogs, err := ls.taskLogs.MakeReport(userid)
	if err != nil {
		return []TaskDetailReport{}, err
	}

	var tasksReport []TaskDetailReport
	for _, tl := range taskLogs {

		currentTime := time.Now()
		currentDate := currentTime.Format("2006-01-02")

		// Split time from " "
		taskTime := strings.Split(tl.GetTime(), " ")

		log.Debug().Msgf("Task_Date : %s, currentDate: %s\n", taskTime[0], currentDate)

		if currentDate == taskTime[0] {
			taskRP := TaskDetailReport{
				Name: tl.GetEventName(),
				Key:  tl.GetRoutingKey(),
				Data: tl.GetContent(),
				Time: tl.GetTime(),
			}
			tasksReport = append(tasksReport, taskRP)
		} else {
			var findErr = errors.New("Err in mathicng")
			return nil, findErr
		}

	}

	return tasksReport, nil
}
