package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/issy20/go-websocket/auth"
	"github.com/issy20/go-websocket/repository"
)

type LoginUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserInput struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type API struct {
	UserRepository *repository.UserRepository
}

func (api *API) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var user LoginUser

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	dbUser := api.UserRepository.FindUserByUsername(user.Username)
	if dbUser == nil {
		returnErrorResponse(w)
		return
	}

	ok, err := auth.ComparePassword(user.Password, dbUser.Password)

	if !ok || err != nil {
		returnErrorResponse(w)
		return
	}

	token, err := auth.CreateJWTToken(dbUser)
	if err != nil {
		returnErrorResponse(w)
		return
	}

	w.Write([]byte(token))
}

func (api *API) HandleAddUser(w http.ResponseWriter, r *http.Request) {
	var userInput *UserInput
	err := json.NewDecoder(r.Body).Decode(&userInput)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	hashedPassword, err := auth.GeneratePassword(userInput.Password)

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	id := uuid.New().String()
	name := userInput.Name
	username := userInput.Username
	password := hashedPassword

	if err := api.UserRepository.AddUser(id, name, username, password); err != nil {
		return
	}
}

func returnErrorResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"status\": \"error\"}"))
}
