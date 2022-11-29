package main

import "net/http"

func httpsRedirect(writer http.ResponseWriter, request *http.Request) {
	httpsUrl := "https://" + request.Host + request.URL.Path
	http.Redirect(writer, request, httpsUrl, http.StatusMovedPermanently)
}
