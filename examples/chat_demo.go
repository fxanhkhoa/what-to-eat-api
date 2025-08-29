package main

import (
	"fmt"
	"log"
	"time"
	"what-to-eat/be/model"
	"what-to-eat/be/service"
)

// Example usage of the chat system
func main() {
	fmt.Println("Chat System Demo")
	fmt.Println("================")

	// Initialize chat service
	chatService := &service.ChatService{}

	// Example 1: Create a message
	fmt.Println("\n1. Creating a test message...")
	createMessageDto := model.CreateChatMessageDto{
		Content:      "Hello from the chat system!",
		Type:         "text",
		RoomID:       "voteGame123",
		SenderID:     "user123",
		SenderName:   "Test User",
		SenderAvatar: nil,
		Reactions:    make(map[string]int),
	}

	message, err := chatService.CreateMessage(createMessageDto)
	if err != nil {
		log.Printf("Error creating message: %v", err)
		return
	}
	fmt.Printf("✅ Message created with ID: %s", message.ID.Hex())

	// Example 2: Create or get a room
	fmt.Println("\n\n2. Creating/getting a chat room...")
	room, err := chatService.CreateOrGetRoom("123", "voteGame")
	if err != nil {
		log.Printf("Error creating room: %v", err)
		return
	}
	fmt.Printf("✅ Room created/retrieved: %s (Type: %s)", room.Name, room.Type)

	// Example 3: Get message history
	fmt.Println("\n\n3. Retrieving message history...")
	messages, err := chatService.GetMessageHistory("voteGame123", 10, 0)
	if err != nil {
		log.Printf("Error getting message history: %v", err)
		return
	}
	fmt.Printf("✅ Retrieved %d messages from history", len(messages))

	for i, msg := range messages {
		fmt.Printf("   Message %d: %s (from %s at %s)",
			i+1,
			msg.Content,
			msg.SenderName,
			time.Unix(int64(msg.Timestamp), 0).Format("15:04:05"))
	}

	// Example 4: Add reactions to a message
	if len(messages) > 0 {
		fmt.Println("\n\n4. Adding reactions to the first message...")
		reactions := map[string]int{
			"👍":  3,
			"❤️": 1,
			"😄":  2,
		}

		updatedMessage, err := chatService.UpdateMessageReactions(messages[0].ID.Hex(), reactions)
		if err != nil {
			log.Printf("Error updating reactions: %v", err)
			return
		}
		fmt.Printf("✅ Updated message reactions: %v", updatedMessage.Reactions)
	}

	fmt.Println("\n\n🎉 Chat system demo completed successfully!")
	fmt.Println("\nNext steps:")
	fmt.Println("- Start the server with the Socket.IO integration")
	fmt.Println("- Connect with a Socket.IO client")
	fmt.Println("- Test real-time messaging features")
}
