package main

import (
	"net/http"
	"time"
	"github.com/dgrijalva/jwt-go"
	"github.com/auth0/go-jwt-middleware"
	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
	"os"
	"io/ioutil"
	"github.com/buger/jsonparser"
	"Models"
	"golang.org/x/crypto/bcrypt"
	"log"
	"dboperation"
)

var mySigningKey = []byte("ZXCfdsa1208")

type Message struct {
	Name string
	Body string
	Time int64
}



func main() {

	//dboperation.UserRegistration(Models.User{})

	var r *mux.Router = mux.NewRouter()

	r.Handle("/", mainHandle).Methods("GET")

	r.Handle("/login", LoginHandle).Methods("POST")
	r.Handle("/registration", RegistrationHandle).Methods("POST")
	r.Handle("/get-token", GetTokenHandle).Methods("GET")

	var loggetRoute http.Handler = handlers.LoggingHandler(os.Stdout, r)
	http.ListenAndServe(":3000", loggetRoute)
}

var mainHandle = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from Belarus and gollang!!"))
})

var LoginHandle = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Can't read Body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var user Models.User

	username, _, _, erru := jsonparser.Get(body, "newUser", "username")
	if erru != nil {
		panic(erru.Error())
		http.Error(w, "Username parse error", http.StatusInternalServerError)
		return
	}

	password, _, _, errp := jsonparser.Get(body, "newUser", "password")
	if errp != nil {
		panic(errp.Error())
		http.Error(w, "Password parse error", http.StatusInternalServerError)
		return
	}

	user.Username = string(username)

	hash, errf := dboperation.GetHashFromDb(user)
	if errf == false {
		http.Error(w, "Db operation error", http.StatusBadRequest)
		return
	}

	var isGood bool = comparePassword(hash, password)

	if isGood == true {
		token := jwt.New(jwt.SigningMethodHS256)
		claims := token.Claims.(jwt.MapClaims)

		claims["admin"] = true
		claims["name"] = string(username)
		claims["exp"] = time.Now().Add(time.Hour * 24).Unix()
		key := []byte("ZXCfdsa1208")

		JsonWebToken, _ := token.SignedString(key)
		user.PasswordHash = JsonWebToken
		w.Write([]byte(JsonWebToken))
		var isErr = dboperation.UpdateTokenForUser(user)
		if isErr == false {
			http.Error(w, "Error when update Hash for user", http.StatusInternalServerError)
			return
		}

	} else {
		http.Error(w, "Passwords do not match!", http.StatusUnauthorized)
		return
	}
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

var RegistrationHandle http.HandlerFunc = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil{
		http.Error(w, "Can't read", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	fullname, _, _, errf := jsonparser.Get(body, "newUser", "fullname")
	if errf != nil {
		panic(errf.Error())
		http.Error(w, "Fullname parse error", http.StatusInternalServerError)
	}
	username, _, _, erru := jsonparser.Get(body, "newUser", "username")
	if erru != nil {
		panic(erru.Error())
		http.Error(w, "Username parse error", http.StatusInternalServerError)
	}
	password, _, _, errp := jsonparser.Get(body, "newUser", "password")
	if errp != nil {
		panic(errp.Error())
		http.Error(w, "Password parse error", http.StatusInternalServerError)
	}
	email, _, _, erre := jsonparser.Get(body, "newUser", "email")
	if erre != nil {
		panic(erre.Error())
		http.Error(w, "Email parse error", http.StatusInternalServerError)
	}

	if len(fullname) < 5 || len(username) < 5 || len(password) < 5 || len(email) < 5 {
		http.Error(w, "Lenght less that 5", http.StatusNotAcceptable)
		return
	}

	var user Models.User
	user.FullName = string(fullname)
	user.Username = string(username)
	user.PasswordHash = HashAndSalt(password)
	user.Email = string(email)
	user.IsDisables = false

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	//TODO: admin?? How to choose
	claims["admin"] = true
	claims["name"] = user.Username
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	JsonWT, _ := token.SignedString(mySigningKey)
	user.JsonToken = JsonWT

	var isSuccsess bool = dboperation.WriteDataToDb(user)
	if isSuccsess == false{
		http.Error(w, "Error when writing to database", http.StatusInternalServerError)
		return
	}

	w.Write([]byte(JsonWT))
})

func HashAndSalt (pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}

	return string(hash)

}

func comparePassword(hashedPwd string, plainPwd []byte) bool {
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPwd)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

var NotImplemented = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Not Implemented"))
})

var jwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
	ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
		return mySigningKey, nil
	},

	SigningMethod: jwt.SigningMethodHS256,
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
