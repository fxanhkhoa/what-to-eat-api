package controller

import (
	"net/http"
	"what-to-eat/be/model"
	"what-to-eat/be/service"

	"github.com/labstack/echo/v4"
)

type AuthController struct {
}

func (cr *AuthController) LoginWithGoogle(c echo.Context) error {
	var dto model.LoginDto
	if err := c.Bind(&dto); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	var service = &service.AuthService{}
	result, err := service.Login(dto.Token)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}
	return c.JSON(http.StatusOK, result)
}
