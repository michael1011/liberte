package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/sync/errgroup"
	"io"
	"net/http"
	"path"
)

const (
	addr            = "0.0.0.0"
	nodeEndpoint    = "http://127.0.0.1:9546"
	applicationJson = "application/json"

	certDir = "cert"
)

type errorResponse struct {
	Error string `json:"error"`
}

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

	res, err := http.Post(nodeEndpoint, applicationJson, bytes.NewBuffer(body))
	if err != nil {
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

	err := g.Wait()

	fmt.Println("Could not start liberte: " + err.Error())
}
