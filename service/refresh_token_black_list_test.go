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
// Mock collection
// ---------------------------------------------------------------------------

type rtblMockCollection struct {
	insertOneResult *mongo.InsertOneResult
	insertOneErr    error

	findOneResult SingleResult

	updateOneResult *mongo.UpdateResult
	updateOneErr    error

	deleteOneResult *mongo.DeleteResult
	deleteOneErr    error
}

func (m *rtblMockCollection) InsertOne(_ context.Context, _ interface{}, _ ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	return m.insertOneResult, m.insertOneErr
}

func (m *rtblMockCollection) FindOne(_ context.Context, _ interface{}, _ ...*options.FindOneOptions) SingleResult {
	if m.findOneResult != nil {
		return m.findOneResult
	}
	return &errSingleResult{err: errors.New("not found")}
}

func (m *rtblMockCollection) FindOneAndUpdate(_ context.Context, _, _ interface{}, _ ...*options.FindOneAndUpdateOptions) SingleResult {
	return &errSingleResult{err: errors.New("not implemented")}
}

func (m *rtblMockCollection) CountDocuments(_ context.Context, _ interface{}, _ ...*options.CountOptions) (int64, error) {
	return 0, nil
}

func (m *rtblMockCollection) Find(_ context.Context, _ interface{}, _ ...*options.FindOptions) (Cursor, error) {
	return emptyCursor(), nil
}

func (m *rtblMockCollection) UpdateOne(_ context.Context, _, _ interface{}, _ ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return m.updateOneResult, m.updateOneErr
}

func (m *rtblMockCollection) UpdateMany(_ context.Context, _, _ interface{}, _ ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return nil, nil
}

func (m *rtblMockCollection) DeleteOne(_ context.Context, _ interface{}, _ ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return m.deleteOneResult, m.deleteOneErr
}

func (m *rtblMockCollection) Aggregate(_ context.Context, _ interface{}, _ ...*options.AggregateOptions) (Cursor, error) {
	return emptyCursor(), nil
}

// ---------------------------------------------------------------------------
// SingleResult helper
// ---------------------------------------------------------------------------

type rtblSingleResult struct {
	entry model.RefreshTokenBlackList
	err   error
}

func (r *rtblSingleResult) Decode(v interface{}) error {
	if r.err != nil {
		return r.err
	}
	*(v.(*model.RefreshTokenBlackList)) = r.entry
	return nil
}
func (r *rtblSingleResult) Err() error { return r.err }

// rtblCaptureCollection overrides InsertOne to record the inserted document.
type rtblCaptureCollection struct {
	rtblMockCollection
	inserted interface{}
}

func (c *rtblCaptureCollection) InsertOne(_ context.Context, doc interface{}, _ ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	c.inserted = doc
	return c.rtblMockCollection.insertOneResult, c.rtblMockCollection.insertOneErr
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newRTBLSvc(col CollectionInterface) *RefreshTokenBlackListService {
	return NewRefreshTokenBlackListService(col)
}

func sampleRTBLEntry() model.RefreshTokenBlackList {
	return model.RefreshTokenBlackList{
		ID:        primitive.NewObjectID(),
		Token:     "sample.jwt.token",
		UserID:    primitive.NewObjectID(),
		ExpiredAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestRefreshTokenBlackList_Create_Success(t *testing.T) {
	oid := primitive.NewObjectID()
	col := &rtblMockCollection{
		insertOneResult: &mongo.InsertOneResult{InsertedID: oid},
	}
	svc := newRTBLSvc(col)

	input := model.RefreshTokenBlackList{Token: "tok", UserID: primitive.NewObjectID()}
	id, err := svc.Create(input)
	if err != nil {
		t.Fatal(err)
	}
	if id.IsZero() {
		t.Error("expected non-zero ObjectID to be returned")
	}
}

func TestRefreshTokenBlackList_Create_SetsID(t *testing.T) {
	col := &rtblMockCollection{
		insertOneResult: &mongo.InsertOneResult{InsertedID: primitive.NewObjectID()},
	}
	svc := newRTBLSvc(col)

	input := model.RefreshTokenBlackList{Token: "tok"}
	id, err := svc.Create(input)
	if err != nil {
		t.Fatal(err)
	}
	if id.IsZero() {
		t.Error("expected service to assign a non-zero ID")
	}
}

func TestRefreshTokenBlackList_Create_SetsTimestamp(t *testing.T) {
	before := time.Now().Add(-time.Millisecond)
	cc := &rtblCaptureCollection{
		rtblMockCollection: rtblMockCollection{
			insertOneResult: &mongo.InsertOneResult{InsertedID: primitive.NewObjectID()},
		},
	}
	svc := newRTBLSvc(cc)

	_, err := svc.Create(model.RefreshTokenBlackList{Token: "tok"})
	if err != nil {
		t.Fatal(err)
	}

	tok, ok := cc.inserted.(model.RefreshTokenBlackList)
	if !ok {
		t.Fatal("inserted document is not a RefreshTokenBlackList")
	}
	after := time.Now().Add(time.Millisecond)
	if tok.CreatedAt.Before(before) || tok.CreatedAt.After(after) {
		t.Errorf("expected CreatedAt between %v and %v, got %v", before, after, tok.CreatedAt)
	}
}

func TestRefreshTokenBlackList_Create_InsertError(t *testing.T) {
	col := &rtblMockCollection{
		insertOneErr: errors.New("db error"),
	}
	svc := newRTBLSvc(col)

	_, err := svc.Create(model.RefreshTokenBlackList{Token: "tok"})
	if err == nil {
		t.Error("expected error from InsertOne, got nil")
	}
}

// ---------------------------------------------------------------------------
// GetByID
// ---------------------------------------------------------------------------

func TestRefreshTokenBlackList_GetByID_Success(t *testing.T) {
	entry := sampleRTBLEntry()
	col := &rtblMockCollection{
		findOneResult: &rtblSingleResult{entry: entry},
	}
	svc := newRTBLSvc(col)

	result, err := svc.GetByID(entry.ID)
	if err != nil {
		t.Fatal(err)
	}
	if result.Token != entry.Token {
		t.Errorf("expected token %q, got %q", entry.Token, result.Token)
	}
}

func TestRefreshTokenBlackList_GetByID_NotFound(t *testing.T) {
	col := &rtblMockCollection{
		findOneResult: &errSingleResult{err: errors.New("mongo: no documents in result")},
	}
	svc := newRTBLSvc(col)

	result, err := svc.GetByID(primitive.NewObjectID())
	if err == nil {
		t.Error("expected error, got nil")
	}
	if result != nil {
		t.Error("expected nil result on error")
	}
}

func TestRefreshTokenBlackList_GetByID_DecodeError(t *testing.T) {
	col := &rtblMockCollection{
		findOneResult: &errSingleResult{err: errors.New("decode error")},
	}
	svc := newRTBLSvc(col)

	_, err := svc.GetByID(primitive.NewObjectID())
	if err == nil {
		t.Error("expected decode error, got nil")
	}
}

// ---------------------------------------------------------------------------
// GetByToken
// ---------------------------------------------------------------------------

func TestRefreshTokenBlackList_GetByToken_Success(t *testing.T) {
	entry := sampleRTBLEntry()
	col := &rtblMockCollection{
		findOneResult: &rtblSingleResult{entry: entry},
	}
	svc := newRTBLSvc(col)

	result, err := svc.GetByToken(entry.Token)
	if err != nil {
		t.Fatal(err)
	}
	if result.Token != entry.Token {
		t.Errorf("expected token %q, got %q", entry.Token, result.Token)
	}
}

func TestRefreshTokenBlackList_GetByToken_NotFound(t *testing.T) {
	col := &rtblMockCollection{
		findOneResult: &errSingleResult{err: errors.New("not found")},
	}
	svc := newRTBLSvc(col)

	result, err := svc.GetByToken("unknown.token")
	if err == nil {
		t.Error("expected error, got nil")
	}
	if result != nil {
		t.Error("expected nil result on not-found")
	}
}

func TestRefreshTokenBlackList_GetByToken_DecodeError(t *testing.T) {
	col := &rtblMockCollection{
		findOneResult: &rtblSingleResult{err: errors.New("decode failure")},
	}
	svc := newRTBLSvc(col)

	_, err := svc.GetByToken("any.token")
	if err == nil {
		t.Error("expected error from decode, got nil")
	}
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestRefreshTokenBlackList_Update_Success(t *testing.T) {
	col := &rtblMockCollection{
		updateOneResult: &mongo.UpdateResult{MatchedCount: 1, ModifiedCount: 1},
	}
	svc := newRTBLSvc(col)

	err := svc.Update(primitive.NewObjectID(), map[string]interface{}{"token": "new"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRefreshTokenBlackList_Update_NotFound(t *testing.T) {
	col := &rtblMockCollection{
		updateOneResult: &mongo.UpdateResult{MatchedCount: 0},
	}
	svc := newRTBLSvc(col)

	err := svc.Update(primitive.NewObjectID(), map[string]interface{}{"token": "x"})
	if err == nil || err.Error() != "not found" {
		t.Errorf("expected 'not found' error, got %v", err)
	}
}

func TestRefreshTokenBlackList_Update_DBError(t *testing.T) {
	col := &rtblMockCollection{
		updateOneErr: errors.New("db write error"),
	}
	svc := newRTBLSvc(col)

	err := svc.Update(primitive.NewObjectID(), map[string]interface{}{"token": "x"})
	if err == nil {
		t.Error("expected DB error, got nil")
	}
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestRefreshTokenBlackList_Delete_Success(t *testing.T) {
	col := &rtblMockCollection{
		deleteOneResult: &mongo.DeleteResult{DeletedCount: 1},
	}
	svc := newRTBLSvc(col)

	err := svc.Delete(primitive.NewObjectID())
	if err != nil {
		t.Fatal(err)
	}
}

func TestRefreshTokenBlackList_Delete_NotFound(t *testing.T) {
	col := &rtblMockCollection{
		deleteOneResult: &mongo.DeleteResult{DeletedCount: 0},
	}
	svc := newRTBLSvc(col)

	err := svc.Delete(primitive.NewObjectID())
	if err == nil || err.Error() != "not found" {
		t.Errorf("expected 'not found' error, got %v", err)
	}
}

func TestRefreshTokenBlackList_Delete_DBError(t *testing.T) {
	col := &rtblMockCollection{
		deleteOneErr: errors.New("db delete error"),
	}
	svc := newRTBLSvc(col)

	err := svc.Delete(primitive.NewObjectID())
	if err == nil {
		t.Error("expected DB error, got nil")
	}
}
