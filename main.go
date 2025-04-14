package main

import (
	"net/http"
	"what-to-eat/be/config"
	"what-to-eat/be/firebase"
	"what-to-eat/be/router"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	config.GetDBInstance()
	firebase.InitFirebase()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
	}))

	e.Pre(middleware.AddTrailingSlash())
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20)))

	e.Use(middleware.Logger())
	e.Use(echoprometheus.NewMiddleware("myapp"))   // adds middleware to gather metrics
	e.GET("/metrics", echoprometheus.NewHandler()) // adds route to serve gathered metrics

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	userGroup := e.Group("/user")
	router.UseUserGroup(userGroup)

	e.Logger.Fatal(e.Start(":1323"))
}
