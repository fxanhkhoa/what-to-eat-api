package router

import (
	controllers "what-to-eat/be/controller"

	"github.com/labstack/echo/v4"
)

func UseDishRouter(group *echo.Group) {
	controller := &controllers.DishController{}

	group.GET("/", controller.Find)
}
