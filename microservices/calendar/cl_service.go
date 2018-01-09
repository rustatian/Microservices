package calendar

type TaskService interface {
	GetTasks(username string, tr timeRange) (resp []string, e error)
	Add()
	Edit()
	Delete()
}

type trange struct {
	rng timeRange
}

type tcalendarService struct{}

func NewService() TaskService {
	return tcalendarService{}
}

//Not implemented
func (s tcalendarService) Add() {

}

//Not implemented
func (s tcalendarService) Edit() {

}

//Not implemented
func (s tcalendarService) Delete() {

}

func (s tcalendarService) GetTasks(username string, tr timeRange) (resp []string, e error) {
	switch tr {
	case TDay:

	case TWeek:

	case TMonth:

	case TYear:

	default:

	}

	return []string{"", ""}, nil
}
