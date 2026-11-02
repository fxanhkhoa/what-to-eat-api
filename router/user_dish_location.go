package router

import (
	"what-to-eat/be/constants"
	controllers "what-to-eat/be/controller"
	"what-to-eat/be/middleware"

	"github.com/labstack/echo/v4"
)

func UseUserDishLocationRouter(group *echo.Group) {
	aG := middleware.NewAuthGuard()
	rG := middleware.NewRoleGuard()
	controller := controllers.NewUserDishLocationController()

	// All routes are private — personal locations belong to the authenticated user
	group.GET("/", controller.Find, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_USER_DISH_LOCATIONS}))
	group.GET("/:id/", controller.FindOne, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_USER_DISH_LOCATIONS}))
	group.POST("/", controller.Create, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_USER_DISH_LOCATIONS}))
	group.PATCH("/:id/", controller.Update, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_USER_DISH_LOCATIONS}))
	group.DELETE("/:id/", controller.Remove, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_USER_DISH_LOCATIONS}))
}
