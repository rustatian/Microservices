package calendar

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/lib/pq"
	"github.com/spf13/viper"
	"os"
	"time"
	"runtime"
)

var (
	dbCreds string
)

func init() {
	if dev := os.Getenv("DEV"); dev == "False" {
		viper.AddConfigPath("config")
		viper.SetConfigName("app_conf")

		err := viper.ReadInConfig()
		if err != nil {
			panic(err)
		}

		dbCreds = viper.GetString("DbCreds.serverProd")

	} else {
		viper.AddConfigPath("config")
		viper.SetConfigName("app_conf")

		err := viper.ReadInConfig()
		if err != nil {
			panic(err)
		}

		dbCreds = viper.GetString("DbCreds.serverDev")

	}
}

type TaskService interface {
	GetTasks(username string, tr timeRange) (resp []string, e error)
	Add()
	Edit()
	Delete()
}

func NewTaskService() TaskService {
	return tCalendarService{}
}

type ServiceMiddleware func(svc TaskService) TaskService

type tCalendarService struct{}

func NewService() TaskService {
	return tCalendarService{}
}

//Not implemented
func (s tCalendarService) Add() {

}

//Not implemented
func (s tCalendarService) Edit() {

}

//Not implemented
func (s tCalendarService) Delete() {

}

func (s tCalendarService) GetTasks(username string, tr timeRange) (resp []string, e error) {
	switch tr {
	case TDay:

	case TWeek:

	case TMonth:
		connStr := "user=postgres dbname=postgres sslmode=disable"
		db, err := sql.Open("postgres", dbCreds)
		if err != nil {
			err = err.(*pq.Error)
			return nil, err
		}

		currTime := time.Now()
		year, month, day := currTime.Date()

		prevMnt := time.Date(year, month - 1, day, 0, 0, 0, 0, time.UTC).Unix()
		futureMnt := time.Date(year, month + 1, day, 0, 0, 0, 0, time.UTC).Unix()

		sel, err := db.Prepare("SELECT ")
		if err != nil {
			err = err.(*pq.Error)
			return nil, err
		}
		var h []TasksRequest
		rws, err := sel.Query()
		rws.Scan(&h)




	case TYear:

	default:

	}

	return []string{"", ""}, nil
}
