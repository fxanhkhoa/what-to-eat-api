package router

import (
	"what-to-eat/be/constants"
	controllers "what-to-eat/be/controller"
	"what-to-eat/be/middleware"

	"github.com/labstack/echo/v4"
)

func UseUserSavedDishRouter(group *echo.Group) {
	aG := middleware.NewAuthGuard()
	rG := middleware.NewRoleGuard()
	controller := controllers.NewUserSavedDishController()

	group.POST("/", controller.SaveDish, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_USER_SAVED_DISHES}))
	group.GET("/", controller.GetUserSavedDishes, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_USER_SAVED_DISHES}))
	group.PUT("/:id/", controller.UpdateSavedDish, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_USER_SAVED_DISHES}))
	group.DELETE("/:dishSlug/", controller.RemoveSavedDish, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_USER_SAVED_DISHES}))
	group.GET("/:dishSlug/check/", controller.CheckIsSaved, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_USER_SAVED_DISHES}))
}
