package router

import (
	"what-to-eat/be/constants"
	controllers "what-to-eat/be/controller"
	"what-to-eat/be/middleware"

	"github.com/labstack/echo/v4"
)

func UseFoodShopDishRouter(group *echo.Group) {
	aG := middleware.NewAuthGuard()
	rG := middleware.NewRoleGuard()
	controller := controllers.NewFoodShopDishController()

	// Public: browse dishes linked to a food shop (or shops linked to a dish)
	group.GET("/", controller.Find)
	group.GET("/:id/", controller.FindOne)

	// Admin: manage food shop ↔ dish relationships
	group.POST("/", controller.Create, aG.AuthGuard, rG.RoleGuard([]string{constants.CREATE_FOOD_SHOP_DISH}))
	group.PATCH("/:id/", controller.Update, aG.AuthGuard, rG.RoleGuard([]string{constants.UPDATE_FOOD_SHOP_DISH}))
	group.DELETE("/:id/", controller.Remove, aG.AuthGuard, rG.RoleGuard([]string{constants.REMOVE_FOOD_SHOP_DISH}))
}
