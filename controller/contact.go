package controller

import (
	"net/http"
	"strconv"
	"what-to-eat/be/config"
	"what-to-eat/be/constants"
	"what-to-eat/be/helper"
	"what-to-eat/be/model"
	"what-to-eat/be/service"

	"github.com/labstack/echo/v4"
)

// contactServiceProvider is the subset of ContactService used by ContactController.
type contactServiceProvider interface {
	Find(query model.QueryContactDto) ([]*model.Contact, int64, error)
	FindOne(id string) (*model.Contact, error)
	Create(dto model.CreateContactDto) (*model.Contact, error)
	Update(dto model.UpdateContactDto, profile *model.JwtCustomClaims) (*model.Contact, error)
	Remove(id string, profile *model.JwtCustomClaims) (*model.Contact, error)
}

type ContactController struct {
	svc contactServiceProvider
}

func NewContactController() *ContactController {
	dbName := config.GetDBInstance().GetDbName()
	col := config.GetDBInstance().GetClient().Database(dbName).Collection(constants.CONTACT_COLLECTION)
	svc := service.NewContactService(service.NewMongoCollectionAdapter(col))
	return &ContactController{svc: svc}
}

// NewContactControllerWithService creates a ContactController with an injected service, for testing.
func NewContactControllerWithService(svc contactServiceProvider) *ContactController {
	return &ContactController{svc: svc}
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

	contacts, count, err := cc.svc.Find(query)
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
	contact, err := cc.svc.FindOne(id)
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
	record, err := cc.svc.Create(dto)
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
	claim := c.Get("CLAIM").(*model.JwtCustomClaims)
	record, err := cc.svc.Update(dto, claim)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, record)
}

func (cc *ContactController) Remove(c echo.Context) error {
	id := c.Param("id")
	claim := c.Get("CLAIM").(*model.JwtCustomClaims)
	record, err := cc.svc.Remove(id, claim)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, record)
}
