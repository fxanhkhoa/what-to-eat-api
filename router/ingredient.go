package router

import (
	controllers "what-to-eat/be/controller"

	"github.com/labstack/echo/v4"
)

func UseIngredientRouter(group *echo.Group) {
	// aG := middleware.NewAuthGuard()
	// rG := middleware.NewRoleGuard()
	controller := &controllers.IngredientController{}
	group.GET("/", controller.Find)
	group.GET("/byTitleLang", controller.Find)
	group.GET("/:id", controller.FindOne)
	group.GET("/slug/:slug", controller.FindOneBySlug)
	// group.POST("/", controller.Create, aG.AuthGuard, rG.RoleGuard([]string{constants.CREATE_DISH}))
}
