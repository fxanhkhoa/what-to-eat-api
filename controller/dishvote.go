package controller

import (
	"net/http"
	"strconv"
	"what-to-eat/be/helper"
	"what-to-eat/be/model"
	"what-to-eat/be/service"

	"github.com/labstack/echo/v4"
)

type DishVoteController struct{}

func (dc *DishVoteController) Find(c echo.Context) error {
	var query model.QueryDishVoteDto

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

	s := &service.DishVoteService{}
	dishes, count, err := s.Find(query)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return err
	}

	return c.JSON(http.StatusOK, helper.PaginationObject{
		Data:  dishes,
		Count: count})
}

func (dvc *DishVoteController) FindOne(c echo.Context) error {
	id := c.Param("id")
	s := &service.DishVoteService{}
	dishVote, err := s.FindOne(id)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return err
	}

	return c.JSON(http.StatusOK, dishVote)
}

func (dvc *DishVoteController) Create(c echo.Context) error {
	var dto model.CreateDishVoteDto

	if err := c.Bind(&dto); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return err
	}

	s := &service.DishVoteService{}
	claim := c.Get("CLAIM").(*model.JwtCustomClaims)
	result, err := s.Create(dto, claim)

	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return err
	}

	return c.JSON(http.StatusOK, result)
}

func (dvc *DishVoteController) Update(c echo.Context) error {

	var dto model.UpdateDishVoteDto

	if err := c.Bind(&dto); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return err
	}

	id := c.Param("id")
	dto.ID = id
	s := &service.DishVoteService{}
	claim := c.Get("CLAIM").(*model.JwtCustomClaims)

	result, err := s.Update(dto, claim)

	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return err
	}

	return c.JSON(http.StatusOK, result)
}
