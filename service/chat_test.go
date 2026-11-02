package service

import (
	"context"
	"errors"
	"testing"
	"time"
	"what-to-eat/be/model"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ---------------------------------------------------------------------------
// Chat-specific mocks
// ---------------------------------------------------------------------------

type chatMockCollection struct {
	// InsertOne
	insertOneResult *mongo.InsertOneResult
	insertOneErr    error

	// FindOne
	findOneResult SingleResult

	// FindOneAndUpdate
	findOneAndUpdateResult SingleResult

	// Find
	findResult Cursor
	findErr    error

	// UpdateOne
	updateOneErr   error
	updateOneCalls int

	// Call tracking
	insertOneCalls int
}

func (m *chatMockCollection) InsertOne(ctx context.Context, doc interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	m.insertOneCalls++
	return m.insertOneResult, m.insertOneErr
}

func (m *chatMockCollection) FindOneAndUpdate(ctx context.Context, filter, update interface{}, opts ...*options.FindOneAndUpdateOptions) SingleResult {
	return m.findOneAndUpdateResult
}

func (m *chatMockCollection) CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	return 0, nil
}

func (m *chatMockCollection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (Cursor, error) {
	if m.findResult != nil {
		return m.findResult, m.findErr
	}
	return emptyCursor(), m.findErr
}

func (m *chatMockCollection) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) SingleResult {
	return m.findOneResult
}

func (m *chatMockCollection) UpdateOne(ctx context.Context, filter, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	m.updateOneCalls++
	if m.updateOneErr != nil {
		return nil, m.updateOneErr
	}
	return &mongo.UpdateResult{MatchedCount: 1, ModifiedCount: 1}, nil
}

func (m *chatMockCollection) UpdateMany(ctx context.Context, filter, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return nil, nil
}

func (m *chatMockCollection) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return &mongo.DeleteResult{DeletedCount: 1}, nil
}

func (m *chatMockCollection) Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (Cursor, error) {
	return emptyCursor(), nil
}

// ---------------------------------------------------------------------------
// SingleResult helpers
// ---------------------------------------------------------------------------

// chatMockSingleResult decodes a ChatMessage.
type chatMsgSingleResult struct {
	msg model.ChatMessage
	err error
}

func (r *chatMsgSingleResult) Decode(v interface{}) error {
	if r.err != nil {
		return r.err
	}
	*(v.(*model.ChatMessage)) = r.msg
	return nil
}
func (r *chatMsgSingleResult) Err() error { return r.err }

// chatRoomSingleResult decodes a ChatRoom.
type chatRoomSingleResult struct {
	room model.ChatRoom
	err  error
}

func (r *chatRoomSingleResult) Decode(v interface{}) error {
	if r.err != nil {
		return r.err
	}
	*(v.(*model.ChatRoom)) = r.room
	return nil
}
func (r *chatRoomSingleResult) Err() error { return r.err }

// ---------------------------------------------------------------------------
// Cursor helpers for messages
// ---------------------------------------------------------------------------

func msgsCursor(msgs []*model.ChatMessage) Cursor {
	return &genericCursor{
		decodeFn: func(results interface{}) error {
			*(results.(*[]*model.ChatMessage)) = msgs
			return nil
		},
	}
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func sampleMessage() model.ChatMessage {
	return model.ChatMessage{
		ID:         primitive.NewObjectID(),
		Content:    "hello",
		SenderID:   "user1",
		SenderName: "Alice",
		Type:       "text",
		RoomID:     "room1",
		Timestamp:  float64(time.Now().Unix()),
		Reactions:  map[string]int{},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Deleted:    false,
	}
}

func sampleRoom() model.ChatRoom {
	return model.ChatRoom{
		ID:           primitive.NewObjectID(),
		Name:         "generalroom1",
		Type:         "general",
		RoomID:       "room1",
		Participants: []string{"user1"},
		OnlineUsers:  []string{"user1"},
		TypingUsers:  []string{},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Deleted:      false,
	}
}

func newChatSvc(msgCol, roomCol CollectionInterface) *ChatService {
	return NewChatService(msgCol, roomCol)
}

// ---------------------------------------------------------------------------
// CreateMessage
// ---------------------------------------------------------------------------

func TestChatService_CreateMessage_Success(t *testing.T) {
	oid := primitive.NewObjectID()
	msgCol := &chatMockCollection{
		insertOneResult: &mongo.InsertOneResult{InsertedID: oid},
	}
	svc := newChatSvc(msgCol, nil)

	dto := model.CreateChatMessageDto{
		Content:    "hello",
		Type:       "text",
		RoomID:     "room1",
		SenderID:   "user1",
		SenderName: "Alice",
	}

	msg, err := svc.CreateMessage(dto)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.ID != oid {
		t.Errorf("expected ID %s, got %s", oid.Hex(), msg.ID.Hex())
	}
	if msg.Content != "hello" {
		t.Errorf("unexpected content: %s", msg.Content)
	}
	if msgCol.insertOneCalls != 1 {
		t.Errorf("expected 1 InsertOne call, got %d", msgCol.insertOneCalls)
	}
}

func TestChatService_CreateMessage_SetsDefaultReactions(t *testing.T) {
	oid := primitive.NewObjectID()
	msgCol := &chatMockCollection{
		insertOneResult: &mongo.InsertOneResult{InsertedID: oid},
	}
	svc := newChatSvc(msgCol, nil)

	dto := model.CreateChatMessageDto{Content: "hi", Type: "text", RoomID: "r1", SenderID: "u1", SenderName: "Bob"}
	msg, err := svc.CreateMessage(dto)
	if err != nil {
		t.Fatal(err)
	}
	if msg.Reactions == nil {
		t.Error("expected Reactions to be initialized, got nil")
	}
}

func TestChatService_CreateMessage_PreservesProvidedReactions(t *testing.T) {
	oid := primitive.NewObjectID()
	msgCol := &chatMockCollection{
		insertOneResult: &mongo.InsertOneResult{InsertedID: oid},
	}
	svc := newChatSvc(msgCol, nil)

	reactions := map[string]int{"👍": 3}
	dto := model.CreateChatMessageDto{
		Content:    "hi",
		Type:       "text",
		RoomID:     "r1",
		SenderID:   "u1",
		SenderName: "Bob",
		Reactions:  reactions,
	}
	msg, err := svc.CreateMessage(dto)
	if err != nil {
		t.Fatal(err)
	}
	if msg.Reactions["👍"] != 3 {
		t.Errorf("expected reaction count 3, got %d", msg.Reactions["👍"])
	}
}

func TestChatService_CreateMessage_InsertError(t *testing.T) {
	msgCol := &chatMockCollection{
		insertOneErr: errors.New("db error"),
	}
	svc := newChatSvc(msgCol, nil)

	dto := model.CreateChatMessageDto{Content: "hi", Type: "text", RoomID: "r1", SenderID: "u1", SenderName: "Bob"}
	_, err := svc.CreateMessage(dto)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestChatService_CreateMessage_SetsDeletedFalse(t *testing.T) {
	oid := primitive.NewObjectID()
	msgCol := &chatMockCollection{insertOneResult: &mongo.InsertOneResult{InsertedID: oid}}
	svc := newChatSvc(msgCol, nil)

	dto := model.CreateChatMessageDto{Content: "hi", Type: "text", RoomID: "r1", SenderID: "u1", SenderName: "Bob"}
	msg, _ := svc.CreateMessage(dto)
	if msg.Deleted {
		t.Error("expected Deleted=false on new message")
	}
}

// ---------------------------------------------------------------------------
// GetMessageHistory
// ---------------------------------------------------------------------------

func TestChatService_GetMessageHistory_ReturnsChronological(t *testing.T) {
	now := time.Now().Unix()
	msg1 := &model.ChatMessage{ID: primitive.NewObjectID(), Timestamp: float64(now - 10)}
	msg2 := &model.ChatMessage{ID: primitive.NewObjectID(), Timestamp: float64(now)}
	// cursor returns latest-first (as the real DB would with sort -1)
	msgCol := &chatMockCollection{
		findResult: msgsCursor([]*model.ChatMessage{msg2, msg1}),
	}
	svc := newChatSvc(msgCol, nil)

	msgs, err := svc.GetMessageHistory("room1", 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	// After reversal, oldest (msg1) should be first
	if msgs[0].Timestamp > msgs[1].Timestamp {
		t.Error("expected chronological (oldest first) order")
	}
}

func TestChatService_GetMessageHistory_EmptyRoom(t *testing.T) {
	msgCol := &chatMockCollection{findResult: msgsCursor([]*model.ChatMessage{})}
	svc := newChatSvc(msgCol, nil)

	msgs, err := svc.GetMessageHistory("empty-room", 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 0 {
		t.Errorf("expected 0 messages, got %d", len(msgs))
	}
}

func TestChatService_GetMessageHistory_FindError(t *testing.T) {
	msgCol := &chatMockCollection{findErr: errors.New("find failed")}
	svc := newChatSvc(msgCol, nil)

	_, err := svc.GetMessageHistory("room1", 10, 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestChatService_GetMessageHistory_CursorDecodeError(t *testing.T) {
	msgCol := &chatMockCollection{
		findResult: &genericCursor{err: errors.New("decode error")},
	}
	svc := newChatSvc(msgCol, nil)

	_, err := svc.GetMessageHistory("room1", 10, 0)
	if err == nil {
		t.Fatal("expected decode error, got nil")
	}
}

func TestChatService_GetMessageHistory_SingleMessage(t *testing.T) {
	msg := &model.ChatMessage{ID: primitive.NewObjectID(), Timestamp: 100}
	msgCol := &chatMockCollection{findResult: msgsCursor([]*model.ChatMessage{msg})}
	svc := newChatSvc(msgCol, nil)

	msgs, err := svc.GetMessageHistory("room1", 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
}

// ---------------------------------------------------------------------------
// UpdateMessageReactions
// ---------------------------------------------------------------------------

func TestChatService_UpdateMessageReactions_Success(t *testing.T) {
	msg := sampleMessage()
	msg.Reactions = map[string]int{"❤️": 2}
	msgCol := &chatMockCollection{
		findOneAndUpdateResult: &chatMsgSingleResult{msg: msg},
	}
	svc := newChatSvc(msgCol, nil)

	updated, err := svc.UpdateMessageReactions(msg.ID.Hex(), map[string]int{"❤️": 3})
	if err != nil {
		t.Fatal(err)
	}
	if updated == nil {
		t.Fatal("expected updated message, got nil")
	}
}

func TestChatService_UpdateMessageReactions_InvalidID(t *testing.T) {
	msgCol := &chatMockCollection{}
	svc := newChatSvc(msgCol, nil)

	_, err := svc.UpdateMessageReactions("not-an-objectid", map[string]int{})
	if err == nil {
		t.Fatal("expected error for invalid ID")
	}
}

func TestChatService_UpdateMessageReactions_DecodeError(t *testing.T) {
	oid := primitive.NewObjectID()
	msgCol := &chatMockCollection{
		findOneAndUpdateResult: &chatMsgSingleResult{err: errors.New("not found")},
	}
	svc := newChatSvc(msgCol, nil)

	_, err := svc.UpdateMessageReactions(oid.Hex(), map[string]int{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// FindMessageByID
// ---------------------------------------------------------------------------

func TestChatService_FindMessageByID_Success(t *testing.T) {
	msg := sampleMessage()
	msgCol := &chatMockCollection{
		findOneResult: &chatMsgSingleResult{msg: msg},
	}
	svc := newChatSvc(msgCol, nil)

	found, err := svc.FindMessageByID(msg.ID.Hex())
	if err != nil {
		t.Fatal(err)
	}
	if found.ID != msg.ID {
		t.Errorf("expected ID %s, got %s", msg.ID.Hex(), found.ID.Hex())
	}
}

func TestChatService_FindMessageByID_InvalidID(t *testing.T) {
	svc := newChatSvc(&chatMockCollection{}, nil)
	_, err := svc.FindMessageByID("bad-id")
	if err == nil {
		t.Fatal("expected error for invalid ID")
	}
}

func TestChatService_FindMessageByID_NotFound(t *testing.T) {
	oid := primitive.NewObjectID()
	msgCol := &chatMockCollection{
		findOneResult: &chatMsgSingleResult{err: errors.New("not found")},
	}
	svc := newChatSvc(msgCol, nil)

	_, err := svc.FindMessageByID(oid.Hex())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// CreateOrGetRoom — existing room
// ---------------------------------------------------------------------------

func TestChatService_CreateOrGetRoom_ExistingRoom(t *testing.T) {
	room := sampleRoom()
	roomCol := &chatMockCollection{
		findOneResult: &chatRoomSingleResult{room: room},
	}
	svc := newChatSvc(nil, roomCol)

	got, err := svc.CreateOrGetRoom("room1", "general")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != room.Name {
		t.Errorf("expected room name %s, got %s", room.Name, got.Name)
	}
	if roomCol.insertOneCalls != 0 {
		t.Error("InsertOne should not be called for an existing room")
	}
}

func TestChatService_CreateOrGetRoom_CreatesNew(t *testing.T) {
	oid := primitive.NewObjectID()
	roomCol := &chatMockCollection{
		// FindOne returns ErrNoDocuments → triggers creation
		findOneResult:   &chatRoomSingleResult{err: mongo.ErrNoDocuments},
		insertOneResult: &mongo.InsertOneResult{InsertedID: oid},
	}
	svc := newChatSvc(nil, roomCol)

	got, err := svc.CreateOrGetRoom("room2", "voteGame")
	if err != nil {
		t.Fatal(err)
	}
	if got.RoomID != "room2" {
		t.Errorf("expected roomID 'room2', got '%s'", got.RoomID)
	}
	if got.Type != "voteGame" {
		t.Errorf("expected type 'voteGame', got '%s'", got.Type)
	}
	if roomCol.insertOneCalls != 1 {
		t.Errorf("expected 1 InsertOne call, got %d", roomCol.insertOneCalls)
	}
}

func TestChatService_CreateOrGetRoom_InsertError(t *testing.T) {
	roomCol := &chatMockCollection{
		findOneResult: &chatRoomSingleResult{err: mongo.ErrNoDocuments},
		insertOneErr:  errors.New("insert failed"),
	}
	svc := newChatSvc(nil, roomCol)

	_, err := svc.CreateOrGetRoom("room3", "general")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestChatService_CreateOrGetRoom_FindError(t *testing.T) {
	roomCol := &chatMockCollection{
		findOneResult: &chatRoomSingleResult{err: errors.New("db unavailable")},
	}
	svc := newChatSvc(nil, roomCol)

	_, err := svc.CreateOrGetRoom("room4", "general")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// AddUserToRoom
// ---------------------------------------------------------------------------

func TestChatService_AddUserToRoom_Success(t *testing.T) {
	roomCol := &chatMockCollection{}
	svc := newChatSvc(nil, roomCol)

	err := svc.AddUserToRoom("generalroom1", "user1")
	if err != nil {
		t.Fatal(err)
	}
	if roomCol.updateOneCalls != 1 {
		t.Errorf("expected 1 UpdateOne call, got %d", roomCol.updateOneCalls)
	}
}

func TestChatService_AddUserToRoom_Error(t *testing.T) {
	roomCol := &chatMockCollection{updateOneErr: errors.New("update failed")}
	svc := newChatSvc(nil, roomCol)

	err := svc.AddUserToRoom("generalroom1", "user1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// RemoveUserFromRoom
// ---------------------------------------------------------------------------

func TestChatService_RemoveUserFromRoom_Success(t *testing.T) {
	roomCol := &chatMockCollection{}
	svc := newChatSvc(nil, roomCol)

	err := svc.RemoveUserFromRoom("generalroom1", "user1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestChatService_RemoveUserFromRoom_Error(t *testing.T) {
	roomCol := &chatMockCollection{updateOneErr: errors.New("db error")}
	svc := newChatSvc(nil, roomCol)

	err := svc.RemoveUserFromRoom("generalroom1", "user1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// UpdateTypingUsers
// ---------------------------------------------------------------------------

func TestChatService_UpdateTypingUsers_SetTyping(t *testing.T) {
	roomCol := &chatMockCollection{}
	svc := newChatSvc(nil, roomCol)

	err := svc.UpdateTypingUsers("generalroom1", "user1", true)
	if err != nil {
		t.Fatal(err)
	}
	if roomCol.updateOneCalls != 1 {
		t.Errorf("expected 1 UpdateOne, got %d", roomCol.updateOneCalls)
	}
}

func TestChatService_UpdateTypingUsers_ClearTyping(t *testing.T) {
	roomCol := &chatMockCollection{}
	svc := newChatSvc(nil, roomCol)

	err := svc.UpdateTypingUsers("generalroom1", "user1", false)
	if err != nil {
		t.Fatal(err)
	}
}

func TestChatService_UpdateTypingUsers_Error(t *testing.T) {
	roomCol := &chatMockCollection{updateOneErr: errors.New("db error")}
	svc := newChatSvc(nil, roomCol)

	err := svc.UpdateTypingUsers("generalroom1", "user1", true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// GetOnlineUsers
// ---------------------------------------------------------------------------

func TestChatService_GetOnlineUsers_Success(t *testing.T) {
	room := sampleRoom()
	room.OnlineUsers = []string{"user1", "user2"}
	roomCol := &chatMockCollection{
		findOneResult: &chatRoomSingleResult{room: room},
	}
	svc := newChatSvc(nil, roomCol)

	users, err := svc.GetOnlineUsers("generalroom1")
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 2 {
		t.Errorf("expected 2 online users, got %d", len(users))
	}
}

func TestChatService_GetOnlineUsers_RoomNotFound(t *testing.T) {
	roomCol := &chatMockCollection{
		findOneResult: &chatRoomSingleResult{err: errors.New("not found")},
	}
	svc := newChatSvc(nil, roomCol)

	_, err := svc.GetOnlineUsers("generalroom1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestChatService_GetOnlineUsers_Empty(t *testing.T) {
	room := sampleRoom()
	room.OnlineUsers = []string{}
	roomCol := &chatMockCollection{findOneResult: &chatRoomSingleResult{room: room}}
	svc := newChatSvc(nil, roomCol)

	users, err := svc.GetOnlineUsers("generalroom1")
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 0 {
		t.Errorf("expected 0 users, got %d", len(users))
	}
}

// ---------------------------------------------------------------------------
// GetTypingUsers
// ---------------------------------------------------------------------------

func TestChatService_GetTypingUsers_Success(t *testing.T) {
	room := sampleRoom()
	room.TypingUsers = []string{"user1"}
	roomCol := &chatMockCollection{
		findOneResult: &chatRoomSingleResult{room: room},
	}
	svc := newChatSvc(nil, roomCol)

	users, err := svc.GetTypingUsers("generalroom1")
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 1 {
		t.Errorf("expected 1 typing user, got %d", len(users))
	}
}

func TestChatService_GetTypingUsers_RoomNotFound(t *testing.T) {
	roomCol := &chatMockCollection{
		findOneResult: &chatRoomSingleResult{err: errors.New("not found")},
	}
	svc := newChatSvc(nil, roomCol)

	_, err := svc.GetTypingUsers("generalroom1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestChatService_GetTypingUsers_Empty(t *testing.T) {
	room := sampleRoom()
	room.TypingUsers = []string{}
	roomCol := &chatMockCollection{findOneResult: &chatRoomSingleResult{room: room}}
	svc := newChatSvc(nil, roomCol)

	users, err := svc.GetTypingUsers("generalroom1")
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 0 {
		t.Errorf("expected 0 typing users, got %d", len(users))
	}
}
