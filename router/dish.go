package router

import (
	"what-to-eat/be/constants"
	controllers "what-to-eat/be/controller"
	"what-to-eat/be/middleware"

	"github.com/labstack/echo/v4"
)

func UseDishRouter(group *echo.Group) {
	aG := middleware.NewAuthGuard()
	rG := middleware.NewRoleGuard()
	controller := &controllers.DishController{}
	group.GET("/", controller.Find)
	group.GET("/random/", controller.FindRandom)
	group.GET("/:id/", controller.FindOne)
	group.GET("/slug/:slug/", controller.FindOneBySlug)
	group.POST("/", controller.Create, aG.AuthGuard, rG.RoleGuard([]string{constants.CREATE_DISH}))
	group.PATCH("/:id/", controller.Update, aG.AuthGuard, rG.RoleGuard([]string{constants.UPDATE_DISH}))
	group.DELETE("/:id/", controller.Remove, aG.AuthGuard, rG.RoleGuard([]string{constants.REMOVE_DISH}))
}
