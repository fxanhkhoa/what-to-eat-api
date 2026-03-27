package service

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"
	"what-to-eat/be/config"
	"what-to-eat/be/model"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

// userAuthProvider is the subset of UserService methods that AuthService needs.
type userAuthProvider interface {
	FindUserByAppleID(appleID string) (*model.User, error)
	FindUserByUID(googleID string) (*model.User, error)
	CreateUserWithAppleFromOAuth(userInfo *model.AppleUserInfo) (*model.User, error)
	CreateUserWithGoogleFromOAuth(userInfo *oauth2.Userinfo) (*model.User, error)
	FindByID(id string) (*model.User, error)
}

// blacklistProvider is the subset of RefreshTokenBlackListService methods that AuthService needs.
type blacklistProvider interface {
	Create(token model.RefreshTokenBlackList) (primitive.ObjectID, error)
	GetByToken(token string) (*model.RefreshTokenBlackList, error)
}

// AuthService holds optional injected dependencies for testing.
type AuthService struct {
	userSvc      userAuthProvider
	blacklistSvc blacklistProvider
	appleJWKSURL string
}

func (a *AuthService) getUserSvc() userAuthProvider {
	if a.userSvc != nil {
		return a.userSvc
	}
	return NewUserService()
}

func (a *AuthService) getBlacklistSvc() blacklistProvider {
	if a.blacklistSvc != nil {
		return a.blacklistSvc
	}
	return NewRefreshTokenBlackListService(nil)
}

func (a *AuthService) getAppleJWKSURL() string {
	if a.appleJWKSURL != "" {
		return a.appleJWKSURL
	}
	return "https://appleid.apple.com/auth/keys"
}

// Login verifies Google ID token and authenticates the user
func (a *AuthService) Login(loginDto model.LoginDto, c echo.Context) (*model.TokenResult, error) {
	var data model.TokenResult
	var user *model.User
	var err error

	if loginDto.Type == "apple" {
		// Handle Apple login
		userInfo, err := a.verifyAppleIdToken(loginDto.Token)
		if err != nil {
			log.Println("Failed to verify Apple ID token:", err)
			return nil, err
		}

		user, err = a.getUserSvc().FindUserByAppleID(userInfo.Sub)
		if err != nil && err != mongo.ErrNoDocuments {
			log.Println(err.Error())
			return nil, err
		}

		if user == nil {
			// Create user with Apple info
			user, err = a.getUserSvc().CreateUserWithAppleFromOAuth(userInfo)
			if err != nil {
				log.Println(err.Error())
				return nil, err
			}
		}
	} else {
		userInfo, err := a.verifyIdToken(loginDto.Token)
		if err != nil {
			log.Println("Failed to verify ID token:", err)
			return nil, err
		}

		user, err = a.getUserSvc().FindUserByUID(userInfo.Id)
		if err != nil && err != mongo.ErrNoDocuments {
			log.Println(err.Error())
			return nil, err
		}

		if user == nil {
			// Create user with Google info
			user, err = a.getUserSvc().CreateUserWithGoogleFromOAuth(userInfo)
			if err != nil {
				log.Println(err.Error())
				return nil, err
			}
		}
	}

	// Track user login
	if user != nil {
		userId := user.ID
		ip := c.RealIP() // Default to server-detected IP

		// If client sends IP in request body, use that instead
		if loginDto.IP != "" {
			ip = loginDto.IP
		}

		userAgent := c.Request().UserAgent()
		_ = (&UserLoginTrackService{}).TrackLogin(userId, ip, userAgent)
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

func (a *AuthService) Logout(refreshToken string, profile *model.JwtCustomClaims) error {
	// Parse the refresh token to get user ID
	fmt.Println(refreshToken)
	secretKey := config.GetInstanceConfig().JWTSecret
	token, err := jwt.ParseWithClaims(refreshToken, &model.JwtRefreshCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		log.Println("Failed to parse refresh token:", err)
		return err
	}

	if claims, ok := token.Claims.(*model.JwtRefreshCustomClaims); ok && token.Valid && claims.ID == profile.ID {
		userID := claims.ID
		userObjectID, err := primitive.ObjectIDFromHex(userID)
		// Add the refresh token to the blacklist
		service := a.getBlacklistSvc()
		_, err = service.Create(model.RefreshTokenBlackList{
			Token:     refreshToken,
			UserID:    userObjectID,
			CreatedAt: time.Now(),
		})
		if err != nil {
			log.Println("Failed to blacklist refresh token:", err)
			return err
		}
		return nil
	}

	log.Println("Invalid refresh token claims")
	return fmt.Errorf("invalid refresh token claims")
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

func (a *AuthService) verifyAppleIdToken(idToken string) (*model.AppleUserInfo, error) {
	// Parse the JWT header to get the kid (key ID)
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format")
	}

	// Decode the header
	headerData, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT header: %v", err)
	}

	var header struct {
		Kid string `json:"kid"`
		Alg string `json:"alg"`
	}
	if err := json.Unmarshal(headerData, &header); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JWT header: %v", err)
	}

	// Get Apple's public keys
	jwks, err := a.getAppleJWKS()
	if err != nil {
		return nil, fmt.Errorf("failed to get Apple JWKS: %v", err)
	}

	// Find the matching key
	var matchingKey *model.AppleJWK
	for _, key := range jwks.Keys {
		if key.Kid == header.Kid {
			matchingKey = &key
			break
		}
	}

	if matchingKey == nil {
		return nil, fmt.Errorf("no matching key found for kid: %s", header.Kid)
	}

	// Create RSA public key from JWK
	publicKey, err := a.createRSAPublicKey(matchingKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create RSA public key: %v", err)
	}

	// Parse and verify the token
	token, err := jwt.ParseWithClaims(idToken, &model.AppleClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to verify Apple ID token: %v", err)
	}

	claims, ok := token.Claims.(*model.AppleClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid Apple ID token claims")
	}

	// Verify issuer
	if claims.Iss != "https://appleid.apple.com" {
		return nil, fmt.Errorf("invalid issuer: %s", claims.Iss)
	}

	// Create AppleUserInfo from claims
	var name string
	if claims.Name.FirstName != "" || claims.Name.LastName != "" {
		name = strings.TrimSpace(claims.Name.FirstName + " " + claims.Name.LastName)
	}

	userInfo := &model.AppleUserInfo{
		Sub:           claims.Sub,
		Email:         claims.Email,
		VerifiedEmail: claims.EmailVerified,
		Name:          name,
		GivenName:     claims.Name.FirstName,
		FamilyName:    claims.Name.LastName,
	}

	return userInfo, nil
}

func (a *AuthService) getAppleJWKS() (*model.AppleJWKS, error) {
	resp, err := http.Get(a.getAppleJWKSURL())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var jwks model.AppleJWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, err
	}

	return &jwks, nil
}

func (a *AuthService) createRSAPublicKey(jwk *model.AppleJWK) (interface{}, error) {
	// Decode base64url-encoded n and e
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode n: %v", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode e: %v", err)
	}

	// Convert to big.Int
	n := new(big.Int).SetBytes(nBytes)
	e := int(new(big.Int).SetBytes(eBytes).Int64())

	return &rsa.PublicKey{
		N: n,
		E: e,
	}, nil
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
		AppleID:  *new(string),
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

	refreshTokenBlackListService := a.getBlacklistSvc()
	blackList, _ := refreshTokenBlackListService.GetByToken(refreshToken)

	if blackList != nil {
		return "", fmt.Errorf("refresh token is blacklisted")
	}

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
		log.Println("Failed to parse refresh token:", err)
		return "", err
	} else if claims, ok := token.Claims.(*model.JwtCustomClaims); ok {
		user, err := a.getUserSvc().FindByID(claims.ID)

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

		if user.AppleID == nil {
			user.AppleID = new(string)
		}

		newClaim := model.JwtCustomClaims{
			Email:    claims.Email,
			GoogleID: *user.GoogleID,
			AppleID:  *user.AppleID,
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
