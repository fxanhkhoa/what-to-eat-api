package directive

import (
	"context"
	"fmt"
	"what-to-eat/be/graph/model"

	"github.com/99designs/gqlgen/graphql"
)

type contextKey struct {
	name string
}

var IsPublicKey = &contextKey{"isPublic"}

func Auth(ctx context.Context, obj interface{}, next graphql.Resolver, index *int) (interface{}, error) {
	fmt.Println("go in")
	ctx = context.WithValue(ctx, IsPublicKey, true)
	return next(ctx)
}

func Role(ctx context.Context, obj interface{}, next graphql.Resolver, role model.Role) (interface{}, error) {
	fmt.Println("go in Role")
	fmt.Println(ctx.Value(IsPublicKey))
	return next(ctx)
}
