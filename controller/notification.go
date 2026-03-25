package controller

import (
	"net/http"
	"strconv"
	"what-to-eat/be/helper"
	"what-to-eat/be/model"
	"what-to-eat/be/service"

	"github.com/labstack/echo/v4"
)

type NotificationController struct {
	svc *service.NotificationService
}

func NewNotificationController() *NotificationController {
	svc := service.NewNotificationServiceFromDB()
	return &NotificationController{svc: svc}
}

// RegisterToken registers a device FCM token for the authenticated user
func (nc *NotificationController) RegisterToken(c echo.Context) error {
	claim := c.Get("CLAIM").(*model.JwtCustomClaims)
	var dto model.RegisterDeviceTokenDto
	if err := c.Bind(&dto); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	if err := nc.svc.RegisterDeviceToken(claim.ID, dto); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "token registered"})
}

// UnregisterToken removes a device FCM token for the authenticated user
func (nc *NotificationController) UnregisterToken(c echo.Context) error {
	claim := c.Get("CLAIM").(*model.JwtCustomClaims)
	var dto model.UnregisterDeviceTokenDto
	if err := c.Bind(&dto); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	if err := nc.svc.UnregisterDeviceToken(claim.ID, dto.Token); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "token unregistered"})
}

// GetNotifications returns paginated notification history for the authenticated user
func (nc *NotificationController) GetNotifications(c echo.Context) error {
	claim := c.Get("CLAIM").(*model.JwtCustomClaims)

	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(c.QueryParam("limit"))
	if err != nil || limit < 1 {
		limit = 20
	}

	notifications, total, err := nc.svc.GetUserNotifications(claim.ID, page, limit)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, helper.PaginationObject{
		Data:  notifications,
		Count: total,
	})
}

// GetUnreadCount returns the count of unread notifications
func (nc *NotificationController) GetUnreadCount(c echo.Context) error {
	claim := c.Get("CLAIM").(*model.JwtCustomClaims)
	count, err := nc.svc.GetUnreadCount(claim.ID)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]int64{"count": count})
}

// MarkAsRead marks a single notification as read
func (nc *NotificationController) MarkAsRead(c echo.Context) error {
	id := c.Param("id")
	if err := nc.svc.MarkAsRead(id); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "marked as read"})
}

// MarkAllAsRead marks all notifications as read for the authenticated user
func (nc *NotificationController) MarkAllAsRead(c echo.Context) error {
	claim := c.Get("CLAIM").(*model.JwtCustomClaims)
	if err := nc.svc.MarkAllAsRead(claim.ID); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "all marked as read"})
}

// GetPreferences returns notification preferences for the authenticated user
func (nc *NotificationController) GetPreferences(c echo.Context) error {
	claim := c.Get("CLAIM").(*model.JwtCustomClaims)
	pref, err := nc.svc.GetUserPreferences(claim.ID)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pref)
}

// UpdatePreferences updates notification preferences for the authenticated user
func (nc *NotificationController) UpdatePreferences(c echo.Context) error {
	claim := c.Get("CLAIM").(*model.JwtCustomClaims)
	var dto model.UpdateNotificationPreferenceDto
	if err := c.Bind(&dto); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	pref, err := nc.svc.UpdateUserPreferences(claim.ID, dto)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pref)
}

// Send allows admins to send a notification to a specific user
func (nc *NotificationController) Send(c echo.Context) error {
	var dto model.SendNotificationDto
	if err := c.Bind(&dto); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	if err := nc.svc.SendToUser(dto.UserID, dto.Title, dto.Body, dto.ImageURL, dto.Type, dto.Data); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "notification sent"})
}

// SendBroadcast sends a notification to all users (or schedules it)
func (nc *NotificationController) SendBroadcast(c echo.Context) error {
	claim := c.Get("CLAIM").(*model.JwtCustomClaims)
	var dto model.SendBroadcastDto
	if err := c.Bind(&dto); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	if err := nc.svc.SendBroadcast(dto, claim.ID); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	msg := "broadcast sent"
	if dto.ScheduledAt != nil {
		msg = "broadcast scheduled"
	}
	return c.JSON(http.StatusOK, map[string]string{"message": msg})
}

// SendSegment sends a notification to a filtered user segment (or schedules it)
func (nc *NotificationController) SendSegment(c echo.Context) error {
	claim := c.Get("CLAIM").(*model.JwtCustomClaims)
	var dto model.SendSegmentDto
	if err := c.Bind(&dto); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	if err := nc.svc.SendToSegment(dto, claim.ID); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	msg := "segment notification sent"
	if dto.ScheduledAt != nil {
		msg = "segment notification scheduled"
	}
	return c.JSON(http.StatusOK, map[string]string{"message": msg})
}

// GetAdminLogs returns paginated admin broadcast / segment logs
func (nc *NotificationController) GetAdminLogs(c echo.Context) error {
	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(c.QueryParam("limit"))
	if err != nil || limit < 1 {
		limit = 20
	}
	sentTo := c.QueryParam("sentTo")
	logs, total, err := nc.svc.GetAdminLogs(page, limit, sentTo)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, helper.PaginationObject{Data: logs, Count: total})
}

// CreateTemplate creates a new notification template
func (nc *NotificationController) CreateTemplate(c echo.Context) error {
	claim := c.Get("CLAIM").(*model.JwtCustomClaims)
	var dto model.CreateNotificationTemplateDto
	if err := c.Bind(&dto); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	tmpl, err := nc.svc.CreateTemplate(dto, claim.ID)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, tmpl)
}

// GetTemplates returns paginated notification templates
func (nc *NotificationController) GetTemplates(c echo.Context) error {
	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(c.QueryParam("limit"))
	if err != nil || limit < 1 {
		limit = 50
	}
	templates, total, err := nc.svc.GetTemplates(page, limit)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, helper.PaginationObject{Data: templates, Count: total})
}

// GetTemplate returns a single notification template by ID
func (nc *NotificationController) GetTemplate(c echo.Context) error {
	id := c.Param("id")
	tmpl, err := nc.svc.GetTemplateByID(id)
	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}
	return c.JSON(http.StatusOK, tmpl)
}

// UpdateTemplate updates a notification template by ID
func (nc *NotificationController) UpdateTemplate(c echo.Context) error {
	id := c.Param("id")
	var dto model.UpdateNotificationTemplateDto
	if err := c.Bind(&dto); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	tmpl, err := nc.svc.UpdateTemplate(id, dto)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, tmpl)
}

// DeleteTemplate deletes a notification template by ID
func (nc *NotificationController) DeleteTemplate(c echo.Context) error {
	id := c.Param("id")
	if err := nc.svc.DeleteTemplate(id); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "template deleted"})
}
