package main

import (
	"bytes"
	"dboperation"
	"encoding/gob"
	"encoding/json"
	"errors"
	"github.com/auth0/go-jwt-middleware"
	"github.com/buger/jsonparser"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"golang.org/x/crypto/bcrypt"
	"io/ioutil"
	"log"
	"models"
	"net/http"
	"os"
	"time"
)

func main() {
	var r = mux.NewRouter()

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "HEAD", "POST", "PUT", "OPTIONS"},
	})

	r.HandleFunc("/", mainHandle).Methods("GET")
	r.HandleFunc("/login", LoginHandle).Methods("POST")
	r.HandleFunc("/registration", RegistrationHandle).Methods("POST")
	r.HandleFunc("/get-token", GetTokenHandle).Methods("GET")
	r.HandleFunc("/registration/email", validationMailHandle).Methods("POST")
	r.HandleFunc("/registration/user", validationUserHandle).Methods("POST")

	handler := c.Handler(r)

	//var loggetRoute http.Handler = handlers.LoggingHandler(os.Stdout, r)
	log.Fatal(http.ListenAndServe(":3000", handler))
}

var mainHandle = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from Belarus and golang!!"))
})

var validationMailHandle = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Can't read Body", http.StatusInternalServerError)
		return
	}

	defer r.Body.Close()

	var mail interface{}

	var user models.User

	err = json.Unmarshal(body, &mail)

	if err != nil {
		panic(err.Error())
		http.Error(w, "Email parse error", http.StatusInternalServerError)
		return
	}

	UnparsedEmail := mail.(map[string]interface{})
	if str, ok := UnparsedEmail["email"].(string); ok {
		user.Email = str

		var isEmailExist = dboperation.CheckifMailExist(user)
		if isEmailExist == false {
			w.WriteHeader(http.StatusOK)
			return
		} else {
			http.Error(w, "Email already registered!", http.StatusConflict)
		}
	} else {
		http.Error(w, "Email parse error", http.StatusInternalServerError)
	}

})

var validationUserHandle = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Can't read Body", http.StatusInternalServerError)
		return
	}

	defer r.Body.Close()

	var user models.User

	var i interface{}
	json.Unmarshal(body, &i)

	userVars := i.(map[string]interface{})

	if str, ok := userVars["user"].(string); ok {
		user.Username = str
		var isUserExist = dboperation.CheckifUserExist(user)
		if isUserExist == false {
			w.WriteHeader(http.StatusOK)
			return
		} else {
			http.Error(w, "User exist!", http.StatusConflict)
		}

	} else {
		panic(err.Error())
		http.Error(w, "Username parse error", http.StatusInternalServerError)
		return
	}

})

var LoginHandle = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Can't read Body", http.StatusInternalServerError)
		return
	}

	defer r.Body.Close()

	var user models.User

	username, _, _, erru := jsonparser.Get(body, "user", "username")
	if erru != nil {
		panic(erru.Error())
		http.Error(w, "Username parse error", http.StatusInternalServerError)
		return
	}

	user.Username = string(username)

	var isUserExist bool = dboperation.CheckifUserExist(user)
	if isUserExist == false {
		http.Error(w, "User does't exist", http.StatusNotFound)
		return
	}

	password, _, _, err := jsonparser.Get(body, "user", "password")
	if err != nil {
		panic(err.Error())
		http.Error(w, "Password parse error", http.StatusInternalServerError)
		return
	}

	hash, uerr := dboperation.GetHashFromDb(user)
	if uerr == false {
		http.Error(w, "Db operation error", http.StatusInternalServerError)
		return
	}

	var isGood bool = comparePassword(hash, password)

	if isGood == true {
		token := jwt.New(jwt.SigningMethodHS256)
		claims := token.Claims.(jwt.MapClaims)

		claims["admin"] = true
		claims["name"] = string(username)
		claims["exp"] = time.Now().Add(time.Hour * 24).Unix()
		key := []byte(configuration("config.json").DbCreds)

		JsonWebToken, _ := token.SignedString(key)
		user.JsonToken = JsonWebToken

		js, err := json.Marshal(JsonWebToken)
		if err != nil {
			w.Write(js)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		var isErr = dboperation.UpdateTokenForUser(user)
		if isErr == false {
			http.Error(w, "Error when update Hash for user", http.StatusInternalServerError)
			return
		}

	} else {
		http.Error(w, "Incorrect username or password", http.StatusUnauthorized)
		return
	}
})

var GetTokenHandle = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["admin"] = true
	claims["name"] = "Artsiom Piashchynski"
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	tokenString, _ := token.SignedString(configuration("config.json").DbCreds)

	w.Write([]byte(tokenString))
})

var RegistrationHandle = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Can't read", http.StatusBadRequest)
		return
	}

	var user models.User
	var i interface{}
	json.Unmarshal(body, &i)

	newUser := i.(map[string]interface{})

	jsonData := newUser["newUser"].(map[string]interface{})
	if fullname, ok := jsonData["fullname"].(string); ok {
		user.FullName = string(fullname)
	} else {
		panic(err.Error())
		http.Error(w, "Fullname parse error", http.StatusInternalServerError)
	}

	username := jsonData["username"].(string)

	// Convert to bytes
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(jsonData["password"])
	if err != nil {
		log.Print(err.Error())
		return
	}

	password := buf.Bytes()
	email := jsonData["email"].(string)

	defer r.Body.Close()

	user.Username = string(username)
	user.PasswordHash = HashAndSalt(password)
	user.Email = string(email)
	user.IsDisables = false

	var isUserExist bool = dboperation.CheckifUserExist(user)
	if isUserExist == true {
		http.Error(w, "User exist", http.StatusConflict)
		return
	}

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	//TODO: admin?? How to choose
	claims["admin"] = true
	claims["name"] = user.Username
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	JsonWebToken, _ := token.SignedString(configuration("config.json").DbCreds)
	user.JsonToken = JsonWebToken

	var isSuccsess bool = dboperation.WriteDataToDb(user)
	if isSuccsess == false {
		http.Error(w, "Error when writing to database", http.StatusInternalServerError)
		return
	}

	js, err := json.Marshal(JsonWebToken)
	if err != nil {
		w.Write(js)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
})

func registrationVarsParsing(data []byte, w http.ResponseWriter) ([]byte, []byte, []byte, []byte, error) {
	fullname, _, _, err := jsonparser.Get(data, "newUser", "fullname")
	if err != nil {
		panic(err.Error())
		http.Error(w, "Fullname parse error", http.StatusInternalServerError)
		return nil, nil, nil, nil, err
	}
	username, _, _, err := jsonparser.Get(data, "newUser", "username")
	if err != nil {
		panic(err.Error())
		http.Error(w, "Username parse error", http.StatusInternalServerError)
		return nil, nil, nil, nil, err
	}
	password, _, _, err := jsonparser.Get(data, "newUser", "password")
	if err != nil {
		panic(err.Error())
		http.Error(w, "Password parse error", http.StatusInternalServerError)
		return nil, nil, nil, nil, err
	}
	email, _, _, err := jsonparser.Get(data, "newUser", "email")
	if err != nil {
		panic(err.Error())
		http.Error(w, "Email parse error", http.StatusInternalServerError)
		return nil, nil, nil, nil, err
	}

	if len(fullname) < 5 || len(username) < 5 || len(password) < 7 || len(email) < 5 {
		http.Error(w, "Lenght less that 5", http.StatusNotAcceptable)
		return nil, nil, nil, nil, errors.New("lenght less that 5")
	}
	return fullname, username, password, email, nil
}

func HashAndSalt(pwd []byte) string {
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
		return configuration("config.json").DbCreds, nil
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

func configuration(configName string) models.Configuration {
	file, _ := os.Open(configName)
	decode := json.NewDecoder(file)
	configuration := models.Configuration{}
	err := decode.Decode(&configuration)
	if err != nil {
		panic(err.Error())
	}
	return configuration
}
