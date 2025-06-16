package router

import (
	"what-to-eat/be/controller"

	"github.com/labstack/echo/v4"
)

func UseAuthGroup(group *echo.Group) {
	// aG := middleware.NewAuthGuard()
	// rG := middleware.NewRoleGuard()
	controller := &controller.AuthController{}
	group.POST("/login/", controller.LoginWithGoogle)
}
