package gateway

import (
	"io/ioutil"
	"net/http"
)

func NewCollectorHandler(d *Collector) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		payload, err := ioutil.ReadAll(request.Body)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := d.Collect(string(payload)); err != nil {
			writer.WriteHeader(http.StatusServiceUnavailable)
			return
		}
	}
}
