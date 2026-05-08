package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func respondJsonError(w http.ResponseWriter, code int, errorMessage string, err error) {
	if err != nil {
		errorMessage = fmt.Sprintf("%s: %v", errorMessage, err)
	}

	type errResp struct {
		Error string `json:"error"`
	}

	respondJson(w, code, errResp{Error: errorMessage})
}

func respondJson(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	data, err := json.Marshal(payload)
	if err != nil {
		respondInternalError(w, err)
		return
	}

	w.WriteHeader(code)
	w.Write(data)
}

func respondInternalError(w http.ResponseWriter, err error) {
	fmt.Printf("INTERNAL ERROR: %v", err)
	w.WriteHeader(http.StatusInternalServerError)
}
