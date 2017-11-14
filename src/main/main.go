package main

import (
	//"encoding/json"
	"net/http"
	"os"
	"time"
	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"encoding/json"
	"Models"
	"dboperation"
)

var mySigningKey = []byte("1")




func main() {


	dboperation.UserRegistration(Models.User{})

	var r *mux.Router = mux.NewRouter()

	r.Handle("/login", jwtMiddleware.Handler(LoginHandle)).Methods("POST")
	r.Handle("/registration", jwtMiddleware.Handler(RegistrationHandle)).Methods("POST")
	r.Handle("/get-token", GetTokenHandle).Methods("GET")

	var loggetRoute http.Handler = handlers.LoggingHandler(os.Stdout, r)
	http.ListenAndServe(":3000", loggetRoute)
}

var LoginHandle = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello"))

})

var GetTokenHandle = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["admin"] = true
	claims["name"] = "Artsiom Piashchynski"
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	tokenString, _ := token.SignedString(mySigningKey)

	w.Write([]byte(tokenString))
})

var RegistrationHandle = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	data, _ := json.Marshal(Models.User{})
	w.Write([]byte(data))
	//vars := mux.Vars(r)
	//w.Write([]byte(vars))

})

var NotImplemented = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Not Implemented"))
})


var jwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
	ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
		return mySigningKey, nil
	},

	SigningMethod: jwt.SigningMethodHS256,
})

var StatusHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("API is up and running"))
})

var ProductsHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//payload, _ := json.Marshal(products)
	//w.Header().Set("Content-Type", "application/json")
	//w.Write([]byte(payload))
})

var AddFeedbackHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//var product Product
	//var vars map[string]string = mux.Vars(r)
	//var slug string = vars["slug"]
	//
	//for _, p := range products {
	//	if p.Slug == slug {
	//		product = p
	//	}
	//}
	//
	//w.Header().Set("Content-Type", "application/json")
	//
	//if product.Slug != "" {
	//	payload, _ := json.Marshal(product)
	//	w.Write([]byte(payload))
	//} else {
	//	w.Write([]byte("Product Not Found"))
	//}
})
