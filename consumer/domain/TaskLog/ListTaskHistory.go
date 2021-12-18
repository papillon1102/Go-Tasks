package domain

import "errors"

type TaskLog struct {
	userID     string      `json:"userid"`
	taskID     string      `json:"taskid"`
	eventName  string      `json:"eventname"`
	routingKey string      `json:"routing"`
	content    interface{} `json:"content"`
	time       string      `json:"time"`
}

var (
	ErrInvalidUser      = errors.New("need to have valid user")
	ErrInvalidTask      = errors.New("need to have specific task")
	ErrInvalidEventname = errors.New("need to have event name")
)

func NewTaskLog(userid, taskid, eventname, routingkey, time string, content interface{}) (TaskLog, error) {
	if userid == "" {
		return TaskLog{}, ErrInvalidUser
	}

	if taskid == "" {
		return TaskLog{}, ErrInvalidTask
	}

	if eventname == "" {
		return TaskLog{}, ErrInvalidEventname
	}

	return TaskLog{
		userID:     userid,
		taskID:     taskid,
		eventName:  eventname,
		routingKey: routingkey,
		content:    content,
		time:       time,
	}, nil
}

// Getter TaskLog
func (t *TaskLog) GetUserID() string {
	return t.userID
}

func (t *TaskLog) GetTaskID() string {
	return t.taskID
}

func (t *TaskLog) GetEventName() string {
	return t.eventName
}

func (t *TaskLog) GetRoutingKey() string {
	return t.routingKey
}

func (t *TaskLog) GetContent() string {

	taskContent, ok := t.content.(string)
	if !ok {
		return ""
	}

	return taskContent
}

func (t *TaskLog) GetTime() string {
	return t.time
}

// Setter TaskLog

func (t *TaskLog) SetUserId(userid string) {
	t.userID = userid
}

func (t *TaskLog) SetTaskId(taskid string) {
	t.taskID = taskid
}

func (t *TaskLog) SetEventName(eventname string) {
	t.eventName = eventname
}

func (t *TaskLog) SetRoutingKey(routingkey string) {
	t.routingKey = routingkey
}

func (t *TaskLog) SetContent(content interface{}) {
	t.content = content
}

func (t *TaskLog) SetTime(time string) {
	t.time = time
}
