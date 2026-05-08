package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/etuhoha/chirpy/internal/database"
	"github.com/google/uuid"
)

func handleCreateUser(db *database.Queries, w http.ResponseWriter, req *http.Request) {
	type requestData struct {
		Email *string `json:"email"`
	}

	reqData := requestData{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&reqData)
	if err != nil || reqData.Email == nil {
		respondJsonError(w, http.StatusBadRequest, "malformed request")
		return
	}

	user, err := db.CreateUser(req.Context(), *reqData.Email)
	if err != nil {
		respondJsonError(w, http.StatusBadRequest, fmt.Sprintf("can not create user: %v", err))
		return
	}

	type responseData struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt string    `json:"created_at"`
		UpdatedAt string    `json:"updated_at"`
		Email     string    `json:"email"`
	}

	resData := responseData{
		Id:        user.ID,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
		Email:     user.Email,
	}

	respondJson(w, http.StatusCreated, resData)
}
