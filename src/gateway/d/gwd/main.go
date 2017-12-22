package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"github.com/go-kit/kit/log"
	"TaskManager/src/gateway"
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
		errc <- http.ListenAndServe(*httpAddr, handler)
	}()


	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	chErr := <- errc
	logger.Log("exit", chErr)
}
