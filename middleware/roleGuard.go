package middleware

import (
	"fmt"
	"slices"
	"sync"
	"what-to-eat/be/model"
	"what-to-eat/be/service"

	"github.com/labstack/echo/v4"
)

type RoleGuard struct {
	mutex sync.RWMutex
}

func NewRoleGuard() *RoleGuard {
	return &RoleGuard{}
}

func (rg *RoleGuard) RoleGuard(permissions []string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			rg.mutex.Lock()
			defer rg.mutex.Unlock()

			rs := service.RolePermissionService{}
			claims := c.Get("CLAIM").(*model.JwtCustomClaims)
			foundRole, err := rs.FindByName(claims.RoleName)
			if err != nil {
				fmt.Println(err.Error())
				return echo.ErrForbidden
			}

			for _, permission := range permissions {
				if !slices.Contains(foundRole.Permission, permission) {
					return echo.ErrForbidden
				}
			}

			if err := next(c); err != nil {
				c.Error(err)
			}

			return nil
		}
	}
}

func (rg *RoleGuard) Permission(permissions []string) echo.HandlerFunc {
	return func(c echo.Context) error {

		rg.mutex.Lock()
		defer rg.mutex.Unlock()

		rs := service.RolePermissionService{}
		claims := c.Get("CLAIM").(*model.JwtCustomClaims)
		foundRole, err := rs.FindByName(claims.Name)
		if err != nil {
			return echo.ErrForbidden
		}

		for _, permission := range permissions {
			if !slices.Contains(foundRole.Permission, permission) {
				return echo.ErrForbidden
			}
		}

		return nil
	}
}
