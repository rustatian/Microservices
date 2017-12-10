package auth

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/leonelquinteros/gorand"
	"strings"
	"time"
)

var (
	key = []byte("%kxkstXG%@uEG4^fj_gt8*XK?tzG@ddY#+wAd")
	method = jwt.SigningMethodHS256
)

func JwtEndpoint(consulAddress string, consulPort string, log log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			req := request.(AuthRequest)
			response, err = next(ctx, request)

			if err != nil {
				return nil, err
			}

			resp := response.(AuthResponse)
			if strings.EqualFold("login", req.Type) {
				err = loginHandler(consulAddress, consulPort,
					req.Username, &resp, log)
			} else if strings.EqualFold("logout", req.Type) {
				println("logout")
				err = logoutHandler(consulAddress, consulPort,
					req, &resp, log)
			}

			return resp, err
		}
	}
}
//create jwt keyFunc to retrieve kid
//func keyFunc(token *jwt.Token) (interface{}, error) {
//	return key, nil
//}
// handling login
func loginHandler(consulAddress string, consulPort string, username string, resp *AuthResponse, log log.Logger) error {
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
	cid = uuid

	token := jwt.New(method)
	claims := token.Claims.(jwt.MapClaims)

	//m := map[string]interface{} {
	//	"username": username,
	//	"roles": resp.Roles,
	//}
	//val, _ := json.Marshal(m)

	claims["admin"] = true
	claims["name"] = username
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	JsonWebToken, err := token.SignedString(key)
	tokenString = JsonWebToken[:20] + "..."
	if err != nil {

	}

	//claims.SetIssuer("ru-rocker.com")
	//claims.SetIssuedAt(time.Now())
	//claims.SetExpiration(time.Now().Add(time.Hour*24*30).Unix())
	//claims.SetJWTID(cid)
	//j := jws.NewJWT(claims, method)
	//b, err := token.Serialize(key)
	//if err != nil {
	//	return err
	//}
	//tokenString = string(b[:])

	resp.TokenString = JsonWebToken

	//errChan := make(chan error)
	////register UUID on Consul KV
	//go func() {
	//	client := ConsulClient(consulAddress, consulPort, log)
	//	kv := client.KV()
	//	key := "session/" + uuid
	//	p := &api.KVPair{Key: key, Value: []byte(val)}
	//	_, e := kv.Put(p, nil)
	//	if e != nil {
	//		errChan <- e
	//	} else {
	//		errChan <- nil
	//	}
	//}()
	//
	//if err = <- errChan; err != nil {
	//	return err
	//}
	return nil
}

//TODO create logout with database
// handling logout
func logoutHandler(consulAddress string, consulPort string, req AuthRequest, resp *AuthResponse, log log.Logger) error {

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