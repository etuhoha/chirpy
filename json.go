package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func respondJsonError(w http.ResponseWriter, code int, errorMessage string) {
	type errResp struct {
		Error string `json:"error"`
	}

	respondJson(w, code, errResp{Error: errorMessage})
}

func respondJson(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	data, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("error marshalling: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(code)
	w.Write(data)
}
