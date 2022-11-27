package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
	"io"
	"net/http"
	"os"
	"path"
)

const (
	addr            = "0.0.0.0"
	applicationJson = "application/json"

	certDir = "cert"

	nodeEndpointFlag = "node"
)

type errorResponse struct {
	Error string `json:"error"`
}

var (
	nodeEndpoint string
	verbose      bool
)

func proxyRequest(writer http.ResponseWriter, request *http.Request) {
	if request.TLS == nil {
		redirectHttp(writer, request)
		return
	}

	if request.Method == http.MethodOptions {
		setCorsHeader(writer)
		writer.WriteHeader(http.StatusOK)
		return
	} else if request.Method != http.MethodPost {
		handleProxyErr(writer, "invalid HTTP method")
		return
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		handleProxyErr(writer, err.Error())
		return
	}

	if verbose {
		fmt.Println("Got request: " + string(body))
	}

	res, err := http.Post(nodeEndpoint, applicationJson, bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Could not connect to node: " + err.Error())
		handleProxyErr(writer, err.Error())
		return
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		handleProxyErr(writer, err.Error())
		return
	}

	writer.Header().Set("Content-Type", applicationJson)
	setCorsHeader(writer)
	writer.WriteHeader(res.StatusCode)
	_, _ = writer.Write(resBody)
}

func redirectHttp(writer http.ResponseWriter, request *http.Request) {
	httpsUrl := "https://" + request.Host + request.URL.Path
	http.Redirect(writer, request, httpsUrl, http.StatusMovedPermanently)
}

func handleProxyErr(writer http.ResponseWriter, msg string) {
	writer.Header().Set("Content-Type", applicationJson)
	setCorsHeader(writer)
	writer.WriteHeader(http.StatusInternalServerError)

	_, err := writer.Write(jsonStringify(errorResponse{Error: "could not handle request: " + msg}))
	if err != nil {
		fmt.Println("Could not write errors response: " + err.Error())
	}
}

func setCorsHeader(writer http.ResponseWriter) {
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
	writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, X-Auth-Token")
}

func jsonStringify(msg any) []byte {
	res, _ := json.Marshal(msg)
	return res
}

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
			http.HandleFunc("/", proxyRequest)
			g, _ := errgroup.WithContext(context.Background())

			g.Go(func() error {
				bindAddr := addr + ":443"
				fmt.Println("Starting HTTPS reverse proxy on: " + bindAddr)

				return http.ListenAndServeTLS(
					bindAddr,
					path.Join(certDir, "infura.io.crt"),
					path.Join(certDir, "infura.io.key"),
					nil,
				)
			})

			// Just in case
			g.Go(func() error {
				bindAddr := addr + ":80"
				fmt.Println("Starting HTTP redirect on: " + bindAddr)

				return http.ListenAndServe(
					bindAddr,
					nil,
				)
			})

			return g.Wait()
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(fmt.Sprintf("Could not start %s: %s", app.Name, err.Error()))
	}
}
