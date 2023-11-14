package auth

import (
	"context"
	"net/http"
	"what-to-eat/be/graph/model"

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
			// secret := os.Getenv("SECRET_KEY")
			// header := r.Header.Get("Authorization")

			// // Allow unauthenticated users in
			// if header == "" {
			// 	next.ServeHTTP(w, r)
			// 	return
			// }

			// //validate jwt token
			// tokenStr := header
			// token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			// 	return []byte(secret), nil
			// })

			// if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
			// 	fmt.Printf("%v %v", claims.ID, claims.RegisteredClaims.Issuer)
			// } else {
			// 	fmt.Println(err)
			// }

			// if err != nil {
			// 	next.ServeHTTP(w, r)
			// 	return
			// }
			// user := model.User{}
			// // put it in context
			// ctx := context.WithValue(r.Context(), userCtxKey, &user)

			// and call the next with our new context
			// r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// ForContext finds the user from the context. REQUIRES Middleware to have run.
func ForContext(ctx context.Context) *model.User {
	raw, _ := ctx.Value(userCtxKey).(*model.User)
	return raw
}
