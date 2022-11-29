package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

func nodeRequest(body []byte) (*http.Response, []byte, error) {
	res, err := http.Post(nodeEndpoint, applicationJson, bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Could not connect to node: " + err.Error())
		return nil, nil, err
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, nil, err
	}

	return res, resBody, nil
}

func nodeRequestFailedFormat(msg string) []byte {
	return jsonStringify(errorResponse{Error: "could not handle request: " + msg})
}
