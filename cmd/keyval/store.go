package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/SimonRichardson/gexec"
	httpStore "github.com/SimonRichardson/keyval/pkg/http"
	"github.com/SimonRichardson/keyval/pkg/store"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

func runStore(args []string) error {
	var (
		flags = flag.NewFlagSet("store", flag.ExitOnError)

		debug   = flags.Bool("debug", false, "debug logging")
		apiAddr = flags.String("api", defaultAPIAddr, "listen address for query API")
	)

	flags.Usage = usageFor(flags, "store [flags]")
	if err := flags.Parse(args); err != nil {
		return nil
	}

	// Setup the logger.
	var logger log.Logger
	{
		logLevel := level.AllowInfo()
		if *debug {
			logLevel = level.AllowAll()
		}
		logger = log.NewLogfmtLogger(os.Stdout)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = level.NewFilter(logger, logLevel)
	}

	apiNetwork, apiAddress, err := parseAddr(*apiAddr, defaultAPIPort)
	if err != nil {
		return err
	}
	apiListener, err := net.Listen(apiNetwork, apiAddress)
	if err != nil {
		return err
	}

	level.Debug(logger).Log("API", fmt.Sprintf("%s://%s", apiNetwork, apiAddress))

	keyval := store.New()

	// Execution group.
	g := gexec.NewGroup()
	gexec.Block(g)
	{
		g.Add(func() error {
			mux := http.NewServeMux()
			mux.Handle("/store/", http.StripPrefix("/store",
				httpStore.NewAPI(
					keyval,
					log.With(logger, "component", "store_http_api"),
				),
			))

			return http.Serve(apiListener, mux)
		}, func(error) {
			apiListener.Close()
		})
	}
	gexec.Interrupt(g)
	return g.Run()
}
