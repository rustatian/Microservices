package gateway

import (
	"Microservices/svcdiscovery"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/spf13/viper"
)

var (
	consulAddress string
	vaultSvcName  string
	regSvcName    string
	authSvcName   string
	tcalSvcName   string
	tag           string
	Client        servicediscovery.Client
)

func init() {
	if dev := os.Getenv("DEV"); dev == "False" {
		viper.AddConfigPath("config")
		viper.SetConfigName("app_conf")

		err := viper.ReadInConfig()
		if err != nil {
			panic(err)
		}

		consulAddress = viper.GetString("consul.addressProd")

		//Service names
		vaultSvcName = viper.GetString("services.vault")
		regSvcName = viper.GetString("services.registration")
		authSvcName = viper.GetString("services.auth")
		tcalSvcName = viper.GetString("services.tcal")

		tag = viper.GetString("tags.tag")

		// TODO temporary, rewrite client init
		Client, err = servicediscovery.NewDefaultConsulHTTPClient(consulAddress)
		if err != nil {
			panic(err)
		}
	} else {
		viper.AddConfigPath("config")
		viper.SetConfigName("app_conf")

		err := viper.ReadInConfig()
		if err != nil {
			panic(err)
		}

		consulAddress = viper.GetString("consul.addressDev")

		//Service names
		vaultSvcName = viper.GetString("services.vault")
		regSvcName = viper.GetString("services.registration")
		authSvcName = viper.GetString("services.auth")
		tcalSvcName = viper.GetString("services.tcal")

		tag = viper.GetString("tags.tag")

		// TODO temporary, rewrite client init
		Client, err = servicediscovery.NewDefaultConsulHTTPClient(consulAddress)
		if err != nil {
			panic(err)
		}
	}
}

func MakeHttpHandler() http.Handler {
	r := mux.NewRouter()

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "HEAD", "POST", "PUT", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
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

	//task calendar
	r.Methods("POST").HandlerFunc(tcal).Path("/taskManager/getTasks")

	handler := c.Handler(r)

	return handler
}

//task calendar
func tcal(writer http.ResponseWriter, request *http.Request) {
	addr, _, err := Client.FindService(&consulAddress, &tcalSvcName, &tag)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}
	defer request.Body.Close()
	request.Close = true

	client := &http.Client{}
	req, err := http.NewRequest("POST", addr+"/taskManager/getTasks", request.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}
	defer req.Body.Close()

	req.Header = request.Header
	resp, err := client.Do(req)

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}
	defer resp.Body.Close()
	resp.Close = true

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	writer.WriteHeader(resp.StatusCode)
	writer.Write(data)
}

//authorization
func login(writer http.ResponseWriter, request *http.Request) {
	addr, err := ServiceDiscovery().Find(&consulAddress, &authSvcName, &tag)

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}
	defer request.Body.Close()
	request.Close = true

	resp, err := http.Post(addr+"/login", "application/json; charset=utf-8", request.Body)

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}
	defer resp.Body.Close()
	resp.Close = true

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	writer.WriteHeader(resp.StatusCode)
	writer.Write(data)
}

//registration
func regvaluser(writer http.ResponseWriter, request *http.Request) {
	addr, err := svcdiscovery.ServiceDiscovery().Find(&consulAddress, &regSvcName, &tag)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}
	defer request.Body.Close()
	request.Close = true

	resp, err := http.Post(addr+"/registration/user", "application/json", request.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}
	defer resp.Body.Close()
	resp.Close = true

	data, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	//writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(resp.StatusCode)
	writer.Write(data)
}

//registration
func regvalemail(writer http.ResponseWriter, request *http.Request) {
	addr, err := svcdiscovery.ServiceDiscovery().Find(&consulAddress, &regSvcName, &tag)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}
	defer request.Body.Close()
	request.Close = true

	resp, err := http.Post(addr+"/registration/email", "application/json", request.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}
	defer resp.Body.Close()
	resp.Close = true

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	writer.WriteHeader(resp.StatusCode)
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
	defer request.Body.Close()

	resp, err := http.Post(addr+"/registration", "application/json; charset=utf-8", request.Body)

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	writer.WriteHeader(resp.StatusCode)
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
	defer request.Body.Close()
	request.Close = true

	resp, err := http.Post(addr+"/validate", "application/json", request.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}
	defer resp.Body.Close()
	resp.Close = true

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	writer.WriteHeader(resp.StatusCode)
	writer.Write(data)
}

// /hash
func hash(writer http.ResponseWriter, r *http.Request) {
	addr, err := svcdiscovery.ServiceDiscovery().Find(&consulAddress, &vaultSvcName, &tag)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	r.Close = true

	fmt.Println(addr)

	resp, err := http.Post(addr+"/hash", "application/json", r.Body)

	if err != nil {
		writer.Write([]byte(err.Error()))
		return
	}
	defer resp.Body.Close()
	resp.Close = true

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	writer.WriteHeader(resp.StatusCode)
	writer.Write(data)
}
