package router

import (
	"time"
	"what-to-eat/be/controller"
	"what-to-eat/be/middleware"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

func UseFeedbackRouter(group *echo.Group) {
	aG := middleware.NewAuthGuard()
	controller := &controller.FeedbackController{}

	// Rate limiter for public feedback creation endpoint
	feedbackPostLimiter := echoMiddleware.RateLimiterWithConfig(echoMiddleware.RateLimiterConfig{
		Store: echoMiddleware.NewRateLimiterMemoryStoreWithConfig(
			echoMiddleware.RateLimiterMemoryStoreConfig{
				Rate:      3,               // 3 requests
				Burst:     5,               // Burst of 5 requests
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
				"error": "Too many feedback submissions. Please try again later.",
			})
		},
	})

	// Public route - anyone can create feedback (with rate limiting)
	group.POST("/", controller.Create, feedbackPostLimiter)

	// Protected routes - require authentication
	group.GET("/", controller.GetAll, aG.AuthGuard)
	group.GET("/:id/", controller.GetById, aG.AuthGuard)
	group.PATCH("/:id/", controller.Update, aG.AuthGuard)
	group.DELETE("/:id/", controller.Delete, aG.AuthGuard)
}
