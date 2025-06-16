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
	group.GET("/by-name/:roleName/", controller.FindByName, rG.RoleGuard([]string{constants.FIND_ONE_ROLE}))
}
