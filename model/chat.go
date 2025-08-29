package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChatMessage represents a chat message in the system
type ChatMessage struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Content      string             `json:"content" bson:"content"`
	SenderID     string             `json:"senderId" bson:"senderId"`
	SenderName   string             `json:"senderName" bson:"senderName"`
	SenderAvatar *string            `json:"senderAvatar,omitempty" bson:"senderAvatar,omitempty"`
	Type         string             `json:"type" bson:"type"` // text, image, file, system, vote, poll
	Timestamp    float64            `json:"timestamp" bson:"timestamp"`
	Reactions    map[string]int     `json:"reactions" bson:"reactions"`
	RoomID       string             `json:"roomId" bson:"roomId"`
	CreatedAt    time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt    time.Time          `json:"updatedAt" bson:"updatedAt"`
	Deleted      bool               `json:"deleted" bson:"deleted"`
}

// ChatUser represents a user in chat context
type ChatUser struct {
	ID       string   `json:"id" bson:"_id"`
	Name     string   `json:"name" bson:"name"`
	Avatar   *string  `json:"avatar,omitempty" bson:"avatar,omitempty"`
	IsOnline bool     `json:"isOnline" bson:"isOnline"`
	LastSeen *float64 `json:"lastSeen,omitempty" bson:"lastSeen,omitempty"`
}

// ChatRoom represents a chat room
type ChatRoom struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name         string             `json:"name" bson:"name"`
	Type         string             `json:"type" bson:"type"`                 // voteGame, general, direct, group
	RoomID       string             `json:"roomId" bson:"roomId"`             // External reference ID (e.g., vote game ID)
	Participants []string           `json:"participants" bson:"participants"` // User IDs
	OnlineUsers  []string           `json:"onlineUsers" bson:"onlineUsers"`
	TypingUsers  []string           `json:"typingUsers" bson:"typingUsers"`
	CreatedAt    time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt    time.Time          `json:"updatedAt" bson:"updatedAt"`
	Deleted      bool               `json:"deleted" bson:"deleted"`
}

// Socket.IO Event Models

// SocketioChatJoinRoom represents the data for joining a chat room
type SocketioChatJoinRoom struct {
	SenderId   string  `json:"senderId"`
	SenderName string  `json:"senderName"`
	RoomID     string  `json:"roomId"`
	RoomType   string  `json:"roomType"`
	Timestamp  float64 `json:"timestamp"`
}

// SocketioChatMessage represents the data for sending a chat message
type SocketioChatMessage struct {
	SenderID   string  `json:"senderId"`
	SenderName string  `json:"senderName"`
	Content    string  `json:"content"`
	Type       string  `json:"type"`
	Room       string  `json:"room"`
	Timestamp  float64 `json:"timestamp"`
}

// SocketioChatTyping represents typing indicator data
type SocketioChatTyping struct {
	SenderId string `json:"senderId"`
	Room     string `json:"room"`
}

// SocketioChatMessageHistory represents message history request
type SocketioChatMessageHistory struct {
	Room   string  `json:"room"`
	Limit  int     `json:"limit"`
	Before float64 `json:"before"`
}

// SocketioChatReaction represents message reaction data
type SocketioChatReaction struct {
	MessageID string  `json:"messageId"`
	Reaction  string  `json:"reaction"`
	Room      string  `json:"room"`
	Timestamp float64 `json:"timestamp"`
}

// CreateChatMessageDto for creating new messages
type CreateChatMessageDto struct {
	Content      string         `json:"content"`
	Type         string         `json:"type"`
	RoomID       string         `json:"roomId"`
	SenderID     string         `json:"senderId"`
	SenderName   string         `json:"senderName"`
	SenderAvatar *string        `json:"senderAvatar,omitempty"`
	Reactions    map[string]int `json:"reactions"`
}

// UpdateChatMessageDto for updating messages (mainly for reactions)
type UpdateChatMessageDto struct {
	ID        string         `json:"id"`
	Reactions map[string]int `json:"reactions"`
}
