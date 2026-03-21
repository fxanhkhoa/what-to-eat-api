package router

import (
	"what-to-eat/be/constants"
	"what-to-eat/be/controller"
	"what-to-eat/be/middleware"

	"github.com/labstack/echo/v4"
)

func UseNotificationRouter(group *echo.Group) {
	aG := middleware.NewAuthGuard()
	rG := middleware.NewRoleGuard()
	ctrl := controller.NewNotificationController()

	// Device token management
	group.POST("/register-token/", ctrl.RegisterToken, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_NOTIFICATIONS}))
	group.POST("/unregister-token/", ctrl.UnregisterToken, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_NOTIFICATIONS}))

	// Notification history
	group.GET("/", ctrl.GetNotifications, aG.AuthGuard, rG.RoleGuard([]string{constants.VIEW_NOTIFICATIONS}))
	group.GET("/unread-count/", ctrl.GetUnreadCount, aG.AuthGuard, rG.RoleGuard([]string{constants.VIEW_NOTIFICATIONS}))
	group.PUT("/:id/read/", ctrl.MarkAsRead, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_NOTIFICATIONS}))
	group.PUT("/read-all/", ctrl.MarkAllAsRead, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_NOTIFICATIONS}))

	// Preferences
	group.GET("/preferences/", ctrl.GetPreferences, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_NOTIFICATIONS}))
	group.PUT("/preferences/", ctrl.UpdatePreferences, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_NOTIFICATIONS}))

	// Admin: send notification manually
	group.POST("/send/", ctrl.Send, aG.AuthGuard, rG.RoleGuard([]string{constants.SEND_NOTIFICATIONS}))
}
