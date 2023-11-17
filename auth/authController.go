package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"what-to-eat/be/graph/model"
	"what-to-eat/be/graph/service"
)

type AuthController struct{}

func NewAuthController() *AuthController {
	return &AuthController{}
}

func (au *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var loginInput model.LoginInput
	err := json.NewDecoder(r.Body).Decode(&loginInput)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	data, err := service.NewAuthService().Login(loginInput.Token)
	if err != nil {
		log.Printf("Login Error: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	decoded, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(decoded)
}

func (au *AuthController) RetrieveToken(w http.ResponseWriter, r *http.Request) {
	var retrieveTokenInput model.RetrieveTokenInput
	err := json.NewDecoder(r.Body).Decode(&retrieveTokenInput)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	token, err := service.NewAuthService().GenerateToken(retrieveTokenInput.RefreshToken)
	data := model.TokenResult{
		Token:        token,
		RefreshToken: retrieveTokenInput.RefreshToken,
	}
	decoded, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(decoded)
}
