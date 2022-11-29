package main

import (
	"fmt"
	"io"
	"net/http"
)

const (
	applicationJson = "application/json"
)

type errorResponse struct {
	Error string `json:"error"`
}

func proxyRequest(writer http.ResponseWriter, request *http.Request) {
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

	res, resBody, err := nodeRequest(body)
	if err != nil {
		handleProxyErr(writer, err.Error())
		return
	}

	writer.Header().Set("Content-Type", applicationJson)
	setCorsHeader(writer)
	writer.WriteHeader(res.StatusCode)
	_, _ = writer.Write(resBody)
}

func handleProxyErr(writer http.ResponseWriter, msg string) {
	writer.Header().Set("Content-Type", applicationJson)
	setCorsHeader(writer)
	writer.WriteHeader(http.StatusInternalServerError)

	_, err := writer.Write(nodeRequestFailedFormat(msg))
	if err != nil {
		fmt.Println("Could not write errors response: " + err.Error())
	}
}
