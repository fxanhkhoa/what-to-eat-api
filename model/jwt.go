package model

import "github.com/golang-jwt/jwt/v5"

type JwtCustomClaims struct {
	Name     string `json:"name"`
	RoleName string `json:"roleName"`
	jwt.RegisteredClaims
}

type JwtRefreshCustomClaims struct {
	jwt.RegisteredClaims
}

type TokenResult struct {
	Token        string `json:"token" bson:"token"`
	RefreshToken string `json:"refreshToken" bson:"refreshToken"`
}
