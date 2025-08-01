package router

import (
	"what-to-eat/be/constants"
	controllers "what-to-eat/be/controller"
	"what-to-eat/be/middleware"

	"github.com/labstack/echo/v4"
)

func UseDishVoteRouter(group *echo.Group) {
	aG := middleware.NewAuthGuard()
	rG := middleware.NewRoleGuard()
	controller := &controllers.DishVoteController{}
	group.GET("/", controller.Find, aG.AuthGuard, rG.RoleGuard([]string{constants.FIND_ALL_DISH_VOTE}))
	group.GET("/:id/", controller.FindOne, aG.AuthGuard, rG.RoleGuard([]string{constants.FIND_ONE_DISH_VOTE}))
	group.POST("/", controller.Create, aG.AuthGuard, rG.RoleGuard([]string{constants.CREATE_DISH_VOTE}))
}
