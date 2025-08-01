package router

import (
	"time"
	"what-to-eat/be/constants"
	"what-to-eat/be/controller"
	"what-to-eat/be/middleware"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

func UseContactRouter(group *echo.Group) {
	aG := middleware.NewAuthGuard()
	rG := middleware.NewRoleGuard()
	controller := &controller.ContactController{}

	contactPostLimiter := echoMiddleware.RateLimiterWithConfig(echoMiddleware.RateLimiterConfig{
		Store: echoMiddleware.NewRateLimiterMemoryStoreWithConfig(
			echoMiddleware.RateLimiterMemoryStoreConfig{
				Rate:      5,               // 5 requests
				Burst:     10,              // Burst of 10 requests
				ExpiresIn: 1 * time.Minute, // Per minute
			},
		),
		// Use IP address as identifier
		IdentifierExtractor: func(c echo.Context) (string, error) {
			return c.RealIP(), nil
		},
		// Custom error response
		DenyHandler: func(c echo.Context, identifier string, err error) error {
			return c.JSON(429, map[string]string{
				"error": "Too many contact form submissions. Please try again later.",
			})
		},
	})

	group.GET("/", controller.Find, aG.AuthGuard, rG.RoleGuard([]string{constants.FIND_ALL_CONTACT}))
	group.GET("/:id/", controller.FindOne, aG.AuthGuard, rG.RoleGuard([]string{constants.FIND_ONE_CONTACT}))
	group.POST("/", controller.Create, contactPostLimiter)
	group.PATCH("/:id/", controller.Update, aG.AuthGuard, rG.RoleGuard([]string{constants.UPDATE_CONTACT}))
	group.DELETE("/:id/", controller.Remove, aG.AuthGuard, rG.RoleGuard([]string{constants.REMOVE_CONTACT}))
}
