package domain

// TaskLogRepository : is the interface that define the commands that TasKLog domain able to make
// We will use this interface as the source for our business logic!
type TaskLogRepository interface {
	Get(userId, taskId string) ([]TaskLog, error)
	MakeReport(userId string) ([]TaskLog, error)
}
