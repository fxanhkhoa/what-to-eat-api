package controller

import (
	"fmt"
	"net/http"
	"what-to-eat/be/service"

	"github.com/labstack/echo/v4"
)

type RolePermission struct{}

func (r *RolePermission) FindByName(c echo.Context) error {
	roleName := c.Param("roleName")
	s := &service.RolePermissionService{}

	fmt.Println(roleName)

	rolePermission, err := s.FindByName(roleName)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return err
	}

	return c.JSON(http.StatusOK, rolePermission)
}
