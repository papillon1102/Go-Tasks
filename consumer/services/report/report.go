package report

import "github.com/papillon1102/Go-Tasks-Re/services/list"

type ReportService struct {
	ListService *list.ListService
}

type ReportConfig func(rs *ReportService) error

func NewReportService(rcfs ...ReportConfig) (*ReportService, error) {

	// Create the listservice
	rs := &ReportService{}
	// Apply all Configurations passed in
	for _, rcf := range rcfs {
		// Pass the service into the configuration function
		err := rcf(rs)
		if err != nil {
			return nil, err
		}
	}
	return rs, nil

}

func WithListService(ls *list.ListService) ReportConfig {

	return func(r *ReportService) error {
		r.ListService = ls
		return nil
	}
}

// func (rs *ReportService) ReturnReport() {
// 	TaskLogs, err := rs.
// }
