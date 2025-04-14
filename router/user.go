package router

import (
	"what-to-eat/be/constants"
	"what-to-eat/be/controller"
	"what-to-eat/be/middleware"

	"github.com/labstack/echo/v4"
)

func UseUserGroup(group *echo.Group) {
	aG := middleware.NewAuthGuard()
	rG := middleware.NewRoleGuard()
	controller := &controller.UserController{}
	group.Use(aG.AuthGuard)
	group.POST("/", controller.Create, rG.RoleGuard([]string{constants.CREATE_USER}))
	group.GET("/:id/", controller.FindOne, rG.RoleGuard([]string{constants.FIND_ONE_USER}))
	group.GET("/", controller.FindAll, rG.RoleGuard([]string{constants.FIND_ALL_USER}))
	group.PATCH("/:id/", controller.Update, rG.RoleGuard([]string{constants.UPDATE_USER}))
	group.DELETE("/:id/", controller.Delete, rG.RoleGuard([]string{constants.REMOVE_USER}))
}
