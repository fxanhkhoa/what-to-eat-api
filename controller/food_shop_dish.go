package controller

import (
	"net/http"
	"strconv"
	"what-to-eat/be/model"
	"what-to-eat/be/service"

	"github.com/labstack/echo/v4"
)

type FoodShopDishController struct{}

func NewFoodShopDishController() *FoodShopDishController {
	return &FoodShopDishController{}
}

func (c *FoodShopDishController) Find(ctx echo.Context) error {
	var query model.QueryFoodShopDishDto
	var err error

	query.BaseDto.Page, err = strconv.Atoi(ctx.QueryParam("page"))
	if err != nil || query.BaseDto.Page < 1 {
		query.BaseDto.Page = 1
	}
	query.BaseDto.Limit, err = strconv.Atoi(ctx.QueryParam("limit"))
	if err != nil || query.BaseDto.Limit < 1 {
		query.BaseDto.Limit = 10
	}

	if v := ctx.QueryParam("foodShopId"); v != "" {
		query.FoodShopID = &v
	}
	if v := ctx.QueryParam("dishSlug"); v != "" {
		query.DishSlug = &v
	}

	svc := service.NewFoodShopDishService()
	entries, count, err := svc.Find(query)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, map[string]any{
		"data":  entries,
		"count": count,
	})
}

func (c *FoodShopDishController) FindOne(ctx echo.Context) error {
	id := ctx.Param("id")
	svc := service.NewFoodShopDishService()
	entry, err := svc.FindOne(id)
	if err != nil {
		return ctx.String(http.StatusNotFound, err.Error())
	}
	return ctx.JSON(http.StatusOK, entry)
}

func (c *FoodShopDishController) Create(ctx echo.Context) error {
	var dto model.CreateFoodShopDishDto
	if err := ctx.Bind(&dto); err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}

	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)
	svc := service.NewFoodShopDishService()
	entry, err := svc.Create(dto, claim)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusCreated, entry)
}

func (c *FoodShopDishController) Update(ctx echo.Context) error {
	id := ctx.Param("id")
	var dto model.UpdateFoodShopDishDto
	if err := ctx.Bind(&dto); err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	dto.ID = id

	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)
	svc := service.NewFoodShopDishService()
	entry, err := svc.Update(dto, claim)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusOK, entry)
}

func (c *FoodShopDishController) Remove(ctx echo.Context) error {
	id := ctx.Param("id")
	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)

	svc := service.NewFoodShopDishService()
	entry, err := svc.Remove(id, claim)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusOK, entry)
}
