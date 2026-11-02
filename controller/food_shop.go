package controller

import (
	"net/http"
	"strconv"
	"what-to-eat/be/model"
	"what-to-eat/be/service"

	"github.com/labstack/echo/v4"
)

type FoodShopController struct{}

func NewFoodShopController() *FoodShopController {
	return &FoodShopController{}
}

func (c *FoodShopController) Find(ctx echo.Context) error {
	var query model.QueryFoodShopDto
	var err error

	query.BaseDto.Page, err = strconv.Atoi(ctx.QueryParam("page"))
	if err != nil || query.BaseDto.Page < 1 {
		query.BaseDto.Page = 1
	}
	query.BaseDto.Limit, err = strconv.Atoi(ctx.QueryParam("limit"))
	if err != nil || query.BaseDto.Limit < 1 {
		query.BaseDto.Limit = 10
	}

	if keyword := ctx.QueryParam("keyword"); keyword != "" {
		query.Keyword = &keyword
	}

	if isOpenStr := ctx.QueryParam("isOpen"); isOpenStr != "" {
		v := isOpenStr == "true"
		query.IsOpen = &v
	}

	if latStr := ctx.QueryParam("latitude"); latStr != "" {
		v, err := strconv.ParseFloat(latStr, 64)
		if err != nil {
			return ctx.String(http.StatusBadRequest, "invalid latitude")
		}
		query.Latitude = &v
	}
	if lonStr := ctx.QueryParam("longitude"); lonStr != "" {
		v, err := strconv.ParseFloat(lonStr, 64)
		if err != nil {
			return ctx.String(http.StatusBadRequest, "invalid longitude")
		}
		query.Longitude = &v
	}
	if radiusStr := ctx.QueryParam("radius"); radiusStr != "" {
		v, err := strconv.ParseFloat(radiusStr, 64)
		if err != nil || v <= 0 {
			return ctx.String(http.StatusBadRequest, "invalid radius")
		}
		query.Radius = &v
	}

	svc := NewFoodShopService()
	shops, count, err := svc.Find(query)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, map[string]any{
		"data":  shops,
		"count": count,
	})
}

func (c *FoodShopController) FindOne(ctx echo.Context) error {
	id := ctx.Param("id")
	svc := NewFoodShopService()
	shop, err := svc.FindOne(id)
	if err != nil {
		return ctx.String(http.StatusNotFound, err.Error())
	}

	computed := service.ComputeIsOpen(shop)
	return ctx.JSON(http.StatusOK, map[string]any{
		"data":           shop,
		"computedIsOpen": computed,
	})
}

func (c *FoodShopController) Create(ctx echo.Context) error {
	var dto model.CreateFoodShopDto
	if err := ctx.Bind(&dto); err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}

	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)
	svc := NewFoodShopService()
	shop, err := svc.Create(dto, claim)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusCreated, shop)
}

func (c *FoodShopController) Update(ctx echo.Context) error {
	id := ctx.Param("id")
	var dto model.UpdateFoodShopDto
	if err := ctx.Bind(&dto); err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}
	dto.ID = id

	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)
	svc := NewFoodShopService()
	shop, err := svc.Update(dto, claim)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusOK, shop)
}

// VerifyStatus lets an authorized user manually mark a shop as open/closed.
func (c *FoodShopController) VerifyStatus(ctx echo.Context) error {
	id := ctx.Param("id")
	var body struct {
		IsOpen bool `json:"isOpen"`
	}
	if err := ctx.Bind(&body); err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}

	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)
	svc := NewFoodShopService()
	shop, err := svc.VerifyStatus(id, body.IsOpen, claim)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusOK, shop)
}

func (c *FoodShopController) Remove(ctx echo.Context) error {
	id := ctx.Param("id")
	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)

	svc := NewFoodShopService()
	shop, err := svc.Remove(id, claim)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusOK, shop)
}

// NewFoodShopService is a thin wrapper so the controller file stays self-contained.
func NewFoodShopService() *service.FoodShopService {
	return service.NewFoodShopService()
}
