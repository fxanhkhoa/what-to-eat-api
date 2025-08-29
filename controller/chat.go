package controller

import (
	"net/http"
	"strconv"
	"what-to-eat/be/model"
	"what-to-eat/be/service"

	"github.com/labstack/echo/v4"
)

type ChatController struct{}

// GetMessageHistory retrieves message history for a chat room
func (cc *ChatController) GetMessageHistory(c echo.Context) error {
	roomID := c.Param("roomId")
	if roomID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "roomId parameter is required",
		})
	}

	// Parse query parameters
	limitStr := c.QueryParam("limit")
	beforeStr := c.QueryParam("before")

	limit := 50 // default limit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}

	before := float64(0) // default to current time
	if beforeStr != "" {
		if parsedBefore, err := strconv.ParseFloat(beforeStr, 64); err == nil {
			before = parsedBefore
		}
	}

	// Get messages from service
	chatService := &service.ChatService{}
	messages, err := chatService.GetMessageHistory(roomID, limit, before)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve message history",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"messages": messages,
		"count":    len(messages),
	})
}

// SendMessage creates a new chat message via REST API
func (cc *ChatController) SendMessage(c echo.Context) error {
	var createMessageDto model.CreateChatMessageDto

	if err := c.Bind(&createMessageDto); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// TODO: Get authenticated user information
	// For now, we'll use the data from the request body

	chatService := &service.ChatService{}
	message, err := chatService.CreateMessage(createMessageDto)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create message",
		})
	}

	return c.JSON(http.StatusCreated, message)
}

// UpdateMessageReaction updates reactions on a message
func (cc *ChatController) UpdateMessageReaction(c echo.Context) error {
	messageID := c.Param("messageId")
	if messageID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "messageId parameter is required",
		})
	}

	var updateDto struct {
		Reactions map[string]int `json:"reactions"`
	}

	if err := c.Bind(&updateDto); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	chatService := &service.ChatService{}
	updatedMessage, err := chatService.UpdateMessageReactions(messageID, updateDto.Reactions)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update message reactions",
		})
	}

	return c.JSON(http.StatusOK, updatedMessage)
}

// GetRoomInfo retrieves information about a chat room
func (cc *ChatController) GetRoomInfo(c echo.Context) error {
	roomID := c.Param("roomId")
	roomType := c.QueryParam("type")

	if roomID == "" || roomType == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "roomId and type parameters are required",
		})
	}

	chatService := &service.ChatService{}
	room, err := chatService.CreateOrGetRoom(roomID, roomType)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get room information",
		})
	}

	// Get online users
	roomName := roomType + roomID
	onlineUsers, _ := chatService.GetOnlineUsers(roomName)
	typingUsers, _ := chatService.GetTypingUsers(roomName)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"room":        room,
		"onlineUsers": onlineUsers,
		"typingUsers": typingUsers,
	})
}
