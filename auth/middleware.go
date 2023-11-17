package auth

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"what-to-eat/be/graph/model"
	"what-to-eat/be/graph/service"

	"github.com/golang-jwt/jwt/v5"
)

var userCtxKey = &contextKey{"user"}

type contextKey struct {
	name string
}

type CustomClaims struct {
	ID string `json:"id"`
	jwt.RegisteredClaims
}

func Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			secretKey := os.Getenv("SECRET_KEY")
			header := r.Header.Get("Authorization")

			// Allow unauthenticated users in
			if header == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// validate jwt token
			tokenStr := strings.Split(header, " ")
			token, err := jwt.ParseWithClaims(tokenStr[1], &service.CustomClaim{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(secretKey), nil
			})

			if err != nil {
				log.Println(err.Error())
				http.Error(w, err.Error(), http.StatusUnauthorized)
			} else if claims, ok := token.Claims.(*service.CustomClaim); ok {
				user, err := service.NewUserService().FindByID(claims.ID)

				if err != nil {
					http.Error(w, err.Error(), http.StatusUnauthorized)
					return
				}

				ctx := context.WithValue(r.Context(), userCtxKey, user)
				r = r.WithContext(ctx)
			} else {
				log.Println(err.Error())
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ForContext finds the user from the context. REQUIRES Middleware to have run.
func ForContext(ctx context.Context) *model.User {
	raw, _ := ctx.Value(userCtxKey).(*model.User)
	return raw
}
