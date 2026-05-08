package main

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
)

func filterProfanity(message string) string {
	badWords := []string{"kerfuffle", "sharbert", "fornax"}

	filtered := []string{}

	for _, s := range strings.Split(message, " ") {
		if slices.Contains(badWords, strings.ToLower(s)) {
			s = "****"
		}
		filtered = append(filtered, s)
	}

	return strings.Join(filtered, " ")
}

func handleValidateChirp(w http.ResponseWriter, req *http.Request) {
	type reqData struct {
		Body *string `json:"body"`
	}

	type respValid struct {
		CleanedBody string `json:"cleaned_body"`
	}

	data := reqData{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&data)
	if err != nil || data.Body == nil {
		respondJsonError(w, http.StatusBadRequest, "malformed request", err)
		return
	}

	if len(*data.Body) > 140 {
		respondJsonError(w, http.StatusBadRequest, "chirp is too long", nil)
		return
	}

	filtered := filterProfanity(*data.Body)
	respondJson(w, http.StatusOK, respValid{CleanedBody: filtered})
}
