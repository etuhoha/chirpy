package main

import (
	"fmt"
	"net/http"
	"os"
)

type AppHandler struct {
	fileServer http.Handler
}

func (ah *AppHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {

	ah.fileServer.ServeHTTP(response, request)
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	mux.HandleFunc("/healthz", handleHelthz)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("error: %v", err)
		os.Exit(1)
	}
}

func handleHelthz(response http.ResponseWriter, req *http.Request) {
	response.Header().Set("Content-Type", "text/plain; charset=utf-8")
	response.WriteHeader(200)
	response.Write([]byte("OK"))
}
