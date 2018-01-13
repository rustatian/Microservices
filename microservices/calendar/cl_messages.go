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

type task struct {
	TaskId          int
	TaskCaption     string
	TaskDescription string
	From            uint64
	To              uint64
	Err             string
}

type tasksResponce struct {
	Tasks string `json:"tasks"`
	Err   string `json:"err, omitempty"`
}

//TODO Create health logic, check free memory, disk space
type healthRequest struct{}

type healthResponse struct {
	Status bool `json:"status"`
}
