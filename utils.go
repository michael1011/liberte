package main

import (
	"encoding/json"
	"net/http"
)

func setCorsHeader(writer http.ResponseWriter) {
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
	writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, X-Auth-Token")
}

func jsonStringify(msg any) []byte {
	res, _ := json.Marshal(msg)
	return res
}
