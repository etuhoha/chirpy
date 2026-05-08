package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/etuhoha/chirpy/internal/database"
	"github.com/google/uuid"
)

func handleCreateChirp(db *database.Queries, w http.ResponseWriter, req *http.Request) {
	type requestData struct {
		Body   *string    `json:"body"`
		UserId *uuid.UUID `json:"user_id"`
	}

	reqData := requestData{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&reqData)
	if err != nil || reqData.Body == nil || reqData.UserId == nil {
		respondJsonError(w, http.StatusBadRequest, "malformed request", err)
		return
	}

	params := database.CreateChirpParams{}
	params.Body = *reqData.Body
	params.UserID = *reqData.UserId
	chirp, err := db.CreateChirp(req.Context(), params)
	if err != nil {
		respondJsonError(w, http.StatusBadRequest, "can not create chirp", err)
		return
	}

	type responseData struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt string    `json:"created_at"`
		UpdatedAt string    `json:"updated_at"`
		Body      string    `json:"body"`
		UserId    uuid.UUID `json:"user_id"`
	}

	resData := responseData{
		Id:        chirp.ID,
		CreatedAt: chirp.CreatedAt.Format(time.RFC3339),
		UpdatedAt: chirp.UpdatedAt.Format(time.RFC3339),
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	}

	respondJson(w, http.StatusCreated, resData)
}
