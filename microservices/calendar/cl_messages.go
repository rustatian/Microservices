package calendar

type timeRange int

const (
	TDay   timeRange = iota
	TWeek
	TMonth
	TYear
)

type TasksRequest struct {
	User      string `json:"user"`
	TimeRange int    `json:"time_range"`
	Err       string `json:"err, omitempty"`
}

type TasksResponce struct {
	TaskId          int    `json:"task_id"`
	TaskCaption     string `json:"task_caption"`
	TaskDescription string `json:"task_description"`
	From            int64  `json:"from"`
	To              int64  `json:"to"`
	Err             string `json:"err,omitempty"`
}
