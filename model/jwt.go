package model

import "github.com/golang-jwt/jwt/v5"

type JwtCustomClaims struct {
	Email    string `json:"email"`
	GoogleID string `json:"google_id"`
	GithubID string `json:"github_id"`
	RoleName string `json:"role_name"`
	jwt.RegisteredClaims
}

type JwtRefreshCustomClaims struct {
	jwt.RegisteredClaims
}

type TokenResult struct {
	Token        string `json:"token" bson:"token"`
	RefreshToken string `json:"refreshToken" bson:"refreshToken"`
}
