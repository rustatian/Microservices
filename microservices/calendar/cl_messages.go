package calendar

type timeRange int

const (
	TDay timeRange = iota
	TWeek
	TMonth
	TYear
)

type TasksRequest struct {
	User      string `json:"user"`
	TimeRange int    `json:"time_range"`
	Err       string `json:"err, omitempty"`
}

type Task struct {
	TaskId          int
	TaskCaption     string
	TaskDescription string
	From            uint64
	To              uint64
	Err             string
}

type TasksResponce struct {
	Tasks string `json:"tasks"`
	Err   string `json:"err, omitempty"`
}
