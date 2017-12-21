package main

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"github.com/prometheus/common/log"
)

var consulAddress *string
var hashSrvName *string
var tag *string

func init() {
	viper.AddConfigPath("")

}

func MakeHttpHandler() http.Handler {
	r := mux.NewRouter()

	r.Methods("POST").HandlerFunc(hash).Path("/hash")

	return r
}

func hash(w http.ResponseWriter, r *http.Request){

	addr, err := GetServiceDiscovery().Find(consulAddress, hashSrvName, tag)
	if err != nil {
		log.Error(err)
	}


}
