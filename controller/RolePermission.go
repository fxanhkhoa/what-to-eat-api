package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"what-to-eat/be/model"
	"what-to-eat/be/service"

	"github.com/labstack/echo/v4"
)

type RolePermission struct{}

func (r *RolePermission) Create(c echo.Context) error {
	var input model.CreateRolePermissionDto
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	// Get user from context (assuming middleware sets it)
	claim := c.Get("CLAIM").(*model.JwtCustomClaims)
	if claim == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	s := &service.RolePermissionService{}
	rolePermission, err := s.Create(input, claim)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, rolePermission)
}

func (r *RolePermission) Update(c echo.Context) error {
	var input model.UpdateRolePermissionDto
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	// Get user from context (assuming middleware sets it)
	claim := c.Get("CLAIM").(*model.JwtCustomClaims)
	if claim == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	s := &service.RolePermissionService{}
	rolePermission, err := s.Update(input, claim)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, rolePermission)
}

func (r *RolePermission) Remove(c echo.Context) error {
	id := c.Param("id")

	// Get user from context (assuming middleware sets it)
	claim := c.Get("CLAIM").(*model.JwtCustomClaims)
	if claim == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	s := &service.RolePermissionService{}
	rolePermission, err := s.Remove(id, claim)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, rolePermission)
}

func (r *RolePermission) FindAll(c echo.Context) error {
	// Get pagination parameters
	pageStr := c.QueryParam("page")
	limitStr := c.QueryParam("limit")

	page := 1
	limit := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	s := &service.RolePermissionService{}
	rolePermissions, count, err := s.Find(&page, &limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  rolePermissions,
		"count": count,
	})
}

func (r *RolePermission) FindOne(c echo.Context) error {
	id := c.Param("id")
	s := &service.RolePermissionService{}

	rolePermission, err := s.FindOne(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, rolePermission)
}

func (r *RolePermission) FindByName(c echo.Context) error {
	roleName := c.Param("roleName")
	s := &service.RolePermissionService{}

	fmt.Println(roleName)

	rolePermission, err := s.FindByName(roleName)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, rolePermission)
}

func (r *RolePermission) GetAllPermissions(c echo.Context) error {
	s := &service.RolePermissionService{}
	permissions := s.GetAllPermissions()

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  permissions,
		"count": len(permissions),
	})
}
