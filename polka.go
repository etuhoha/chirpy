package main

import (
	"encoding/json"
	"net/http"

	"github.com/etuhoha/chirpy/internal/auth"
	"github.com/etuhoha/chirpy/internal/database"
	"github.com/google/uuid"
)

func handlePolkaWebhook(config *apiConfig, w http.ResponseWriter, req *http.Request) {
	key, err := auth.GetAPIKey(req.Header)
	if err != nil || key != config.polkaKey {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	type requestData struct {
		Event string `json:"event"`
		Data  struct {
			UserId uuid.UUID `json:"user_id"`
		}
	}

	reqData := requestData{}
	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&reqData)
	if err != nil {
		respondJsonError(w, http.StatusBadRequest, "malformed request", err)
		return
	}

	if reqData.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	params := database.SetRedUserParams{ID: reqData.Data.UserId, IsChirpyRed: true}
	_, err = config.db.SetRedUser(req.Context(), params)
	if err != nil {
		respondJsonError(w, http.StatusNotFound, "upgrade error", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
