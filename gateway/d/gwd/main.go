package main

import (
	"TaskManager/gateway"
	"flag"
	"fmt"
	"github.com/go-kit/kit/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var (
		httpAddr = flag.String("http.addr", ":8000", "Address for HTTP server")
	)

	flag.Parse()

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	errc := make(chan error)

	r := gateway.MakeHttpHandler()

	// HTTP transport
	go func() {
		logger.Log("transport", "HTTP", "addr", *httpAddr)
		handler := r
		errc <- http.ListenAndServeTLS(*httpAddr,"../config/fullchain.pem","../config/privkey.pem", handler)
	}()

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	chErr := <-errc
	logger.Log("exit", chErr)
}
