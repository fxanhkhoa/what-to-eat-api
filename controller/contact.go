package controller

import (
	"net/http"
	"strconv"
	"what-to-eat/be/helper"
	"what-to-eat/be/model"
	"what-to-eat/be/service"

	"github.com/labstack/echo/v4"
)

type ContactController struct{}

func NewContactController() *ContactController {
	return &ContactController{}
}

func (cc *ContactController) Find(c echo.Context) error {
	var query model.QueryContactDto

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

	s := &service.ContactService{}
	contacts, count, err := s.Find(query)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, helper.PaginationObject{
		Data:  contacts,
		Count: count,
	})
}

func (cc *ContactController) FindOne(c echo.Context) error {
	id := c.Param("id")
	s := &service.ContactService{}
	contact, err := s.FindOne(id)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, contact)
}

func (cc *ContactController) Create(c echo.Context) error {
	var dto model.CreateContactDto
	if err := c.Bind(&dto); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	var service = &service.ContactService{}
	record, err := service.Create(dto)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, record)
}

func (cc *ContactController) Update(c echo.Context) error {
	id := c.Param("id")
	var dto model.UpdateContactDto
	if err := c.Bind(&dto); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	dto.ID = id
	var service = &service.ContactService{}
	claim := c.Get("CLAIM").(*model.JwtCustomClaims)
	record, err := service.Update(dto, claim)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, record)
}

func (cc *ContactController) Remove(c echo.Context) error {
	id := c.Param("id")
	var service = &service.ContactService{}
	claim := c.Get("CLAIM").(*model.JwtCustomClaims)
	record, err := service.Remove(id, claim)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, record)
}
