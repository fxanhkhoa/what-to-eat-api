package router

import "github.com/labstack/echo/v4"

func InitializeRoutes(e *echo.Echo) {
	authGroup := e.Group("/auth")
	UseAuthGroup(authGroup)

	userGroup := e.Group("/user")
	UseUserGroup(userGroup)

	dishGroup := e.Group("/dish")
	UseDishRouter(dishGroup)

	ingredientGroup := e.Group("/ingredient")
	UseIngredientRouter(ingredientGroup)

	dishVoteGroup := e.Group("/dish-vote")
	UseDishVoteRouter(dishVoteGroup)

	rolePermissionGroup := e.Group("/authorization")
	UseRolePermissionRouter(rolePermissionGroup)

	contactGroup := e.Group("/contact")
	UseContactRouter(contactGroup)
}
