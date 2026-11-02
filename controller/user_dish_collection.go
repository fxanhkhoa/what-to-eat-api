package controller

import (
	"net/http"
	"strconv"
	"what-to-eat/be/model"
	"what-to-eat/be/service"

	"github.com/labstack/echo/v4"
)

type UserDishCollectionController struct{}

func NewUserDishCollectionController() *UserDishCollectionController {
	return &UserDishCollectionController{}
}

func (c *UserDishCollectionController) Create(ctx echo.Context) error {
	var dto model.CreateUserDishCollectionDto
	if err := ctx.Bind(&dto); err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}

	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)
	dto.UserID = claim.ID

	svc := service.NewUserDishCollectionService()
	collection, err := svc.Create(dto, claim)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, collection)
}

func (c *UserDishCollectionController) Update(ctx echo.Context) error {
	id := ctx.Param("id")
	var dto model.UpdateUserDishCollectionDto
	if err := ctx.Bind(&dto); err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}

	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)
	dto.ID = id
	dto.UserID = claim.ID

	svc := service.NewUserDishCollectionService()
	collection, err := svc.Update(dto, claim)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, collection)
}

func (c *UserDishCollectionController) Delete(ctx echo.Context) error {
	id := ctx.Param("id")
	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)

	svc := service.NewUserDishCollectionService()
	if err := svc.Delete(id, claim.ID, claim); err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, map[string]string{"message": "Collection deleted successfully"})
}

func (c *UserDishCollectionController) GetCollections(ctx echo.Context) error {
	var query model.QueryUserDishCollectionDto

	var err error
	query.BaseDto.Page, err = strconv.Atoi(ctx.QueryParam("page"))
	if err != nil || query.BaseDto.Page < 0 {
		query.BaseDto.Page = 1
	}

	query.BaseDto.Limit, err = strconv.Atoi(ctx.QueryParam("limit"))
	if err != nil || query.BaseDto.Limit < 0 {
		query.BaseDto.Limit = 10
	}

	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)
	userID := claim.ID
	query.UserID = &userID

	occasion := ctx.QueryParam("occasion")
	if occasion != "" {
		query.Occasion = &occasion
	}

	keyword := ctx.QueryParam("keyword")
	if keyword != "" {
		query.Keyword = &keyword
	}

	tags := ctx.Request().URL.Query()["tags"]
	if len(tags) > 0 {
		query.Tags = &tags
	}

	isPublicStr := ctx.QueryParam("isPublic")
	if isPublicStr != "" {
		isPublic := isPublicStr == "true"
		query.IsPublic = &isPublic
	}

	svc := service.NewUserDishCollectionService()
	result, err := svc.GetCollections(query)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, result)
}

func (c *UserDishCollectionController) GetCollectionByID(ctx echo.Context) error {
	id := ctx.Param("id")
	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)

	svc := service.NewUserDishCollectionService()
	collection, err := svc.GetCollectionByID(id, claim.ID)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, collection)
}

func (c *UserDishCollectionController) AddDish(ctx echo.Context) error {
	var dto model.AddDishToCollectionDto
	if err := ctx.Bind(&dto); err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}

	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)
	dto.UserID = claim.ID

	svc := service.NewUserDishCollectionService()
	if err := svc.AddDish(dto, claim); err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, map[string]string{"message": "Dish added to collection successfully"})
}

func (c *UserDishCollectionController) RemoveDish(ctx echo.Context) error {
	collectionID := ctx.Param("id")
	dishSlug := ctx.Param("dishSlug")
	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)

	dto := model.RemoveDishFromCollectionDto{
		UserID:       claim.ID,
		CollectionID: collectionID,
		DishSlug:     dishSlug,
	}

	svc := service.NewUserDishCollectionService()
	if err := svc.RemoveDish(dto, claim); err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, map[string]string{"message": "Dish removed from collection successfully"})
}

func (c *UserDishCollectionController) ReorderDishes(ctx echo.Context) error {
	var dto model.ReorderDishesInCollectionDto
	if err := ctx.Bind(&dto); err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}

	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)
	dto.UserID = claim.ID

	svc := service.NewUserDishCollectionService()
	if err := svc.ReorderDishes(dto, claim); err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, map[string]string{"message": "Dishes reordered successfully"})
}

func (c *UserDishCollectionController) Duplicate(ctx echo.Context) error {
	var dto model.DuplicateCollectionDto
	if err := ctx.Bind(&dto); err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}

	claim := ctx.Get("CLAIM").(*model.JwtCustomClaims)
	dto.UserID = claim.ID

	svc := service.NewUserDishCollectionService()
	collection, err := svc.Duplicate(dto, claim)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, collection)
}
