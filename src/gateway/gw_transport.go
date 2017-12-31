package gateway

import (
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"TaskManager/src/svcdiscovery"
	"github.com/rs/cors"
)

var (
	consulAddress string
	vaultSvcName string
	regSvcName string
	authSvcName string
	tag string
)


func init() {
	viper.AddConfigPath("src/config")
	viper.SetConfigName("app_conf")

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	consulAddress = viper.GetString("consul.address")

	//Service names
	vaultSvcName = viper.GetString("services.vault")
	regSvcName = viper.GetString("services.registration")
	authSvcName = viper.GetString("services.auth")

	tag = viper.GetString("tags.tag")
}

func MakeHttpHandler() http.Handler {
	r := mux.NewRouter()

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "HEAD", "POST", "PUT", "OPTIONS"},
	})

	//vault path
	r.Methods("POST").HandlerFunc(hash).Path("/hash")
	r.Methods("POST").HandlerFunc(validate).Path("/validate")

	//reg path
	r.Methods("POST").HandlerFunc(registration).Path("/registration")
	r.Methods("POST").HandlerFunc(regvalemail).Path("/registration/email")
	r.Methods("POST").HandlerFunc(regvaluser).Path("/registration/user")

	//auth
	r.Methods("POST").HandlerFunc(login).Path("/login")

	handler := c.Handler(r)

	return handler
}

//authorization
func login(writer http.ResponseWriter, request *http.Request) {
	addr, err := svcdiscovery.ServiceDiscovery().Find(&consulAddress, &authSvcName, &tag)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
	}

	resp, err := http.Post(addr + "/login", "application/json", request.Body)

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
	}

	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.Write(data)
}

//registration
func regvaluser(writer http.ResponseWriter, request *http.Request) {
	addr, err := svcdiscovery.ServiceDiscovery().Find(&consulAddress, &regSvcName, &tag)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
	}

	resp, err := http.Post(addr + "/registration/user", "application/json", request.Body)

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
	}

	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.Write(data)
}

//registration
func regvalemail(writer http.ResponseWriter, request *http.Request) {
	addr, err := svcdiscovery.ServiceDiscovery().Find(&consulAddress, &regSvcName, &tag)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
	}

	resp, err := http.Post(addr + "/registration/email", "application/json", request.Body)

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
	}

	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.Write(data)
}

// /registration
func registration(writer http.ResponseWriter, request *http.Request) {

	addr, err := svcdiscovery.ServiceDiscovery().Find(&consulAddress, &regSvcName, &tag)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	resp, err := http.Post(addr + "/registration", "application/json", request.Body)
	defer resp.Body.Close()
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.Write(data)
}

// /validate
func validate(writer http.ResponseWriter, request *http.Request) {
	addr, err := svcdiscovery.ServiceDiscovery().Find(&consulAddress, &vaultSvcName, &tag)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	resp, err := http.Post(addr + "/validate", "application/json", request.Body)
	defer resp.Body.Close()
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.Write(data)
}

// /hash
func hash(writer http.ResponseWriter, r *http.Request) {

	//Get service address
	addr, err := svcdiscovery.ServiceDiscovery().Find(&consulAddress, &vaultSvcName, &tag)

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := http.Post(addr+"/hash", "application/json", r.Body)
	defer resp.Body.Close()

	if err != nil {
		writer.Write([]byte(err.Error()))
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.Write(data)

}
