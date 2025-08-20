package model

import "github.com/golang-jwt/jwt/v5"

type LoginDto struct {
	Token string `json:"token" bson:"token"`
	Type  string `json:"type" bson:"type" default:"google"` // 'google' or 'apple', default is 'google'
}

// GoogleUserInfo represents the user info returned by Google OAuth2 API
type GoogleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Locale        string `json:"locale"`
}

type GoogleClaims struct {
	Iss           string `json:"iss"`
	Sub           string `json:"sub"`
	Aud           string `json:"aud"`
	Exp           int64  `json:"exp"`
	Iat           int64  `json:"iat"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Locale        string `json:"locale"`
	jwt.RegisteredClaims
}

// AppleClaims represents the claims in an Apple ID token
type AppleClaims struct {
	Iss           string `json:"iss"`
	Aud           string `json:"aud"`
	Exp           int64  `json:"exp"`
	Iat           int64  `json:"iat"`
	Sub           string `json:"sub"`
	Email         string `json:"email,omitempty"`
	EmailVerified bool   `json:"email_verified,omitempty"`
	Name          struct {
		FirstName string `json:"firstName,omitempty"`
		LastName  string `json:"lastName,omitempty"`
	} `json:"name,omitempty"`
	jwt.RegisteredClaims
}

// AppleUserInfo represents processed user info from Apple ID token
type AppleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"email_verified"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
}

// AppleJWKS represents Apple's JSON Web Key Set
type AppleJWKS struct {
	Keys []AppleJWK `json:"keys"`
}

type AppleJWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}
