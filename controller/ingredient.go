package controller

import (
	"net/http"
	"strconv"
	"what-to-eat/be/helper"
	"what-to-eat/be/model"
	"what-to-eat/be/service"

	"github.com/labstack/echo/v4"
)

type IngredientController struct{}

func (ic *IngredientController) Find(c echo.Context) error {
	var query model.QueryIngredientDto

	var err error
	query.BaseDto.Page, err = strconv.Atoi(c.QueryParam("page"))
	if err != nil || query.BaseDto.Page < 0 {
		query.BaseDto.Page = 1
	}

	query.BaseDto.Limit, err = strconv.Atoi(c.QueryParam("limit"))
	if err != nil || query.BaseDto.Limit < 0 {
		query.BaseDto.Limit = 10
	}

	keyword := c.QueryParam("keyword")
	if keyword != "" {
		query.Keyword = &keyword
	}

	q := c.Request().URL.Query()

	ingredientCategories := q["ingredientCategory"]
	query.IngredientCategory = &ingredientCategories

	s := &service.IngredientService{}

	ingredients, count, err := s.Find(query)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return err
	}
	return c.JSON(http.StatusOK, helper.PaginationObject{
		Data:  ingredients,
		Count: count})
}

func (ic *IngredientController) FindRandom(c echo.Context) error {

	limit, err := strconv.Atoi(c.QueryParam("limit"))
	if err != nil || limit < 0 {
		limit = 10
	}

	q := c.Request().URL.Query()

	ingredientCategories := q["ingredientCategory"]

	s := &service.IngredientService{}
	dishes, err := s.Random(&limit, &ingredientCategories)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return err
	}

	return c.JSON(http.StatusOK, dishes)
}

func (ic *IngredientController) FindOne(c echo.Context) error {
	id := c.Param("id")
	service := &service.IngredientService{}
	ingredient, err := service.FindOne(id)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return err
	}
	return c.JSON(http.StatusOK, ingredient)
}

func (ic *IngredientController) FindOneBySlug(c echo.Context) error {
	slug := c.Param("slug")
	service := &service.IngredientService{}
	ingredient, err := service.FindOneBySlug(slug)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return err
	}
	return c.JSON(http.StatusOK, ingredient)
}

func (ic *IngredientController) FindOneByTitleLang(c echo.Context) error {
	title := c.QueryParam("title")
	lang := c.QueryParam("lang")
	service := &service.IngredientService{}
	ingredient, err := service.FindTitleByLang(title, lang)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return err
	}
	return c.JSON(http.StatusOK, ingredient)
}

func (ic *IngredientController) Create(c echo.Context) error {
	var dto model.CreateIngredientDto
	if err := c.Bind(&dto); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	var service = &service.IngredientService{}
	claim := c.Get("CLAIM").(*model.JwtCustomClaims)
	record, err := service.Create(dto, claim)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, record)
}

func (ic *IngredientController) Update(c echo.Context) error {
	id := c.Param("id")
	var dto model.UpdateIngredientDto
	if err := c.Bind(&dto); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	dto.ID = id
	var service = &service.IngredientService{}
	claim := c.Get("CLAIM").(*model.JwtCustomClaims)
	record, err := service.Update(dto, claim)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, record)
}

func (ic *IngredientController) Remove(c echo.Context) error {
	id := c.Param("id")
	var service = &service.IngredientService{}
	claim := c.Get("CLAIM").(*model.JwtCustomClaims)
	record, err := service.Remove(id, claim)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, record)
}
