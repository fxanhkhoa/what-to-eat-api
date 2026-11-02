package router

import (
	"what-to-eat/be/constants"
	controllers "what-to-eat/be/controller"
	"what-to-eat/be/middleware"

	"github.com/labstack/echo/v4"
)

func UseUserFavoriteRouter(group *echo.Group) {
	aG := middleware.NewAuthGuard()
	rG := middleware.NewRoleGuard()
	controller := controllers.NewUserFavoriteController()

	group.POST("/", controller.AddFavorite, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_USER_FAVORITES}))
	group.GET("/", controller.GetUserFavorites, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_USER_FAVORITES}))
	group.DELETE("/:dishSlug/", controller.RemoveFavorite, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_USER_FAVORITES}))
	group.GET("/:dishSlug/check/", controller.CheckIsFavorite, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_USER_FAVORITES}))
}
