package router

import (
	"what-to-eat/be/constants"
	controllers "what-to-eat/be/controller"
	"what-to-eat/be/middleware"

	"github.com/labstack/echo/v4"
)

func UseUserDishInteractionRouter(group *echo.Group) {
	aG := middleware.NewAuthGuard()
	rG := middleware.NewRoleGuard()
	controller := controllers.NewUserDishInteractionController()

	group.POST("/view/", controller.RecordView, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_USER_DISH_INTERACTIONS}))
	group.POST("/cooked/", controller.RecordCooked, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_USER_DISH_INTERACTIONS}))
	group.POST("/rate/", controller.RateDish, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_USER_DISH_INTERACTIONS}))
	group.POST("/share/", controller.RecordShare, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_USER_DISH_INTERACTIONS}))
	group.GET("/", controller.GetUserInteractions, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_USER_DISH_INTERACTIONS}))
	group.GET("/top/", controller.GetTopInteractedDishes, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_USER_DISH_INTERACTIONS}))
}
