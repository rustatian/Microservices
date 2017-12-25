package gateway

import (
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
)

var consulAddress string
var vaultSvcName string
var regSvcName string
var tag string

func init() {
	viper.AddConfigPath("src/gateway/config")
	viper.SetConfigName("gw_conf")

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	consulAddress = viper.GetString("consul.address")

	//Service names
	vaultSvcName = viper.GetString("services.vault")
	regSvcName = viper.GetString("services.registration")

	tag = viper.GetString("tags.tag")
}

func MakeHttpHandler() http.Handler {
	r := mux.NewRouter()

	//vault path
	r.Methods("POST").HandlerFunc(hash).Path("/hash")
	r.Methods("POST").HandlerFunc(validate).Path("/validate")

	//reg path
	r.Methods("POST").HandlerFunc(registration).Path("/registration")

	return r
}

// /registration
func registration(writer http.ResponseWriter, request *http.Request) {

	addr, err := GetServiceDiscovery().Find(&consulAddress, &regSvcName, &tag)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	resp, err := http.Post(addr+"/registration", "application/json", request.Body)
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
	addr, err := GetServiceDiscovery().Find(&consulAddress, &vaultSvcName, &tag)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	resp, err := http.Post(addr+"/validate", "application/json", request.Body)
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
	addr, err := GetServiceDiscovery().Find(&consulAddress, &vaultSvcName, &tag)

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
