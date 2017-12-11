package registration

import (
	"context"
	"database/sql"
	"github.com/go-kit/kit/endpoint"
	"github.com/spf13/viper"
)

var dbCreds string

func init() {
	viper.SetConfigName("config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	dbCreds = viper.GetString("DbCreds.server")
}

type Service interface {
	Registration(username, fullname, email, passwordHash, jwtToken string, isDisabled bool) (bool, error)
	UsernameValidation()
	EmailValidation()
}

func NewRegService() Service {
	return newRegService{}
}

type newRegService struct {}


func(newRegService) Registration(username, fullname, email, passwordHash, jwtToken string, isDisabled bool) (bool, error) {
	db, err := sql.Open("mysql", dbCreds)
	if err != nil {
		return false, err
	}

	defer db.Close()

	stmIns, err := db.Prepare("INSERT INTO USER (Username, FullName, email, PasswordHash, jwtToken, IsDisabled) VALUES (?, ?, ?, ?, ?, ?);")
	defer stmIns.Close()

	_, err = stmIns.Exec(username, fullname, email, passwordHash, jwtToken, isDisabled)
	if err != nil {
		return false, err
	}

	return true, nil
}

func(newRegService) UsernameValidation() {

}

func(newRegService) EmailValidation() {

}


type RegRequest struct {
	Username string `json:"username"`
	Fullname string `json:"fullname"`
	Email string `json:"email"`
	PasswordHash string `json:"password_hash"`
	JwtToken string `json:"jwt_token"`
	isDisabled bool `json:"is_disabled"`
}

type RegResponce struct {
	Status bool `json:"status"`
	Err string `json:"err, omitempty"`
}

type UsernameValidationRequest struct {
	User string `json:"user"`
}

type UsernameValidationResponce struct {
	Status string `json:"status"`
	Err string `json:"err, omitempty"`
}

type EmailValidationRequest struct {
	Email string `json:"email"`
}

type EmailValidationResponce struct {
	Status string `json:"status"`
	Err string `json:"err, omitempty"`
}

type Endpoints struct {
	RegEndpoint endpoint.Endpoint
	UsernameValidEndpoint endpoint.Endpoint
	EmailValidEndpoint endpoint.Endpoint
}


func(e Endpoints) MakeRegEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req, err := request.(RegRequest)
		if err != nil {
			return nil, err
		}

		ok, err := svc.Registration(req.Username, req.Fullname, req.Email, req.PasswordHash, req.JwtToken, req.isDisabled); if !ok {
			return nil, err
		}

		return RegResponce{Err:"", Status: ok}, nil
	}
}

func NewEnpoints() endpoint.Middleware {
	return
}





