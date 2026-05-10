package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/etuhoha/chirpy/internal/auth"
	"github.com/etuhoha/chirpy/internal/database"
	"github.com/google/uuid"
)

type responseChirp struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    uuid.UUID `json:"user_id"`
}

func handleCreateChirp(config *apiConfig, w http.ResponseWriter, req *http.Request) {
	type requestData struct {
		Body *string `json:"body"`
	}

	reqData := requestData{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&reqData)
	if err != nil || reqData.Body == nil {
		respondJsonError(w, http.StatusBadRequest, "malformed request", err)
		return
	}

	userId, err := auth.AuthenticateByJWT(req.Header, config.authSecret)
	if err != nil {
		respondJsonError(w, http.StatusUnauthorized, "authentication error", err)
		return
	}

	params := database.CreateChirpParams{}
	params.Body = *reqData.Body
	params.UserID = userId
	chirp, err := config.db.CreateChirp(req.Context(), params)
	if err != nil {
		respondJsonError(w, http.StatusBadRequest, "can not create chirp", err)
		return
	}

	resData := responseChirp{
		Id:        chirp.ID,
		CreatedAt: chirp.CreatedAt.Format(time.RFC3339),
		UpdatedAt: chirp.UpdatedAt.Format(time.RFC3339),
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	}

	respondJson(w, http.StatusCreated, resData)
}

func handleGetChirps(config *apiConfig, w http.ResponseWriter, req *http.Request) {
	var authorId *uuid.UUID = nil

	queryValues := req.URL.Query()
	authorIdStr := queryValues.Get("author_id")
	if len(authorIdStr) > 0 {
		aId, err := uuid.Parse(authorIdStr)
		if err != nil {
			respondJsonError(w, http.StatusBadRequest, "malformed author ID", err)
			return
		}

		authorId = &aId
	}

	var chirps []database.Chirp
	var err error
	if authorId != nil {
		chirps, err = config.db.GetChirpsByUser(req.Context(), *authorId)
	} else {
		chirps, err = config.db.GetChirps(req.Context())
	}

	if err != nil {
		respondJsonError(w, http.StatusBadRequest, "can't get chirps", err)
		return
	}

	result := []responseChirp{}
	for _, chirp := range chirps {
		chirpData := responseChirp{
			Id:        chirp.ID,
			CreatedAt: chirp.CreatedAt.Format(time.RFC3339),
			UpdatedAt: chirp.UpdatedAt.Format(time.RFC3339),
			Body:      chirp.Body,
			UserId:    chirp.UserID,
		}
		result = append(result, responseChirp(chirpData))
	}

	respondJson(w, http.StatusOK, result)
}

func handleGetChirp(config *apiConfig, w http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("chirpID")
	if len(idStr) == 0 {
		respondJsonError(w, http.StatusBadRequest, "no chirp id provided", nil)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		respondJsonError(w, http.StatusBadRequest, "bad ID", err)
		return
	}

	chirp, err := config.db.GetChirp(req.Context(), id)
	if err != nil {
		respondJsonError(w, http.StatusNotFound, "can't get chirp", err)
		return
	}

	chirpData := responseChirp{
		Id:        chirp.ID,
		CreatedAt: chirp.CreatedAt.Format(time.RFC3339),
		UpdatedAt: chirp.UpdatedAt.Format(time.RFC3339),
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	}

	respondJson(w, http.StatusOK, chirpData)
}

func handleDeleteChirp(config *apiConfig, w http.ResponseWriter, req *http.Request) {
	userId, err := auth.AuthenticateByJWT(req.Header, config.authSecret)
	if err != nil {
		respondJsonError(w, http.StatusUnauthorized, "authentication error", err)
		return
	}

	idStr := req.PathValue("chirpID")
	if len(idStr) == 0 {
		respondJsonError(w, http.StatusBadRequest, "no chirp id provided", nil)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		respondJsonError(w, http.StatusBadRequest, "bad ID", err)
		return
	}

	chirp, err := config.db.GetChirp(req.Context(), id)
	if err != nil {
		respondJsonError(w, http.StatusNotFound, "can't get chirp", err)
		return
	}

	if chirp.UserID != userId {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	err = config.db.DeleteChirp(req.Context(), id)
	if err != nil {
		respondJsonError(w, http.StatusNotFound, "can't delete chirp", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
