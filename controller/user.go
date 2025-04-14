package controller

import (
	"math"
	"net/http"
	"strconv"
	"what-to-eat/be/model"
	"what-to-eat/be/service"

	"github.com/labstack/echo/v4"
)

type UserController struct{}

func (cr *UserController) Create(c echo.Context) error {
	var dto model.CreateUserDto
	if err := c.Bind(&dto); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	var service = &service.UserService{}
	claim := c.Get("CLAIM").(*model.JwtCustomClaims)
	record, err := service.Create(dto, claim)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, record)
}

func (cr *UserController) FindAll(c echo.Context) error {
	var query model.QueryUserDto

	var err error
	query.BaseDto.Page, err = strconv.Atoi(c.QueryParam("page"))
	if err != nil || query.BaseDto.Page < 0 {
		query.BaseDto.Page = 1
	}

	query.BaseDto.Limit, err = strconv.Atoi(c.QueryParam("limit"))
	if err != nil || query.BaseDto.Limit < 0 {
		query.BaseDto.Limit = 10
	}
	query.Keyword = c.QueryParam("keyword")

	var service = &service.UserService{}
	records, count, err := service.FindAll(query)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	result := model.PaginationResponse{
		Data: &records,
		Metadata: model.CountMetaData{
			TotalItems:   count,
			ItemCount:    len(records),
			ItemsPerPage: query.BaseDto.Limit,
			TotalPages:   math.Ceil(float64(count) / float64(query.BaseDto.Limit)),
			CurrentPage:  query.BaseDto.Page,
		},
	}
	return c.JSON(http.StatusOK, result)
}

func (cr *UserController) FindOne(c echo.Context) error {
	id := c.Param("id")
	var service = &service.UserService{}
	record, err := service.FindOne(id)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, record)
}

func (cr *UserController) Update(c echo.Context) error {
	id := c.Param("id")
	var dto model.UpdateUserDto
	if err := c.Bind(&dto); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	dto.ID = id
	var service = &service.UserService{}
	claim := c.Get("CLAIM").(*model.JwtCustomClaims)
	record, err := service.Update(dto, claim)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, record)
}

func (cr *UserController) Delete(c echo.Context) error {
	id := c.Param("id")
	var service = &service.UserService{}
	claim := c.Get("CLAIM").(*model.JwtCustomClaims)
	record, err := service.Remove(id, claim)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, record)
}
