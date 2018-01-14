package calendar

type timeRange int

const (
	TDay timeRange = iota
	TWeek
	TMonth
	TYear
)

type tasksRequest struct {
	User      string `json:"user"`
	TimeRange int    `json:"time_range"`
	Err       string `json:"err, omitempty"`
}

type Task struct {
	TaskId          int    `json:"task_id"`
	TaskCaption     string `json:"task_caption"`
	TaskDescription string `json:"task_description"`
	From            uint64 `json:"from"`
	To              uint64 `json:"to"`
}

type tasksResponce struct {
	Tasks string `json:"tasks, string"`
	Err   string `json:"err, omitempty"`
}

//TODO Create health logic, check free memory, disk space
type healthRequest struct{}

type healthResponse struct {
	Status bool `json:"status"`
}
