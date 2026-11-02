package controller

import (
	"net/http"
	"strconv"
	"what-to-eat/be/model"
	"what-to-eat/be/service"

	"github.com/labstack/echo/v4"
)

type UserDishLocationController struct{}

func NewUserDishLocationController() *UserDishLocationController {
	return &UserDishLocationController{}
}

func (c *UserDishLocationController) Find(ctx echo.Context) error {
	var query model.QueryUserDishLocationDto
	var err error

	query.BaseDto.Page, err = strconv.Atoi(ctx.QueryParam("page"))
	if err != nil || query.BaseDto.Page < 1 {
		query.BaseDto.Page = 1
	}
	query.BaseDto.Limit, err = strconv.Atoi(ctx.QueryParam("limit"))
	if err != nil || query.BaseDto.Limit < 1 {
		query.BaseDto.Limit = 10
	}

	if v := ctx.QueryParam("dishSlug"); v != "" {
		query.DishSlug = &v
	}

	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)
	svc := service.NewUserDishLocationService()
	locations, count, err := svc.Find(query, claim.ID)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, map[string]any{
		"data":  locations,
		"count": count,
	})
}

func (c *UserDishLocationController) FindOne(ctx echo.Context) error {
	id := ctx.Param("id")
	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)

	svc := service.NewUserDishLocationService()
	location, err := svc.FindOne(id, claim.ID)
	if err != nil {
		return ctx.String(http.StatusNotFound, err.Error())
	}
	return ctx.JSON(http.StatusOK, location)
}

func (c *UserDishLocationController) Create(ctx echo.Context) error {
	var dto model.CreateUserDishLocationDto
	if err := ctx.Bind(&dto); err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}

	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)
	dto.UserID = claim.ID

	svc := service.NewUserDishLocationService()
	location, err := svc.Create(dto, claim)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusCreated, location)
}

func (c *UserDishLocationController) Update(ctx echo.Context) error {
	id := ctx.Param("id")
	var dto model.UpdateUserDishLocationDto
	if err := ctx.Bind(&dto); err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	dto.ID = id

	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)
	svc := service.NewUserDishLocationService()
	location, err := svc.Update(dto, claim)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusOK, location)
}

func (c *UserDishLocationController) Remove(ctx echo.Context) error {
	id := ctx.Param("id")
	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)

	svc := service.NewUserDishLocationService()
	location, err := svc.Remove(id, claim)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusOK, location)
}
