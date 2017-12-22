package gateway

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"io/ioutil"
	"time"
	"github.com/go-redis/redis"
)

var consulAddress string
var vaultSrvName string
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
	vaultSrvName = viper.GetString("services.vault")


	tag = viper.GetString("tags.tag")
}

func MakeHttpHandler() http.Handler {
	r := mux.NewRouter()

	//vault path
	r.Methods("POST").HandlerFunc(hash).Path("/hash")
	r.Methods("POST").HandlerFunc(validate).Path("/validate")

	return r
}

func validate(writer http.ResponseWriter, request *http.Request) {
	addr, err := GetServiceDiscovery().Find(&consulAddress, &vaultSrvName, &tag)
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
func hash(writer http.ResponseWriter, r *http.Request){
	adrFromRedis := <-getUrlInRedis("/hash")

	if adrFromRedis == "" {
		addr, err := GetServiceDiscovery().Find(&consulAddress, &vaultSrvName, &tag)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
			}


		resp, err := http.Post(addr + "/hash", "application/json", r.Body)
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

		setUrlInRedis("/hash", addr + "/hash")

	} else {
		addr := adrFromRedis
		resp, err := http.Post(addr, "application/json", r.Body)
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

}

func getUrlInRedis(serviceEndpoint string) (address <-chan string) {
	c := make(chan string)

	go func(svcName string) {
		client := redis.NewClient(
			&redis.Options{
				Addr: "localhost:6379",
				Password: "",
				DB: 0,
			})

		res, err := client.Get(serviceEndpoint).Result()
		if err != nil {
			c <- ""
		}
		c <- res
	}(serviceEndpoint)

	return c
}

func setUrlInRedis(serviceEndpoint, address string) {
	go func(svcEnp, addr string) {
		client := redis.NewClient(
			&redis.Options{
				Addr: "localhost:6379",
				Password: "",
				DB: 0,
			})
		client.Set(serviceEndpoint, address, time.Duration(time.Hour * 24))
	}(serviceEndpoint, address)
}

