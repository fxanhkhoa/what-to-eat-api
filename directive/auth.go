package directive

import (
	"context"
	"errors"
	"fmt"
	"what-to-eat/be/auth"
	"what-to-eat/be/graph/model"
	"what-to-eat/be/graph/service"

	"github.com/99designs/gqlgen/graphql"
)

type contextKey struct {
	name string
}

var IsPublicKey = &contextKey{"isPublic"}
var userCtxKey = &contextKey{"user"}

func Auth(ctx context.Context, obj interface{}, next graphql.Resolver, index *int) (interface{}, error) {
	fmt.Println("go in")
	fmt.Println(next)

	// secretKey := os.Getenv("SECRET_KEY")
	// token, err := jwt.ParseWithClaims(refreshToken, &CustomClaim{}, func(token *jwt.Token) (interface{}, error) {
	// 	return []byte(secretKey), nil
	// })

	ctx = context.WithValue(ctx, IsPublicKey, true)
	return next(ctx)
}

func Role(ctx context.Context, obj interface{}, next graphql.Resolver, role model.Role) (interface{}, error) {
	user := auth.ForContext(ctx)
	roleName := user.RoleName
	rolePermission, err := service.NewRolePermissionService().FindByName(roleName)
	if err != nil {
		return nil, err
	}

	if !IsAllowedPermission(rolePermission.Permission, role) {
		return nil, errors.New("NO PERMISSION")
	}

	return next(ctx)
}

func IsAllowedPermission(userPermissions []*string, allowedRole model.Role) bool {
	allowedRoleStr := allowedRole.String()
	for _, value := range userPermissions {
		if *value == allowedRoleStr {
			return true
		}
	}
	return false
}
