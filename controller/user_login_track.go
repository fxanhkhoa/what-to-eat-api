package controller

import (
	"net/http"
	"strconv"
	"what-to-eat/be/helper"
	"what-to-eat/be/service"

	"github.com/labstack/echo/v4"
)

type UserLoginTrackController struct{}

func (ctrl *UserLoginTrackController) GetAllUserLogins(c echo.Context) error {
	limitStr := c.QueryParam("limit")
	limit := int64(50)
	if limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 64); err == nil {
			limit = l
		}
	}
	tracks, count, err := (&service.UserLoginTrackService{}).GetAllUserLogins(limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, helper.PaginationObject{
		Data:  tracks,
		Count: count,
	})
}

func (ctrl *UserLoginTrackController) GetUserLogins(c echo.Context) error {
	userId := c.Param("userId")
	limitStr := c.QueryParam("limit")
	limit := int64(10)
	if limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 64); err == nil {
			limit = l
		}
	}
	tracks, count, err := (&service.UserLoginTrackService{}).GetUserLogins(userId, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, helper.PaginationObject{
		Data:  tracks,
		Count: count,
	})
}
