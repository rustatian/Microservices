package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"github.com/go-kit/kit/log"
)

func main() {
	var (
		httpAddr     = flag.String("http.addr", ":8000", "Address for HTTP (JSON) server")
	)

	flag.Parse()


	// Logging domain.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}


	r := MakeHttpHandler()

	// Interrupt handler.
	errc := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// HTTP transport.
	go func() {
		logger.Log("transport", "HTTP", "addr", *httpAddr)
		errc <- http.ListenAndServe(*httpAddr, r)
	}()

	// Run!
	logger.Log("exit", <- errc)
}
