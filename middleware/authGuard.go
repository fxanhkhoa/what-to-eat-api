package middleware

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
	"what-to-eat/be/config"
	"what-to-eat/be/model"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type (
	AuthGuard struct {
		Uptime       time.Time      `json:"uptime"`
		RequestCount uint64         `json:"requestCount"`
		Statuses     map[string]int `json:"statuses"`
		mutex        sync.RWMutex
	}
)

func NewAuthGuard() *AuthGuard {
	return &AuthGuard{
		Uptime:   time.Now(),
		Statuses: map[string]int{},
	}
}

// Process is the middleware function.
func (aG *AuthGuard) AuthGuard(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		aG.mutex.Lock()
		defer aG.mutex.Unlock()
		aG.RequestCount++
		status := strconv.Itoa(c.Response().Status)
		aG.Statuses[status]++

		authorization := c.Request().Header.Clone().Get("Authorization")
		if authorization == "" {
			return echo.ErrUnauthorized
		}
		bearerToken := strings.Split(authorization, "Bearer ")[1]

		token, err := jwt.ParseWithClaims(bearerToken, &model.JwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.GetInstanceConfig().JWTSecret), nil
		})
		if err != nil {
			fmt.Print(err.Error())
			return err
		} else if claims, ok := token.Claims.(*model.JwtCustomClaims); ok {
			c.Set("CLAIM", claims)
			if err := next(c); err != nil {
				c.Error(err)
			}

			return nil
		} else {
			return errors.New("unknown claims type, cannot proceed")
		}
	}
}

// Example with input param
func (aG *AuthGuard) AuthGuardWithParam(isPublic bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if err := next(c); err != nil {
				c.Error(err)
			}
			aG.mutex.Lock()
			defer aG.mutex.Unlock()
			aG.RequestCount++
			status := strconv.Itoa(c.Response().Status)
			aG.Statuses[status]++
			return nil
		}
	}
}
