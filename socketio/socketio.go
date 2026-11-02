package socketio

import (
	"encoding/json"
	"log"
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
		log.Printf("[Socket.IO] New client connected: %s", client.Id())

		// Existing dish voting functionality
		client.On("join-room", func(datas ...any) {
			log.Printf("[join-room] Client %s triggered join-room event", client.Id())
			jsonStr, err := json.Marshal(datas[0])
			if err != nil {
				log.Printf("[join-room] Failed to marshal join-room data for client %s: %v", client.Id(), err)
				return
			}
			log.Printf("[join-room] Received data: %s", string(jsonStr))
			var socketioJoinRoomData model.SocketioJoinRoom
			if err := json.Unmarshal(jsonStr, &socketioJoinRoomData); err != nil {
				log.Printf("[join-room] Failed to unmarshal join-room data for client %s: %v", client.Id(), err)
				return
			}

			client.Join(socket.Room(socketioJoinRoomData.RoomID))
			log.Printf("[join-room] Client %s joined room: %s", client.Id(), socketioJoinRoomData.RoomID)
		})

		client.On("dish-vote-update", func(data ...any) {
			log.Printf("[dish-vote-update] Client %s triggered dish-vote-update event", client.Id())
			updated, socketioJoinRoomData := socketio_service.ProcessDishVoteUpdate(data...)
			io.To(socket.Room(socketioJoinRoomData.RoomID)).Emit("dish-vote-update-client", updated)
			log.Printf("[dish-vote-update] Vote update broadcasted to room: %s", socketioJoinRoomData.RoomID)
		})

		// ========== CHAT FUNCTIONALITY ==========

		// Join chat room
		client.On("join_chat_room", func(datas ...any) {
			log.Printf("[join_chat_room] Client %s triggered join_chat_room event", client.Id())
			room, err := chatService.ProcessJoinChatRoom(client, datas...)
			if err != nil {
				log.Printf("[join_chat_room] Failed to join chat room for client %s: %v", client.Id(), err)
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

			log.Printf("[join_chat_room] Client %s joined chat room: %s, sender: %s", client.Id(), roomName, joinRoomData.SenderName)
			io.To(socket.Room(roomName)).Emit("user_joined_chat", userData)
			io.To(socket.Room(roomName)).Emit("chat_room_updated", map[string]interface{}{
				"room":        room,
				"onlineUsers": onlineUsers,
			})
			log.Printf("[join_chat_room] Notified %d online users in room %s", len(onlineUsers), roomName)
		})

		// Leave chat room
		client.On("leave_chat_room", func(datas ...any) {
			log.Printf("[leave_chat_room] Client %s triggered leave_chat_room event", client.Id())
			jsonStr, _ := json.Marshal(datas[0])
			var roomData map[string]string
			json.Unmarshal(jsonStr, &roomData)
			roomName := roomData["room"]
			senderId := roomData["senderId"]

			err := chatService.ProcessLeaveChatRoom(client, roomName, senderId)
			if err != nil {
				log.Printf("[leave_chat_room] Failed to leave chat room %s for client %s: %v", roomName, client.Id(), err)
				return
			}
			log.Printf("[leave_chat_room] Client %s left room: %s, sender: %s", client.Id(), roomName, senderId)

			// Notify other users
			userData := chatService.CreateUserData(senderId, "Anonymous")
			io.To(socket.Room(roomName)).Emit("user_left_chat", userData)
			log.Printf("[leave_chat_room] Notified users in room %s about departure", roomName)
		})

		// Send message
		client.On("send_message", func(datas ...any) {
			log.Printf("[send_message] Client %s triggered send_message event", client.Id())
			message, err := chatService.ProcessSendMessage(client, datas...)
			if err != nil {
				log.Printf("[send_message] Failed to send message for client %s: %v", client.Id(), err)
				return
			}

			// Broadcast message to room
			io.To(socket.Room(message.RoomID)).Emit("message_received", message)
			log.Printf("[send_message] Message broadcasted to room %s, message ID: %s", message.RoomID, message.ID.Hex())
		})

		// Typing indicators
		client.On("typing_start", func(datas ...any) {
			log.Printf("[typing_start] Client %s triggered typing_start event", client.Id())
			err := chatService.ProcessTypingStart(client, datas...)
			if err != nil {
				log.Printf("[typing_start] Failed to process typing start for client %s: %v", client.Id(), err)
				return
			}

			jsonStr, _ := json.Marshal(datas[0])
			var typingData model.SocketioChatTyping
			json.Unmarshal(jsonStr, &typingData)

			userData := chatService.CreateUserData(typingData.SenderId, "Anonymous")
			client.To(socket.Room(typingData.Room)).Emit("user_typing_start", userData)
			log.Printf("[typing_start] User %s started typing in room %s", typingData.SenderId, typingData.Room)
		})

		client.On("typing_stop", func(datas ...any) {
			log.Printf("[typing_stop] Client %s triggered typing_stop event", client.Id())
			err := chatService.ProcessTypingStop(client, datas...)
			if err != nil {
				log.Printf("[typing_stop] Failed to process typing stop for client %s: %v", client.Id(), err)
				return
			}

			jsonStr, _ := json.Marshal(datas[0])
			var typingData model.SocketioChatTyping
			json.Unmarshal(jsonStr, &typingData)

			userData := chatService.CreateUserData(typingData.SenderId, "Anonymous")
			client.To(socket.Room(typingData.Room)).Emit("user_typing_stop", userData)
			log.Printf("[typing_stop] User %s stopped typing in room %s", typingData.SenderId, typingData.Room)
		})

		// Message history
		client.On("get_message_history", func(datas ...any) {
			log.Printf("[get_message_history] Client %s requested message history", client.Id())
			messages, err := chatService.ProcessGetMessageHistory(client, datas...)
			if err != nil {
				log.Printf("[get_message_history] Failed to get message history for client %s: %v", client.Id(), err)
				return
			}

			client.Emit("message_history", map[string]interface{}{
				"messages": messages,
			})
			log.Printf("[get_message_history] Sent %d messages to client %s", len(messages), client.Id())
		})

		// Message reactions
		client.On("message_reaction", func(datas ...any) {
			log.Printf("[message_reaction] Client %s triggered message_reaction event", client.Id())
			updatedMessage, err := chatService.ProcessMessageReaction(client, datas...)
			if err != nil {
				log.Printf("[message_reaction] Failed to process message reaction for client %s: %v", client.Id(), err)
				return
			}

			// Broadcast updated reactions to room
			io.To(socket.Room(updatedMessage.RoomID)).Emit("message_reaction_updated", map[string]interface{}{
				"messageId": updatedMessage.ID.Hex(),
				"reactions": updatedMessage.Reactions,
			})
			log.Printf("[message_reaction] Reaction updated for message %s and broadcasted to room %s", updatedMessage.ID.Hex(), updatedMessage.RoomID)
		})

		// Handle disconnect
		client.On("disconnect", func(...any) {
			log.Printf("[disconnect] Client disconnected: %s", client.Id())
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
