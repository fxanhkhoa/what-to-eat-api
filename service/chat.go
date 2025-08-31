package service

import (
	"context"
	"errors"
	"time"
	"what-to-eat/be/config"
	"what-to-eat/be/constants"
	"what-to-eat/be/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ChatService struct{}

func (cs *ChatService) Collection() *mongo.Collection {
	dbName := config.GetDBInstance().GetDbName()
	col := config.GetDBInstance().GetClient().Database(dbName).Collection(constants.CHAT_MESSAGE_COLLECTION)
	return col
}

func (cs *ChatService) RoomCollection() *mongo.Collection {
	dbName := config.GetDBInstance().GetDbName()
	col := config.GetDBInstance().GetClient().Database(dbName).Collection(constants.CHAT_ROOM_COLLECTION)
	return col
}

// Create a new chat message
func (cs *ChatService) CreateMessage(createMessageInput model.CreateChatMessageDto) (*model.ChatMessage, error) {
	collection := cs.Collection()
	now := time.Now()
	timestamp := float64(now.Unix())

	message := model.ChatMessage{
		Content:      createMessageInput.Content,
		SenderID:     createMessageInput.SenderID,
		SenderName:   createMessageInput.SenderName,
		SenderAvatar: createMessageInput.SenderAvatar,
		Type:         createMessageInput.Type,
		Timestamp:    timestamp,
		Reactions:    createMessageInput.Reactions,
		RoomID:       createMessageInput.RoomID,
		CreatedAt:    now,
		UpdatedAt:    now,
		Deleted:      false,
	}

	if message.Reactions == nil {
		message.Reactions = make(map[string]int)
	}

	result, err := collection.InsertOne(context.TODO(), message)
	if err != nil {
		return nil, err
	}

	message.ID = result.InsertedID.(primitive.ObjectID)
	return &message, nil
}

// Get message history for a room with pagination
func (cs *ChatService) GetMessageHistory(roomID string, limit int, before float64) ([]*model.ChatMessage, error) {
	collection := cs.Collection()

	filter := bson.M{
		"roomId":  roomID,
		"deleted": false,
	}

	if before > 0 {
		filter["timestamp"] = bson.M{"$lt": before}
	}

	opts := options.Find()
	opts.SetSort(bson.D{{Key: "timestamp", Value: -1}}) // Latest first
	opts.SetLimit(int64(limit))

	cursor, err := collection.Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var messages []*model.ChatMessage
	if err = cursor.All(context.TODO(), &messages); err != nil {
		return nil, err
	}

	// Reverse the slice to get chronological order (oldest first)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// Update message reactions
func (cs *ChatService) UpdateMessageReactions(messageID string, reactions map[string]int) (*model.ChatMessage, error) {
	collection := cs.Collection()

	objectID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return nil, errors.New("invalid message ID format")
	}

	filter := bson.M{"_id": objectID, "deleted": false}
	update := bson.M{
		"$set": bson.M{
			"reactions": reactions,
			"updatedAt": time.Now(),
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedMessage model.ChatMessage
	err = collection.FindOneAndUpdate(context.TODO(), filter, update, opts).Decode(&updatedMessage)
	if err != nil {
		return nil, err
	}

	return &updatedMessage, nil
}

// Find a message by ID
func (cs *ChatService) FindMessageByID(messageID string) (*model.ChatMessage, error) {
	collection := cs.Collection()

	objectID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return nil, errors.New("invalid message ID format")
	}

	filter := bson.M{"_id": objectID, "deleted": false}
	var message model.ChatMessage
	err = collection.FindOne(context.TODO(), filter).Decode(&message)
	if err != nil {
		return nil, err
	}

	return &message, nil
}

// Chat Room Management

// Create or get a chat room
func (cs *ChatService) CreateOrGetRoom(roomID, roomType string) (*model.ChatRoom, error) {
	collection := cs.RoomCollection()
	now := time.Now()

	roomName := roomType + roomID
	filter := bson.M{"name": roomName, "deleted": false}

	var room model.ChatRoom
	err := collection.FindOne(context.TODO(), filter).Decode(&room)

	if err == mongo.ErrNoDocuments {
		// Create new room
		room = model.ChatRoom{
			ID:           primitive.NewObjectID(),
			Name:         roomName,
			Type:         roomType,
			RoomID:       roomID,
			Participants: []string{},
			OnlineUsers:  []string{},
			TypingUsers:  []string{},
			CreatedAt:    now,
			UpdatedAt:    now,
			Deleted:      false,
		}

		_, err := collection.InsertOne(context.TODO(), room)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return &room, nil
}

// Add user to room
func (cs *ChatService) AddUserToRoom(roomName, userID string) error {
	collection := cs.RoomCollection()

	filter := bson.M{"name": roomName, "deleted": false}
	update := bson.M{
		"$addToSet": bson.M{
			"participants": userID,
			"onlineUsers":  userID,
		},
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}

	_, err := collection.UpdateOne(context.TODO(), filter, update)
	return err
}

// Remove user from room
func (cs *ChatService) RemoveUserFromRoom(roomName, userID string) error {
	collection := cs.RoomCollection()

	filter := bson.M{"name": roomName, "deleted": false}
	update := bson.M{
		"$pull": bson.M{
			"onlineUsers": userID,
			"typingUsers": userID,
		},
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}

	_, err := collection.UpdateOne(context.TODO(), filter, update)
	return err
}

// Update typing users
func (cs *ChatService) UpdateTypingUsers(roomName, userID string, isTyping bool) error {
	collection := cs.RoomCollection()

	filter := bson.M{"name": roomName, "deleted": false}
	var update bson.M

	if isTyping {
		update = bson.M{
			"$addToSet": bson.M{"typingUsers": userID},
			"$set":      bson.M{"updatedAt": time.Now()},
		}
	} else {
		update = bson.M{
			"$pull": bson.M{"typingUsers": userID},
			"$set":  bson.M{"updatedAt": time.Now()},
		}
	}

	_, err := collection.UpdateOne(context.TODO(), filter, update)
	return err
}

// Get online users for a room
func (cs *ChatService) GetOnlineUsers(roomName string) ([]string, error) {
	collection := cs.RoomCollection()

	filter := bson.M{"name": roomName, "deleted": false}
	var room model.ChatRoom
	err := collection.FindOne(context.TODO(), filter).Decode(&room)
	if err != nil {
		return nil, err
	}

	return room.OnlineUsers, nil
}

// Get typing users for a room
func (cs *ChatService) GetTypingUsers(roomName string) ([]string, error) {
	collection := cs.RoomCollection()

	filter := bson.M{"name": roomName, "deleted": false}
	var room model.ChatRoom
	err := collection.FindOne(context.TODO(), filter).Decode(&room)
	if err != nil {
		return nil, err
	}

	return room.TypingUsers, nil
}
