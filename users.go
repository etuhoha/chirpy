package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/etuhoha/chirpy/internal/auth"
	"github.com/etuhoha/chirpy/internal/database"
	"github.com/google/uuid"
)

var accessTokenExpireIn time.Duration = 1 * time.Hour

type responseUser struct {
	Id          uuid.UUID `json:"id"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

func handleCreateUser(config *apiConfig, w http.ResponseWriter, req *http.Request) {
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
	user, err := config.db.CreateUser(req.Context(), params)
	if err != nil {
		respondJsonError(w, http.StatusBadRequest, "can not create user", err)
		return
	}

	resData := responseUser{
		Id:          user.ID,
		CreatedAt:   user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   user.UpdatedAt.Format(time.RFC3339),
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	}

	respondJson(w, http.StatusCreated, resData)
}

func handleUpdateUser(config *apiConfig, w http.ResponseWriter, req *http.Request) {
	type requestData struct {
		Email    *string `json:"email"`
		Password *string `json:"password"`
	}

	userId, err := auth.AuthenticateByJWT(req.Header, config.authSecret)
	if err != nil {
		respondJsonError(w, http.StatusUnauthorized, "authentication error", err)
		return
	}

	reqData := requestData{}
	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&reqData)
	if err != nil || reqData.Email == nil || reqData.Password == nil {
		respondJsonError(w, http.StatusBadRequest, "malformed request", err)
		return
	}

	params := database.UpdateUserParams{}
	params.ID = userId
	params.Email = *reqData.Email
	params.HashedPassword, err = auth.HashPassword(*reqData.Password)
	if err != nil {
		respondInternalError(w, err)
		return
	}
	user, err := config.db.UpdateUser(req.Context(), params)
	if err != nil {
		respondJsonError(w, http.StatusBadRequest, "can not update user", err)
		return
	}

	resData := responseUser{
		Id:          user.ID,
		CreatedAt:   user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   user.UpdatedAt.Format(time.RFC3339),
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	}

	respondJson(w, http.StatusOK, resData)
}

func handleLogin(config *apiConfig, w http.ResponseWriter, req *http.Request) {
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

	user, err := config.db.GetUserByEmail(req.Context(), *reqData.Email)
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

	params := database.CreateRefreshTokenParams{}
	params.Token = auth.MakeRefreshToken()
	params.UserID = user.ID
	refreshToken, err := config.db.CreateRefreshToken(req.Context(), params)
	if err != nil {
		respondInternalError(w, err)
		return
	}

	token, err := auth.MakeJWT(user.ID, config.authSecret, accessTokenExpireIn)
	if err != nil {
		respondInternalError(w, err)
		return
	}

	type responseData struct {
		responseUser
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	resData := responseData{
		responseUser: responseUser{
			Id:          user.ID,
			CreatedAt:   user.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   user.UpdatedAt.Format(time.RFC3339),
			Email:       user.Email,
			IsChirpyRed: user.IsChirpyRed,
		},
		Token:        token,
		RefreshToken: refreshToken.Token,
	}

	respondJson(w, http.StatusOK, resData)
}

func handleRefresh(config *apiConfig, w http.ResponseWriter, req *http.Request) {
	refTokenStr, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondJsonError(w, http.StatusNotFound, "no refresh token", err)
		return
	}

	refreshToken, err := config.db.GetRefreshToken(req.Context(), refTokenStr)
	if err != nil {
		respondJsonError(w, http.StatusNotFound, "unknown refresh token", nil)
		return
	}

	if refreshToken.RevokedAt.Valid || refreshToken.ExpiresAt.Before(time.Now()) {
		respondJsonError(w, http.StatusUnauthorized, "token expired/revoked", nil)
		return
	}

	token, err := auth.MakeJWT(refreshToken.UserID, config.authSecret, accessTokenExpireIn)
	if err != nil {
		respondInternalError(w, err)
		return
	}

	type responseData struct {
		Token string `json:"token"`
	}

	respondJson(w, http.StatusOK, responseData{Token: token})
}

func handleRevoke(config *apiConfig, w http.ResponseWriter, req *http.Request) {
	refTokenStr, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondJsonError(w, http.StatusNotFound, "no refresh token", err)
		return
	}

	err = config.db.RevokeRefreshToken(req.Context(), refTokenStr)
	if err != nil {
		respondInternalError(w, err)
	}

	w.WriteHeader(http.StatusNoContent)
}
