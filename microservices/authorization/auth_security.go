package authorization

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/leonelquinteros/gorand"
	"github.com/spf13/viper"
	"gopkg.in/redis.v3"
	"os"
	"time"
)

var secret string

func init() {
	if dev := os.Getenv("DEV"); dev == "False" {
		viper.AddConfigPath("config")
		viper.SetConfigName("app_conf")

		err := viper.ReadInConfig()
		if err != nil {
			panic(err)
		}

		secret = viper.GetString("SecretKey.Key")
	} else {
		viper.AddConfigPath("config")
		viper.SetConfigName("app_conf")

		err := viper.ReadInConfig()
		if err != nil {
			panic(err)
		}

		secret = viper.GetString("SecretKey.Key")
	}
}

// Signing method
var (
	method = jwt.SigningMethodHS256
)

func JwtLoginEndpoint(log log.Logger) endpoint.Middleware {
	return func(i endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			req := request.(LoginRequest)
			response, err = i(ctx, req)

			if err != nil {
				return nil, err
			}

			resp := response.(LoginResponce)
			if resp.Err != "" {
				return nil, err
			}

			err = loginHandler(req.Username, &resp, log)
			return resp, err
		}
	}
}

func JwtLogoutEndpoint(log log.Logger) endpoint.Middleware {
	return func(i endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			req := request.(LogoutRequest)
			response, err = i(ctx, req)

			if err != nil {
				return nil, err
			}

			resp := response.(LogoutResponce)
			err = logoutHandler(req, &resp, log)
			return "", err
		}
	}
}

//Just for experiment (Redis)
func loginHandler(username string, resp *LoginResponce, log log.Logger) error {
	var (
		tokenString string
	)

	uuid, err := gorand.UUIDv4()
	if err != nil {
		panic(err.Error())
	}

	defer func() {
		log.Log(
			"username", username,
			"jti", uuid,
			"token", tokenString,
		)
	}()

	token := jwt.New(method)
	claims := token.Claims.(jwt.MapClaims)

	m := map[string]interface{}{
		"username": username,
		"roles":    resp.Roles,
	}
	val, _ := json.Marshal(m)

	claims["admin"] = true
	claims["iat"] = time.Now()
	claims["iss"] = "Valery_P"
	claims["name"] = username
	claims["exp"] = time.Now().Add(time.Hour * 48).Unix()
	claims["jti"] = uuid
	JsonWebToken, err := token.SignedString([]byte(secret))
	tokenString = JsonWebToken[:20] + "..."
	if err != nil {
		return err
	}

	resp.TokenString = JsonWebToken

	errChan := make(chan error)

	go func() {
		client := redis.NewClient(
			&redis.Options{
				Addr:     "127.0.0.1:6379",
				Password: "",
				DB:       0,
			})

		uuid, _ := gorand.MarshalUUID(uuid)
		var err *redis.StatusCmd = client.Set(uuid, val, time.Duration(time.Hour*24))
		if err != nil {
			errChan <- err.Err()
		}
		errChan <- nil
	}()

	if err = <-errChan; err != nil {
		return err
	}

	return nil
}

//TODO create logout with database
// handling logout
func logoutHandler(req LogoutRequest, resp *LogoutResponce, log log.Logger) error {

	var (
		username    string
		jti         string
		tokenString string
	)

	defer func() {
		log.Log(
			"username", username,
			"jti", jti,
			"token", tokenString,
		)

	}()

	tokenString = req.TokenString
	username = req.Username

	kf := func(token *jwt.Token) (interface{}, error) {
		ok := token.Valid
		if !ok {
			return nil, errors.New("token is not valid")
		}
		return []byte(secret), nil
	}

	w, err := jwt.Parse(tokenString, kf)
	println(w)
	if err != nil {
		return err
	}

	errChan := make(chan error)
	//remove UUID on Consul KV
	go func() {
		client := redis.NewClient(
			&redis.Options{
				Addr:     "localhost:6379",
				Password: "",
				DB:       0,
			})
		err := client.Del(jti)

		if err != nil {
			errChan <- err.Err()
		} else {
			errChan <- nil
		}
	}()

	if err = <-errChan; err != nil {
		return err
	}

	return nil
}
