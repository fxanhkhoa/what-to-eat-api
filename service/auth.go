package service

import (
	"context"
	"log"
	"time"
	"what-to-eat/be/config"
	"what-to-eat/be/firebase"
	"what-to-eat/be/model"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct{}

func (a *AuthService) Login(idToken string) (*model.TokenResult, error) {
	var data model.TokenResult
	token, err := firebase.FirebaseClient.VerifyIDToken(context.TODO(), idToken)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	queriedUser, err := firebase.FirebaseClient.GetUser(context.TODO(), token.UID)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	user, err := NewUserService().FindUserByUID(queriedUser.UID)
	if user == nil {
		user, err = NewUserService().CreateUserWithGoogle(queriedUser)
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
		log.Println(err.Error())
		return "", err
	}
}
