package router

import (
	"what-to-eat/be/constants"
	"what-to-eat/be/controller"
	"what-to-eat/be/middleware"

	"github.com/labstack/echo/v4"
)

func UseUserLoginTrackGroup(group *echo.Group) {
	aG := middleware.NewAuthGuard()
	rG := middleware.NewRoleGuard()
	controller := &controller.UserLoginTrackController{}
	group.GET("/", controller.GetAllUserLogins, aG.AuthGuard, rG.RoleGuard([]string{constants.FIND_ALL_USER_LOGIN_TRACK}))
	group.GET("/:userId/", controller.GetUserLogins, aG.AuthGuard, rG.RoleGuard([]string{constants.FIND_USER_LOGIN_TRACK}))
}
