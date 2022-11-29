package main

import (
	"fmt"
	"golang.org/x/net/websocket"
	"io"
	"net/http"
)

func handleWebsocket(ws *websocket.Conn) {
	var msg []byte

	for {
		err := websocket.Message.Receive(ws, &msg)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Could not receive WebSocket message: " + err.Error())
			}

			_ = ws.Close()
			break
		}

		if verbose {
			fmt.Println("Got WebSocket message: " + string(msg))
		}

		_, resBody, err := nodeRequest(msg)

		if err != nil {
			_, err = ws.Write(nodeRequestFailedFormat(err.Error()))
		} else {
			_, err = ws.Write(resBody)
		}

		if err != nil {
			fmt.Println("Could not write WebSocket message: " + err.Error())
			_ = ws.Close()
			break
		}
	}
}

func proxyWs(writer http.ResponseWriter, request *http.Request) {
	s := websocket.Server{Handler: websocket.Handler(handleWebsocket)}
	s.ServeHTTP(writer, request)
}
