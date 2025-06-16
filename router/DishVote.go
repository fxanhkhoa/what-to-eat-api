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
	group.Use(aG.AuthGuard)
	controller := &controllers.DishVoteController{}
	group.GET("/", controller.Find, rG.RoleGuard([]string{constants.FIND_ALL_INGREDIENT}))
	group.GET("/:id", controller.FindOne, rG.RoleGuard([]string{constants.FIND_ONE_INGREDIENT}))
	// group.POST("/", controller.Create, aG.AuthGuard, rG.RoleGuard([]string{constants.CREATE_DISH}))
}
