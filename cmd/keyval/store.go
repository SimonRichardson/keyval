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
	tcpStore "github.com/SimonRichardson/keyval/pkg/tcp"
	udpStore "github.com/SimonRichardson/keyval/pkg/udp"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

func runStore(args []string) error {
	var (
		flags = flag.NewFlagSet("store", flag.ExitOnError)

		debug       = flags.Bool("debug", false, "debug logging")
		apiHTTPAddr = flags.String("api.http", defaultAPIHTTPAddr, "listen address for HTTP API")
		apiTCPAddr  = flags.String("api.tcp", defaultAPITCPAddr, "listen address for TCP API")
		apiUDPAddr  = flags.String("api.udp", defaultAPIUDPAddr, "listen address for UDP API")
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

	// Setup http api
	apiHTTPNetwork, apiHTTPAddress, err := parseAddr(*apiHTTPAddr, defaultAPIHTTPPort)
	if err != nil {
		return err
	}
	apiHTTPListener, err := net.Listen(apiHTTPNetwork, apiHTTPAddress)
	if err != nil {
		return err
	}

	level.Debug(logger).Log("HTTP_API", fmt.Sprintf("%s://%s", apiHTTPNetwork, apiHTTPAddress))

	// Setup tcp api
	apiTCPNetwork, apiTCPAddress, err := parseAddr(*apiTCPAddr, defaultAPITCPPort)
	if err != nil {
		return err
	}
	apiTCPListener, err := net.Listen(apiTCPNetwork, apiTCPAddress)
	if err != nil {
		return err
	}

	level.Debug(logger).Log("TCP_API", fmt.Sprintf("%s://%s", apiTCPNetwork, apiTCPAddress))

	// Setup udp api
	apiUDPNetwork, apiUDPAddress, err := parseAddr(*apiUDPAddr, defaultAPIUDPPort)
	if err != nil {
		return err
	}
	udpAddr, err := net.ResolveUDPAddr(apiUDPNetwork, apiUDPAddress)
	if err != nil {
		return err
	}
	apiUDPListener, err := net.ListenUDP(apiUDPNetwork, udpAddr)
	if err != nil {
		return err
	}

	level.Debug(logger).Log("UDP_API", fmt.Sprintf("%s://%s", apiUDPNetwork, apiUDPAddress))

	// Setup store api
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

			return http.Serve(apiHTTPListener, mux)
		}, func(error) {
			apiHTTPListener.Close()
		})
	}
	{
		g.Add(func() error {
			server := tcpStore.NewServer(
				keyval,
				log.With(logger, "component", "store_tcp_api"),
			)
			return server.Serve(apiTCPListener)
		}, func(error) {
			apiTCPListener.Close()
		})
	}
	{
		g.Add(func() error {
			server := udpStore.NewServer(
				keyval,
				log.With(logger, "component", "store_udp_api"),
			)
			defer server.Stop()

			return server.Serve(apiUDPListener)
		}, func(error) {
			apiTCPListener.Close()
		})
	}
	gexec.Interrupt(g)
	return g.Run()
}
