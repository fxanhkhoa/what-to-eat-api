package socketio_service

import (
	"encoding/json"
	"log"
	"time"
	"what-to-eat/be/model"
	"what-to-eat/be/service"

	"github.com/zishang520/socket.io/socket"
)

type ChatSocketService struct {
	chatService *service.ChatService
}

func NewChatSocketService() *ChatSocketService {
	return &ChatSocketService{
		chatService: &service.ChatService{},
	}
}

// ProcessJoinChatRoom handles user joining a chat room
func (css *ChatSocketService) ProcessJoinChatRoom(client *socket.Socket, data ...any) (*model.ChatRoom, error) {
	jsonStr, err := json.Marshal(data[0])
	if err != nil {
		log.Printf("Failed to marshal join-chat-room data: %v", err)
		return nil, err
	}

	var joinRoomData model.SocketioChatJoinRoom
	if err := json.Unmarshal(jsonStr, &joinRoomData); err != nil {
		log.Printf("Failed to unmarshal join-chat-room data: %v", err)
		return nil, err
	}

	// Create room name based on type and ID
	roomName := joinRoomData.RoomType + joinRoomData.RoomID

	// Create or get the room
	room, err := css.chatService.CreateOrGetRoom(joinRoomData.RoomID, joinRoomData.RoomType)
	if err != nil {
		log.Printf("Failed to create/get chat room: %v", err)
		return nil, err
	}

	// Join the socket room
	client.Join(socket.Room(roomName))

	// Add user to room (assuming we can get user ID from client)
	err = css.chatService.AddUserToRoom(roomName, joinRoomData.SenderId)
	if err != nil {
		log.Printf("Failed to add user to room: %v", err)
	}

	log.Printf("User %s joined chat room: %s", joinRoomData.SenderId, roomName)
	return room, nil
}

// ProcessLeaveChatRoom handles user leaving a chat room
func (css *ChatSocketService) ProcessLeaveChatRoom(client *socket.Socket, roomName string, senderId string) error {
	// Remove user from room
	err := css.chatService.RemoveUserFromRoom(roomName, senderId)
	if err != nil {
		log.Printf("Failed to remove user from room: %v", err)
		return err
	}

	// Leave the socket room
	client.Leave(socket.Room(roomName))

	log.Printf("User %s left chat room: %s", senderId, roomName)
	return nil
}

// ProcessSendMessage handles sending a new chat message
func (css *ChatSocketService) ProcessSendMessage(client *socket.Socket, data ...any) (*model.ChatMessage, error) {
	jsonStr, err := json.Marshal(data[0])
	if err != nil {
		log.Printf("Failed to marshal send-message data: %v", err)
		return nil, err
	}

	var messageData model.SocketioChatMessage
	if err := json.Unmarshal(jsonStr, &messageData); err != nil {
		log.Printf("Failed to unmarshal send-message data: %v", err)
		return nil, err
	}

	// Create message DTO
	createMessageDto := model.CreateChatMessageDto{
		Content:      messageData.Content,
		Type:         messageData.Type,
		RoomID:       messageData.Room,
		SenderID:     messageData.SenderID,
		SenderName:   messageData.SenderName,
		SenderAvatar: nil,
		Reactions:    make(map[string]int),
	}

	// Save message to database
	message, err := css.chatService.CreateMessage(createMessageDto)
	if err != nil {
		log.Printf("Failed to create message: %v", err)
		return nil, err
	}

	log.Printf("Message sent to room %s by user %s", messageData.Room, client.Id())
	return message, nil
}

// ProcessTypingStart handles user starting to type
func (css *ChatSocketService) ProcessTypingStart(client *socket.Socket, data ...any) error {
	jsonStr, err := json.Marshal(data[0])
	if err != nil {
		log.Printf("Failed to marshal typing-start data: %v", err)
		return err
	}

	var typingData model.SocketioChatTyping
	if err := json.Unmarshal(jsonStr, &typingData); err != nil {
		log.Printf("Failed to unmarshal typing-start data: %v", err)
		return err
	}

	// Convert SocketId to string
	err = css.chatService.UpdateTypingUsers(typingData.Room, typingData.SenderId, true)
	if err != nil {
		log.Printf("Failed to update typing users: %v", err)
		return err
	}

	log.Printf("User %s started typing in room: %s", typingData.SenderId, typingData.Room)
	return nil
}

// ProcessTypingStop handles user stopping typing
func (css *ChatSocketService) ProcessTypingStop(client *socket.Socket, data ...any) error {
	jsonStr, err := json.Marshal(data[0])
	if err != nil {
		log.Printf("Failed to marshal typing-stop data: %v", err)
		return err
	}

	var typingData model.SocketioChatTyping
	if err := json.Unmarshal(jsonStr, &typingData); err != nil {
		log.Printf("Failed to unmarshal typing-stop data: %v", err)
		return err
	}

	// Convert SocketId to string
	err = css.chatService.UpdateTypingUsers(typingData.Room, typingData.SenderId, false)
	if err != nil {
		log.Printf("Failed to update typing users: %v", err)
		return err
	}

	log.Printf("User %s stopped typing in room: %s", typingData.SenderId, typingData.Room)
	return nil
}

// ProcessGetMessageHistory handles fetching message history
func (css *ChatSocketService) ProcessGetMessageHistory(client *socket.Socket, data ...any) ([]*model.ChatMessage, error) {
	jsonStr, err := json.Marshal(data[0])
	if err != nil {
		log.Printf("Failed to marshal get-message-history data: %v", err)
		return nil, err
	}

	var historyData model.SocketioChatMessageHistory
	if err := json.Unmarshal(jsonStr, &historyData); err != nil {
		log.Printf("Failed to unmarshal get-message-history data: %v", err)
		return nil, err
	}

	// Get message history from database
	messages, err := css.chatService.GetMessageHistory(historyData.Room, historyData.Limit, historyData.Before)
	if err != nil {
		log.Printf("Failed to get message history: %v", err)
		return nil, err
	}

	log.Printf("Retrieved %d messages for room: %s", len(messages), historyData.Room)
	return messages, nil
}

// ProcessMessageReaction handles message reactions
func (css *ChatSocketService) ProcessMessageReaction(client *socket.Socket, data ...any) (*model.ChatMessage, error) {
	jsonStr, err := json.Marshal(data[0])
	if err != nil {
		log.Printf("Failed to marshal message-reaction data: %v", err)
		return nil, err
	}

	var reactionData model.SocketioChatReaction
	if err := json.Unmarshal(jsonStr, &reactionData); err != nil {
		log.Printf("Failed to unmarshal message-reaction data: %v", err)
		return nil, err
	}

	// Get current message
	message, err := css.chatService.FindMessageByID(reactionData.MessageID)
	if err != nil {
		log.Printf("Failed to find message: %v", err)
		return nil, err
	}

	// Update reactions (simple increment for now)
	if message.Reactions == nil {
		message.Reactions = make(map[string]int)
	}
	message.Reactions[reactionData.Reaction]++

	// Save updated reactions
	updatedMessage, err := css.chatService.UpdateMessageReactions(reactionData.MessageID, message.Reactions)
	if err != nil {
		log.Printf("Failed to update message reactions: %v", err)
		return nil, err
	}

	log.Printf("Reaction %s added to message %s", reactionData.Reaction, reactionData.MessageID)
	return updatedMessage, nil
}

// GetOnlineUsers returns online users for a room
func (css *ChatSocketService) GetOnlineUsers(roomName string) ([]string, error) {
	return css.chatService.GetOnlineUsers(roomName)
}

// GetTypingUsers returns typing users for a room
func (css *ChatSocketService) GetTypingUsers(roomName string) ([]string, error) {
	return css.chatService.GetTypingUsers(roomName)
}

// Helper function to create user data for socket events
func (css *ChatSocketService) CreateUserData(userID, userName string) map[string]interface{} {
	return map[string]interface{}{
		"userId":    userID,
		"userName":  userName,
		"timestamp": time.Now().Unix(),
	}
}
