package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/etuhoha/chirpy/internal/auth"
	"github.com/etuhoha/chirpy/internal/database"
	"github.com/google/uuid"
)

type responseUser struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Email     string    `json:"email"`
}

func handleCreateUser(db *database.Queries, w http.ResponseWriter, req *http.Request) {
	type requestData struct {
		Email    *string `json:"email"`
		Password *string `json:"password"`
	}

	reqData := requestData{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&reqData)
	if err != nil || reqData.Email == nil || reqData.Password == nil {
		respondJsonError(w, http.StatusBadRequest, "malformed request", err)
		return
	}

	params := database.CreateUserParams{}
	params.Email = *reqData.Email
	params.HashedPassword, err = auth.HashPassword(*reqData.Password)
	if err != nil {
		respondInternalError(w, err)
		return
	}
	user, err := db.CreateUser(req.Context(), params)
	if err != nil {
		respondJsonError(w, http.StatusBadRequest, "can not create user", err)
		return
	}

	resData := responseUser{
		Id:        user.ID,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
		Email:     user.Email,
	}

	respondJson(w, http.StatusCreated, resData)
}

func handleLogin(db *database.Queries, w http.ResponseWriter, req *http.Request) {
	type requestData struct {
		Email    *string `json:"email"`
		Password *string `json:"password"`
	}

	reqData := requestData{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&reqData)
	if err != nil || reqData.Email == nil || reqData.Password == nil {
		respondJsonError(w, http.StatusBadRequest, "malformed request", err)
		return
	}

	user, err := db.GetUserByEmail(req.Context(), *reqData.Email)
	if err != nil {
		respondJsonError(w, http.StatusBadRequest, "can not find user", err)
		return
	}

	ok, err := auth.CheckPasswordHash(*reqData.Password, user.HashedPassword)
	if err != nil {
		respondInternalError(w, err)
		return
	}

	if !ok {
		respondJsonError(w, http.StatusUnauthorized, "incorrect email or password", nil)
		return
	}

	resData := responseUser{
		Id:        user.ID,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
		Email:     user.Email,
	}

	respondJson(w, http.StatusOK, resData)
}
