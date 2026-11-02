package router

import (
	"what-to-eat/be/constants"
	controllers "what-to-eat/be/controller"
	"what-to-eat/be/middleware"

	"github.com/labstack/echo/v4"
)

func UseFoodShopRouter(group *echo.Group) {
	aG := middleware.NewAuthGuard()
	rG := middleware.NewRoleGuard()
	controller := controllers.NewFoodShopController()

	// Public: browse & search food shops (with proximity and open/closed filter)
	group.GET("/", controller.Find)
	group.GET("/:id/", controller.FindOne)

	// Admin: create, update, delete
	group.POST("/", controller.Create, aG.AuthGuard, rG.RoleGuard([]string{constants.CREATE_FOOD_SHOP}))
	group.PATCH("/:id/", controller.Update, aG.AuthGuard, rG.RoleGuard([]string{constants.UPDATE_FOOD_SHOP}))
	group.DELETE("/:id/", controller.Remove, aG.AuthGuard, rG.RoleGuard([]string{constants.REMOVE_FOOD_SHOP}))

	// Verified user: manually mark shop open/closed (replaces Google Maps sync)
	group.PATCH("/:id/verify/", controller.VerifyStatus, aG.AuthGuard, rG.RoleGuard([]string{constants.VERIFY_FOOD_SHOP}))
}
