package socketio

import (
	"encoding/json"
	"fmt"
	"what-to-eat/be/model"
	socketio_service "what-to-eat/be/socketio/service"

	"github.com/labstack/echo/v4"
	"github.com/zishang520/socket.io/socket"
)

func InitializeSocketIO(e *echo.Echo) *socket.Server {
	// Configure Socket.IO server for WebSocket transport
	io := socket.NewServer(nil, nil)

	// Initialize chat service
	chatService := socketio_service.NewChatSocketService()

	io.On("connection", func(clients ...any) {
		client := clients[0].(*socket.Socket)

		// Existing dish voting functionality
		client.On("join-room", func(datas ...any) {
			jsonStr, err := json.Marshal(datas[0])
			if err != nil {
				fmt.Printf("Failed to marshal join-room data: %v\n", err)
			}
			var socketioJoinRoomData model.SocketioJoinRoom
			if err := json.Unmarshal(jsonStr, &socketioJoinRoomData); err != nil {
				fmt.Printf("Failed to unmarshal join-room data: %v\n", err)
			}

			client.Join(socket.Room(socketioJoinRoomData.RoomID))
		})

		client.On("dish-vote-update", func(data ...any) {
			updated, socketioJoinRoomData := socketio_service.ProcessDishVoteUpdate(data...)
			io.To(socket.Room(socketioJoinRoomData.RoomID)).Emit("dish-vote-update-client", updated)
		})

		// ========== CHAT FUNCTIONALITY ==========

		// Join chat room
		client.On("join_chat_room", func(datas ...any) {
			room, err := chatService.ProcessJoinChatRoom(client, datas...)
			if err != nil {
				fmt.Printf("Failed to join chat room: %v\n", err)
				return
			}

			// Get room name from data
			jsonStr, _ := json.Marshal(datas[0])
			var joinRoomData model.SocketioChatJoinRoom
			json.Unmarshal(jsonStr, &joinRoomData)
			roomName := joinRoomData.RoomType + joinRoomData.RoomID

			// Notify other users in the room
			onlineUsers, _ := chatService.GetOnlineUsers(roomName)
			userData := chatService.CreateUserData(joinRoomData.SenderId, joinRoomData.SenderName)

			io.To(socket.Room(roomName)).Emit("user_joined_chat", userData)
			io.To(socket.Room(roomName)).Emit("chat_room_updated", map[string]interface{}{
				"room":        room,
				"onlineUsers": onlineUsers,
			})
		})

		// Leave chat room
		client.On("leave_chat_room", func(datas ...any) {
			jsonStr, _ := json.Marshal(datas[0])
			var roomData map[string]string
			json.Unmarshal(jsonStr, &roomData)
			roomName := roomData["room"]
			senderId := roomData["senderId"]

			err := chatService.ProcessLeaveChatRoom(client, roomName, senderId)
			if err != nil {
				fmt.Printf("Failed to leave chat room: %v\n", err)
				return
			}

			// Notify other users
			userData := chatService.CreateUserData(senderId, "Anonymous")
			io.To(socket.Room(roomName)).Emit("user_left_chat", userData)
		})

		// Send message
		client.On("send_message", func(datas ...any) {
			message, err := chatService.ProcessSendMessage(client, datas...)
			if err != nil {
				fmt.Printf("Failed to send message: %v\n", err)
				return
			}

			// Broadcast message to room
			io.To(socket.Room(message.RoomID)).Emit("message_received", message)
		})

		// Typing indicators
		client.On("typing_start", func(datas ...any) {
			err := chatService.ProcessTypingStart(client, datas...)
			if err != nil {
				fmt.Printf("Failed to process typing start: %v\n", err)
				return
			}

			jsonStr, _ := json.Marshal(datas[0])
			var typingData model.SocketioChatTyping
			json.Unmarshal(jsonStr, &typingData)

			userData := chatService.CreateUserData(typingData.SenderId, "Anonymous")
			client.To(socket.Room(typingData.Room)).Emit("user_typing_start", userData)
		})

		client.On("typing_stop", func(datas ...any) {
			err := chatService.ProcessTypingStop(client, datas...)
			if err != nil {
				fmt.Printf("Failed to process typing stop: %v\n", err)
				return
			}

			jsonStr, _ := json.Marshal(datas[0])
			var typingData model.SocketioChatTyping
			json.Unmarshal(jsonStr, &typingData)

			userData := chatService.CreateUserData(typingData.SenderId, "Anonymous")
			client.To(socket.Room(typingData.Room)).Emit("user_typing_stop", userData)
		})

		// Message history
		client.On("get_message_history", func(datas ...any) {
			messages, err := chatService.ProcessGetMessageHistory(client, datas...)
			if err != nil {
				fmt.Printf("Failed to get message history: %v\n", err)
				return
			}

			client.Emit("message_history", map[string]interface{}{
				"messages": messages,
			})
		})

		// Message reactions
		client.On("message_reaction", func(datas ...any) {
			updatedMessage, err := chatService.ProcessMessageReaction(client, datas...)
			if err != nil {
				fmt.Printf("Failed to process message reaction: %v\n", err)
				return
			}

			// Broadcast updated reactions to room
			io.To(socket.Room(updatedMessage.RoomID)).Emit("message_reaction_updated", map[string]interface{}{
				"messageId": updatedMessage.ID.Hex(),
				"reactions": updatedMessage.Reactions,
			})
		})

		// Handle disconnect
		client.On("disconnect", func(...any) {
			fmt.Printf("Client disconnected: %s\n", client.Id())
			// TODO: Clean up user from all rooms they were in
		})
	})

	// Serve Socket.IO over WebSocket and HTTP
	e.Any("/socket.io/*", func(c echo.Context) error {
		// The Socket.IO server handles both WebSocket upgrades and HTTP polling
		handler := io.ServeHandler(nil)
		handler.ServeHTTP(c.Response().Writer, c.Request())
		return nil
	})

	return io
}
