package controller

import (
	"net/http"
	"what-to-eat/be/model"
	"what-to-eat/be/service"

	"github.com/labstack/echo/v4"
)

type AuthController struct {
}

func (cr *AuthController) Login(c echo.Context) error {
	var dto model.LoginDto
	if err := c.Bind(&dto); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	var service = &service.AuthService{}
	result, err := service.Login(dto)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}
	return c.JSON(http.StatusOK, result)
}

func (cr *AuthController) RefreshToken(c echo.Context) error {
	var dto model.RefreshTokenDto
	if err := c.Bind(&dto); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	var service = &service.AuthService{}
	var data model.TokenResult
	result, err := service.GenerateToken(dto.RefreshToken)
	data.RefreshToken = dto.RefreshToken
	data.Token = result
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, data)
}

func (cr *AuthController) Logout(c echo.Context) error {
	var dto model.LogoutDto
	if err := c.Bind(&dto); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	var service = &service.AuthService{}
	claim := c.Get("CLAIM").(*model.JwtCustomClaims)
	if err := service.Logout(dto.RefreshToken, claim); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusOK)
}

func (cr *AuthController) GetProfile(c echo.Context) error {
	claim := c.Get("CLAIM").(*model.JwtCustomClaims)
	var service = &service.UserService{}
	result, err := service.FindByID(claim.ID)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, result)
}
