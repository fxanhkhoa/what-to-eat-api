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

type NotificationController struct {
	svc *service.NotificationService
}

func NewNotificationController() *NotificationController {
	dbName := config.GetDBInstance().GetDbName()
	db := config.GetDBInstance().GetClient().Database(dbName)
	svc := service.NewNotificationService(
		service.NewMongoCollectionAdapter(db.Collection(constants.NOTIFICATION_COLLECTION)),
		service.NewMongoCollectionAdapter(db.Collection(constants.NOTIFICATION_PREFERENCE_COLLECTION)),
		service.NewMongoCollectionAdapter(db.Collection(constants.USER_COLLECTION)),
	)
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
