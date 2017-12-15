package auth

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/leonelquinteros/gorand"
	"time"
	"gopkg.in/redis.v3"
	"encoding/json"
	"github.com/spf13/viper"
)

var secret string

func init() {
	viper.AddConfigPath("/config")
	viper.SetConfigFile("reg_srv_conf")

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	secret = viper.GetString("SecretKey.Key")
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
			err = loginHandler(req.Username, &resp, log)
			return resp, err
		}
	}
}

func JwtLogoutEndpoint(log log.Logger) endpoint.Middleware {
	return func(i endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			//req := request.(Logou)
			//response, err = i(ctx, req)
			//
			//if err != nil {
			//	return nil, err
			//}
			//
			//resp := response.(LoginResponce)
			//err = logoutHandler(req.Username, &resp, log)
			return "", err
		}
	}
}

//TODO insert token into db
func loginHandler(username string, resp *LoginResponce, log log.Logger) error {
	var (
		cid string
		tokenString string
	)

	defer func(){
		log.Log(
			"username", username,
			"jwtid", cid,
			"token", tokenString,
		)
	}()

	uuid, err := gorand.UUID()
	if err != nil {
		panic(err.Error())
	}


	token := jwt.New(method)
	claims := token.Claims.(jwt.MapClaims)

	m := map[string]interface{} {
		"username": username,
		"roles": resp.Roles,
	}
	val, _ := json.Marshal(m)

	claims["admin"] = true
	claims["iat"] = time.Now()
	claims["iss"] = "Valery_P"
	claims["name"] = username
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	claims["jti"] = cid
	JsonWebToken, err := token.SignedString(secret)
	tokenString = JsonWebToken[:20] + "..."
	if err != nil {

	}

	resp.TokenString = JsonWebToken

	errChan := make(chan error)
	go func() {
		client := redis.NewClient(
			&redis.Options{
				Addr: "localhost:6379",
				Password: "",
				DB: 0,
			})

		var err2 *redis.StatusCmd = client.Set(uuid, val, time.Duration(time.Hour * 24))
		if err2.Err() != nil {
			errChan <- err
		} else {
			errChan <- nil
		}
	}()

	if err = <- errChan; err != nil {
		return err
	} else {
		return nil
	}
}

//TODO create logout with database
// handling logout
func logoutHandler(/*req AuthRequest, resp *AuthResponse,*/ log log.Logger) error {

	//var (
	//	username string
	//	cid string
	//	tokenString string
	//)
	//
	//defer func(){
	//	log.Log(
	//		"username", username,
	//		"jwtid", cid,
	//		"token", tokenString,
	//	)
	//
	//}()
	//
	//leeway := 10 * time.Minute
	//tokenString = req.TokenString
	//username = req.Username
	//w, err := jws.ParseJWT([]byte(tokenString))
	//if err != nil {
	//	return err
	//}
	//
	//claims := w.Claims()
	//
	//if jwtid, ok := claims.JWTID(); ok {
	//	cid = jwtid
	//}
	//
	//err = claims.Validate(time.Now(), leeway, leeway);
	//
	//if err == nil || err == jwt.ErrTokenIsExpired {
	//
	//	errChan := make(chan error)
	//	//remove UUID on Consul KV
	//	go func(){
	//		client := ConsulClient(consulAddress, consulPort, log)
	//		kv := client.KV()
	//		key := "session/" + cid
	//		_, e := kv.Delete (key, nil)
	//		resp.TokenString = ""
	//		if err != nil {
	//			errChan <- err
	//		} else if e != nil {
	//			errChan <- e
	//		} else {
	//			errChan <- nil
	//		}
	//	}()
	//
	//	if err = <- errChan; err != nil {
	//		return err
	//	} else if err == jwt.ErrTokenIsExpired{
	//		return err
	//	}
	//}

	return nil
}
