package main

import (
	"context"
	"fmt"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"path"
)

const (
	addr    = "0.0.0.0"
	certDir = "cert"

	nodeEndpointFlag = "node"
)

var (
	nodeEndpoint string
	verbose      bool
)

func main() {
	app := &cli.App{
		Name:    "liberte",
		Usage:   "proxy Ethereum RPC requests to your own node",
		Version: "1.0.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        nodeEndpointFlag,
				Aliases:     []string{"n"},
				Usage:       "Ethereum RPC that should be used",
				Value:       "http://127.0.0.1:9546",
				Destination: &nodeEndpoint,
			},
			&cli.BoolFlag{
				Name:        "verbose",
				Usage:       "Verbose logging about the requests going through",
				Value:       false,
				Destination: &verbose,
			},
		},
		Action: func(cctx *cli.Context) error {
			g, _ := errgroup.WithContext(context.Background())

			g.Go(func() error {
				mux := http.NewServeMux()
				mux.HandleFunc("/", proxyRequest)
				mux.HandleFunc("/ws/", proxyWs)

				bindAddr := addr + ":443"
				fmt.Println("Starting HTTPS reverse proxy on: " + bindAddr)

				return http.ListenAndServeTLS(
					bindAddr,
					path.Join(certDir, "infura.io.crt"),
					path.Join(certDir, "infura.io.key"),
					mux,
				)
			})

			// Just in case
			g.Go(func() error {
				mux := http.NewServeMux()
				mux.HandleFunc("/", httpsRedirect)

				bindAddr := addr + ":80"
				fmt.Println("Starting HTTP redirect on: " + bindAddr)

				return http.ListenAndServe(
					bindAddr,
					mux,
				)
			})

			return g.Wait()
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(fmt.Sprintf("Could not start %s: %s", app.Name, err.Error()))
	}
}
