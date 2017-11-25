package main

import (
	"dboperation"
	"encoding/json"
	"github.com/auth0/go-jwt-middleware"
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

	var i interface{}
	json.Unmarshal(body, &i)

	userVars := i.(map[string]interface{})

	if str, ok := userVars["user"].(string); ok {
		var isUserExist = dboperation.CheckifUserExist(str)
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

	var i interface{}
	json.Unmarshal(body, &i)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Decoding error", http.StatusInternalServerError)
		return
	}

	unparcedJsn := i.(map[string]interface{})
	if usrStr, ok := unparcedJsn["username"].(string); ok {
		var isUserExist = dboperation.CheckifUserExist(usrStr)

		if isUserExist == false {
			http.Error(w, "User does't exist", http.StatusNotFound)
			return
		} else {
			user.Username = usrStr
		}
	}

	if pswdStr, ok := unparcedJsn["password"].(string); ok {

		hash, ok := dboperation.GetHashFromDb(user)
		if ok == false {
			http.Error(w, "Db operation error", http.StatusInternalServerError)
			return
		}

		var pswd = []byte(pswdStr)

		var isGood = comparePassword(hash, pswd)

		user.PasswordHash = hash

		if isGood == true {
			token := jwt.New(jwt.SigningMethodHS256)
			claims := token.Claims.(jwt.MapClaims)

			claims["admin"] = true
			claims["name"] = string(user.Username)
			claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

			secret := []byte(configuration("config.json").Secret)
			JsonWebToken, _ := token.SignedString(secret)
			user.JsonToken = JsonWebToken

			var isErr = dboperation.UpdateTokenForUser(user)
			if isErr == false {
				http.Error(w, "Error when update Hash for user", http.StatusInternalServerError)
				return
			}

			js, _ := json.Marshal(JsonWebToken)
			w.WriteHeader(http.StatusOK)
			w.Write(js)

		} else {
			http.Error(w, "Incorrect username or password", http.StatusUnauthorized)
			return
		}
	}
})

var GetTokenHandle = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["admin"] = true
	claims["name"] = "Artsiom Piashchynski"
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	secret := []byte(configuration("config.json").Secret)
	tokenString, _ := token.SignedString(secret)

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
	if fullname, ok := newUser["fullname"].(string); ok {
		user.FullName = string(fullname)
	} else {
		panic(err.Error())
		http.Error(w, "Fullname parse error", http.StatusInternalServerError)
	}

	username := newUser["username"].(string)
	password := []byte(newUser["password"].(string))
	email := newUser["email"].(string)

	defer r.Body.Close()

	user.Username = string(username)
	user.PasswordHash = HashAndSalt(password)
	user.Email = string(email)
	user.IsDisables = false

	var isUserExist = dboperation.CheckifUserExist(string(username))
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

	secret := []byte(configuration("config.json").Secret)

	tokenString, err := token.SignedString(secret)
	if err != nil {
		log.Print(err.Error())
	}
	user.JsonToken = tokenString

	var isSuccsess = dboperation.WriteDataToDb(user)
	if isSuccsess == false {
		http.Error(w, "Error when writing to database", http.StatusInternalServerError)
		return
	}

	js, _ := json.Marshal(tokenString)
	w.Write(js)

})

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

var jwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
	ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
		return []byte(configuration("config.json").Secret), nil
	},

	SigningMethod: jwt.SigningMethodHS256,
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
