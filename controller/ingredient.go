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
