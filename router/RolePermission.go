package router

import (
	"what-to-eat/be/constants"
	"what-to-eat/be/controller"
	"what-to-eat/be/middleware"

	"github.com/labstack/echo/v4"
)

func UseRolePermissionRouter(group *echo.Group) {
	aG := middleware.NewAuthGuard()
	rG := middleware.NewRoleGuard()
	group.Use(aG.AuthGuard)
	controller := &controller.RolePermission{}

	// CRUD routes
	group.POST("/", controller.Create, rG.RoleGuard([]string{constants.CREATE_ROLE}))
	group.PATCH("/:id/", controller.Update, rG.RoleGuard([]string{constants.UPDATE_ROLE}))
	group.DELETE("/:id/", controller.Remove, rG.RoleGuard([]string{constants.REMOVE_ROLE}))
	group.GET("/", controller.FindAll, rG.RoleGuard([]string{constants.FIND_ALL_ROLE}))
	group.GET("/:id/", controller.FindOne, rG.RoleGuard([]string{constants.FIND_ONE_ROLE}))
	group.GET("/by-name/:roleName/", controller.FindByName, rG.RoleGuard([]string{constants.FIND_ONE_ROLE}))
}
