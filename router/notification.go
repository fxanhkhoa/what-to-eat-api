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

	// Admin: send notification manually to a single user
	group.POST("/send/", ctrl.Send, aG.AuthGuard, rG.RoleGuard([]string{constants.SEND_NOTIFICATIONS}))

	// Admin: broadcast to all users
	group.POST("/broadcast/", ctrl.SendBroadcast, aG.AuthGuard, rG.RoleGuard([]string{constants.BROADCAST_NOTIFICATIONS}))

	// Admin: send to user segment
	group.POST("/segment/", ctrl.SendSegment, aG.AuthGuard, rG.RoleGuard([]string{constants.BROADCAST_NOTIFICATIONS}))

	// Admin: broadcast / segment logs
	group.GET("/admin/logs/", ctrl.GetAdminLogs, aG.AuthGuard, rG.RoleGuard([]string{constants.SEND_NOTIFICATIONS}))

	// Admin: notification templates
	group.POST("/templates/", ctrl.CreateTemplate, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_NOTIFICATION_TEMPLATES}))
	group.GET("/templates/", ctrl.GetTemplates, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_NOTIFICATION_TEMPLATES}))
	group.GET("/templates/:id/", ctrl.GetTemplate, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_NOTIFICATION_TEMPLATES}))
	group.PUT("/templates/:id/", ctrl.UpdateTemplate, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_NOTIFICATION_TEMPLATES}))
	group.DELETE("/templates/:id/", ctrl.DeleteTemplate, aG.AuthGuard, rG.RoleGuard([]string{constants.MANAGE_NOTIFICATION_TEMPLATES}))
}
