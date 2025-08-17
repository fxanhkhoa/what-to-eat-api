package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	"what-to-eat/be/config"
	"what-to-eat/be/model"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

type AuthService struct{}

// Login verifies Google ID token and authenticates the user
func (a *AuthService) Login(googleIdToken string) (*model.TokenResult, error) {
	var data model.TokenResult

	userInfo, err := a.verifyIdToken(googleIdToken)
	if err != nil {
		log.Println("Failed to verify ID token:", err)
		return nil, err
	}

	user, err := NewUserService().FindUserByUID(userInfo.Id)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Println(err.Error())
		return nil, err
	}
	if user == nil {
		// Create user with Google info
		user, err = NewUserService().CreateUserWithGoogleFromOAuth(userInfo)
		if err != nil {
			log.Println(err.Error())
			return nil, err
		}
	}

	refreshToken, err := a.GenerateRefreshToken(*user)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	accessToken, err := a.GenerateToken(refreshToken)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	data.Token = accessToken
	data.RefreshToken = refreshToken

	return &data, nil
}

func (a *AuthService) verifyIdToken(idToken string) (*oauth2.Userinfo, error) {
	// First verify the token with Google
	var httpClient = &http.Client{}
	oauth2Service, err := oauth2.NewService(context.TODO(), option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}

	// Verify the ID token and get token info
	tokenInfoCall := oauth2Service.Tokeninfo()
	tokenInfoCall.IdToken(idToken)
	tokenInfo, err := tokenInfoCall.Do()
	if err != nil {
		return nil, err
	}

	// Parse the JWT to extract additional user information
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format")
	}

	// Decode the payload (second part)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT payload: %v", err)
	}

	var claims model.GoogleClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JWT claims: %v", err)
	}

	// Create userinfo from both token info and JWT claims
	verifiedEmail := tokenInfo.VerifiedEmail
	userInfo := &oauth2.Userinfo{
		Id:            tokenInfo.UserId,
		Email:         tokenInfo.Email,
		VerifiedEmail: &verifiedEmail,
		Name:          claims.Name,
		GivenName:     claims.GivenName,
		FamilyName:    claims.FamilyName,
		Picture:       claims.Picture,
	}

	return userInfo, nil
}

func (a *AuthService) GenerateRefreshToken(user model.User) (string, error) {
	expireHourRefreshStr := config.GetInstanceConfig().JWTRefreshExpired
	secretKey := config.GetInstanceConfig().JWTSecret
	expireHour, errParse := time.ParseDuration(expireHourRefreshStr)
	if errParse != nil {
		return "", errParse
	}
	claims := model.JwtCustomClaims{
		Email:    user.Email,
		GoogleID: *new(string),
		GithubID: *new(string),
		RoleName: *new(string),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireHour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID:        user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(secretKey))

	if err != nil {
		log.Printf("Error signing refresh token: %s \n", err.Error())
		return "", err
	}

	return ss, err
}

func (a *AuthService) GenerateToken(refreshToken string) (string, error) {
	expireHourStr := config.GetInstanceConfig().JWTExpired
	secretKey := config.GetInstanceConfig().JWTSecret
	expireHour, err := time.ParseDuration(expireHourStr)
	if err != nil {
		return "", err
	}

	token, err := jwt.ParseWithClaims(refreshToken, &model.JwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		log.Println(err)
		return "", err
	} else if claims, ok := token.Claims.(*model.JwtCustomClaims); ok {
		user, err := NewUserService().FindByID(claims.ID)

		if err != nil {
			log.Println(err.Error())
			return "", err
		}

		if user.GithubID == nil {
			user.GithubID = new(string)
		}

		if user.GoogleID == nil {
			user.GoogleID = new(string)
		}

		newClaim := model.JwtCustomClaims{
			Email:    claims.Email,
			GoogleID: *user.GoogleID,
			GithubID: *user.GithubID,
			RoleName: user.RoleName,
			Name:     *user.Name,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireHour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				NotBefore: jwt.NewNumericDate(time.Now()),
				ID:        user.ID,
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaim)
		ss, err := token.SignedString([]byte(secretKey))

		if err != nil {
			log.Println(err.Error())
			return "", err
		}

		return ss, err
	} else {
		log.Println("Invalid token claims")
		return "", fmt.Errorf("invalid token claims")
	}
}
