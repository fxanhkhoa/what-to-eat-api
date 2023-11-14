package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
	"what-to-eat/be/firebase"
	"what-to-eat/be/graph/model"

	"github.com/golang-jwt/jwt/v5"
)

type CustomClaim struct {
	Email    string `json:"email"`
	GoogleID string `json:"google_id"`
	GithubID string `json:"github_id"`
	RoleName string `json:"role_name"`
	jwt.RegisteredClaims
}

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (a *AuthService) Login(idToken string) (*model.TokenResult, error) {
	var data model.TokenResult
	fmt.Println(idToken)
	token, err := firebase.FirebaseClient.VerifyIDToken(context.TODO(), idToken)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	queriedUser, err := firebase.FirebaseClient.GetUser(context.TODO(), token.UID)
	if err != nil {
		return nil, err
	}

	fmt.Println("quereiedUser UID", queriedUser.UID)
	user, err := NewUserService().FindUserByUID(queriedUser.UID)
	fmt.Println("user 1", user)
	if user == nil {
		user, err = NewUserService().CreateUserWithGoogle(queriedUser)
		fmt.Println("user 2", user)
		if err != nil {
			return nil, err
		}
	}

	refreshToken, err := a.generateRefreshToken(*user)
	if err != nil {
		return nil, err
	}
	accessToken, err := a.generateToken(refreshToken)
	if err != nil {
		return nil, err
	}
	data.Token = accessToken
	data.RefreshToken = refreshToken

	return &data, nil
}

func (a *AuthService) generateRefreshToken(user model.User) (string, error) {
	expireHourRefreshStr := os.Getenv("JWT_EXPIRED_REFRESH")
	secretKey := os.Getenv("SECRET_KEY")
	expireHour, errParse := time.ParseDuration(expireHourRefreshStr)
	if errParse != nil {
		return "", errParse
	}
	claims := CustomClaim{
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

func (a *AuthService) generateToken(refreshToken string) (string, error) {
	expireHourStr := os.Getenv("JWT_EXPIRED")
	secretKey := os.Getenv("SECRET_KEY")
	expireHour, err := time.ParseDuration(expireHourStr)
	if err != nil {
		return "", err
	}

	token, err := jwt.ParseWithClaims(refreshToken, &CustomClaim{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		log.Println(err)
		return "", err
	} else if claims, ok := token.Claims.(*CustomClaim); ok {
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

		newClaim := CustomClaim{
			Email:    claims.Email,
			GoogleID: *user.GoogleID,
			GithubID: *user.GithubID,
			RoleName: user.RoleName,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireHour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				NotBefore: jwt.NewNumericDate(time.Now()),
				ID:        user.ID,
			},
		}
		log.Println("newClaim", newClaim)
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaim)
		log.Println("HERE")
		ss, err := token.SignedString([]byte(secretKey))

		if err != nil {
			log.Println(err.Error())
			return "", err
		}

		return ss, err
	} else {
		log.Println(err.Error())
		return "", err
	}
}
