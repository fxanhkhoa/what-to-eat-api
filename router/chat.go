package router

import (
	controller "what-to-eat/be/controller"
	"what-to-eat/be/middleware"

	"github.com/labstack/echo/v4"
)

func UseChatRouter(group *echo.Group) {
	aG := middleware.NewAuthGuard()
	chatController := &controller.ChatController{}

	// Get message history for a room
	group.GET("/rooms/:roomId/messages", chatController.GetMessageHistory, aG.AuthGuard)

	// Send a message (REST API alternative to Socket.IO)
	group.POST("/messages", chatController.SendMessage, aG.AuthGuard)

	// Update message reactions
	group.PUT("/messages/:messageId/reactions", chatController.UpdateMessageReaction, aG.AuthGuard)

	// Get room information
	group.GET("/rooms/:roomId", chatController.GetRoomInfo, aG.AuthGuard)
}
