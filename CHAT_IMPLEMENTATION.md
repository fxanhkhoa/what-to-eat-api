# Chat System Implementation - Go Backend

This document outlines the chat system implementation that matches the Swift client functionality for the what-to-eat application.

## Overview

The chat system supports real-time messaging using Socket.IO with the following features:
- Room-based chat (vote games, general chat, direct messages, groups)
- Real-time message sending and receiving
- Typing indicators
- Message reactions
- Message history
- Online user tracking

## Architecture

### Models (`model/chat.go`)

#### Core Models
- **ChatMessage**: Represents a chat message with content, sender info, timestamp, reactions, etc.
- **ChatUser**: Represents a user in chat context with online status
- **ChatRoom**: Represents a chat room with participants and state
- **Socket.IO Event Models**: Various DTOs for different socket events

### Services (`service/chat.go`)

#### ChatService
Handles database operations for chat functionality:
- `CreateMessage()`: Store new messages
- `GetMessageHistory()`: Retrieve message history with pagination
- `UpdateMessageReactions()`: Update message reactions
- `CreateOrGetRoom()`: Create or retrieve chat rooms
- `AddUserToRoom()`: Add users to rooms
- `UpdateTypingUsers()`: Manage typing indicators

### Socket.IO Service (`socketio/service/chat.go`)

#### ChatSocketService
Handles Socket.IO event processing:
- `ProcessJoinChatRoom()`: Handle user joining rooms
- `ProcessSendMessage()`: Handle message sending
- `ProcessTypingStart/Stop()`: Handle typing indicators
- `ProcessGetMessageHistory()`: Handle history requests
- `ProcessMessageReaction()`: Handle message reactions

### Controllers (`controller/chat.go`)

#### ChatController
REST API endpoints for chat functionality:
- `GET /chat/rooms/:roomId/messages`: Get message history
- `POST /chat/messages`: Send message via REST
- `PUT /chat/messages/:messageId/reactions`: Update reactions
- `GET /chat/rooms/:roomId`: Get room info

## Socket.IO Events

### Client → Server Events

#### 1. Join Chat Room
```javascript
socket.emit("join_chat_room", {
    roomId: "vote-game-123",
    roomType: "voteGame",
    timestamp: Date.now() / 1000
});
```

#### 2. Leave Chat Room
```javascript
socket.emit("leave_chat_room", {
    room: "voteGamevote-game-123"
});
```

#### 3. Send Message
```javascript
socket.emit("send_message", {
    content: "Hello everyone!",
    type: "text",
    room: "voteGamevote-game-123",
    timestamp: Date.now() / 1000
});
```

#### 4. Typing Indicators
```javascript
// Start typing
socket.emit("typing_start", {
    room: "voteGamevote-game-123"
});

// Stop typing
socket.emit("typing_stop", {
    room: "voteGamevote-game-123"
});
```

#### 5. Get Message History
```javascript
socket.emit("get_message_history", {
    room: "voteGamevote-game-123",
    limit: 50,
    before: Date.now() / 1000
});
```

#### 6. React to Message
```javascript
socket.emit("message_reaction", {
    messageId: "64f8a1b2c3d4e5f6a7b8c9d0",
    reaction: "👍",
    room: "voteGamevote-game-123",
    timestamp: Date.now() / 1000
});
```

### Server → Client Events

#### 1. Message Received
```javascript
socket.on("message_received", (message) => {
    // Handle new message
    console.log("New message:", message);
});
```

#### 2. User Joined/Left Chat
```javascript
socket.on("user_joined_chat", (userData) => {
    // Handle user joining
    console.log("User joined:", userData);
});

socket.on("user_left_chat", (userData) => {
    // Handle user leaving
    console.log("User left:", userData);
});
```

#### 3. Typing Indicators
```javascript
socket.on("user_typing_start", (userData) => {
    // Show typing indicator
    console.log("User started typing:", userData);
});

socket.on("user_typing_stop", (userData) => {
    // Hide typing indicator
    console.log("User stopped typing:", userData);
});
```

#### 4. Message History
```javascript
socket.on("message_history", (data) => {
    // Load message history
    console.log("Message history:", data.messages);
});
```

#### 5. Reaction Updates
```javascript
socket.on("message_reaction_updated", (data) => {
    // Update message reactions
    console.log("Reactions updated:", data);
});
```

#### 6. Room Updates
```javascript
socket.on("chat_room_updated", (roomData) => {
    // Handle room state changes
    console.log("Room updated:", roomData);
});
```

## Room Types and Naming Convention

### Room Types
- `voteGame`: Chat associated with voting games
- `general`: General chat rooms
- `direct`: Direct messages between users
- `group`: Group chat rooms

### Room Naming
Rooms are named by concatenating the room type with the room ID:
- Vote game room: `"voteGame" + "123"` = `"voteGame123"`
- General room: `"general" + "main"` = `"generalmain"`

## Database Collections

### chatMessages
```javascript
{
    _id: ObjectId,
    content: "string",
    senderId: "string",
    senderName: "string",
    senderAvatar: "string",
    type: "text|image|file|system|vote|poll",
    timestamp: 1693824000.123,
    reactions: {"👍": 5, "❤️": 2},
    roomId: "voteGame123",
    createdAt: ISODate,
    updatedAt: ISODate,
    deleted: false
}
```

### chatRooms
```javascript
{
    _id: ObjectId,
    name: "voteGame123",
    type: "voteGame",
    roomId: "123",
    participants: ["user1", "user2"],
    onlineUsers: ["user1"],
    typingUsers: ["user2"],
    createdAt: ISODate,
    updatedAt: ISODate,
    deleted: false
}
```

## REST API Endpoints

### Get Message History
```http
GET /chat/rooms/:roomId/messages?limit=50&before=1693824000
Authorization: Bearer <token>
```

### Send Message
```http
POST /chat/messages
Authorization: Bearer <token>
Content-Type: application/json

{
    "content": "Hello!",
    "type": "text",
    "roomId": "voteGame123",
    "senderId": "user123",
    "senderName": "John Doe",
    "senderAvatar": "avatar_url"
}
```

### Update Reactions
```http
PUT /chat/messages/:messageId/reactions
Authorization: Bearer <token>
Content-Type: application/json

{
    "reactions": {"👍": 6, "❤️": 2}
}
```

### Get Room Info
```http
GET /chat/rooms/:roomId?type=voteGame
Authorization: Bearer <token>
```

## Integration with Swift Client

The Swift `ChatSocketService` class maps to these server events:

| Swift Method | Server Event | Description |
|--------------|-------------|-------------|
| `joinChatRoom()` | `join_chat_room` | Join a chat room |
| `leaveChatRoom()` | `leave_chat_room` | Leave a chat room |
| `sendMessage()` | `send_message` | Send a new message |
| `startTyping()` | `typing_start` | Start typing indicator |
| `stopTyping()` | `typing_stop` | Stop typing indicator |
| `loadMessageHistory()` | `get_message_history` | Load message history |
| `reactToMessage()` | `message_reaction` | React to a message |

## Authentication & Authorization

- Socket.IO connections should include authentication tokens
- REST endpoints are protected with `AuthGuard` middleware
- User information is extracted from JWT tokens
- Room access can be controlled based on user permissions

## Error Handling

- Invalid data formats return appropriate error responses
- Database errors are logged and return generic error messages
- Socket events acknowledge errors back to clients
- Connection errors trigger reconnection logic on the client

## Future Enhancements

1. **User Management**: Integrate with actual user authentication
2. **File Uploads**: Support for image/file messages
3. **Push Notifications**: Offline message notifications  
4. **Message Encryption**: End-to-end encryption for sensitive chats
5. **Moderation**: Message filtering and moderation tools
6. **Analytics**: Chat usage analytics and reporting
