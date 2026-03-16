package service

import (
	"context"
	"errors"
	"testing"
	"what-to-eat/be/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

// mockSingleResult satisfies SingleResult.
type mockSingleResult struct {
	err      error
	decodeFn func(v interface{}) error
}

func (m *mockSingleResult) Decode(v interface{}) error {
	if m.decodeFn != nil {
		return m.decodeFn(v)
	}
	return nil
}

func (m *mockSingleResult) Err() error { return m.err }

// mockCursor satisfies Cursor.
type mockCursor struct {
	contacts []*model.Contact
	err      error
}

func (m *mockCursor) All(ctx context.Context, results interface{}) error {
	if m.err != nil {
		return m.err
	}
	target := results.(*[]*model.Contact)
	*target = m.contacts
	return nil
}

func (m *mockCursor) Close(ctx context.Context) error { return nil }

// mockCollection satisfies CollectionInterface.
type mockCollection struct {
	// InsertOne
	insertOneResult *mongo.InsertOneResult
	insertOneErr    error

	// FindOneAndUpdate
	findOneAndUpdateResult SingleResult

	// CountDocuments
	countResult     int64
	countErr        error
	lastCountFilter interface{}

	// Find
	findResult     Cursor
	findErr        error
	lastFindFilter interface{}

	// FindOne
	findOneResult SingleResult
}

func (m *mockCollection) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	return m.insertOneResult, m.insertOneErr
}

func (m *mockCollection) FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...*options.FindOneAndUpdateOptions) SingleResult {
	return m.findOneAndUpdateResult
}

func (m *mockCollection) CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	m.lastCountFilter = filter
	return m.countResult, m.countErr
}

func (m *mockCollection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (Cursor, error) {
	m.lastFindFilter = filter
	return m.findResult, m.findErr
}

func (m *mockCollection) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) SingleResult {
	return m.findOneResult
}

// ---------------------------------------------------------------------------
// Create tests
// ---------------------------------------------------------------------------

func TestCreate_Success(t *testing.T) {
	oid := primitive.NewObjectID()
	col := &mockCollection{
		insertOneResult: &mongo.InsertOneResult{InsertedID: oid},
	}
	svc := NewContactService(col)

	contact, err := svc.Create(model.CreateContactDto{Email: "a@b.com", Name: "Alice", Message: "Hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contact.ID != oid.Hex() {
		t.Errorf("expected ID %s, got %s", oid.Hex(), contact.ID)
	}
	if contact.Email != "a@b.com" {
		t.Errorf("expected email a@b.com, got %s", contact.Email)
	}
	if contact.Name != "Alice" {
		t.Errorf("expected name Alice, got %s", contact.Name)
	}
	if contact.Deleted {
		t.Error("expected Deleted to be false")
	}
	if contact.CreatedAt == nil {
		t.Error("expected CreatedAt to be set")
	}
}

func TestCreate_DBError(t *testing.T) {
	col := &mockCollection{insertOneErr: errors.New("db error")}
	svc := NewContactService(col)

	_, err := svc.Create(model.CreateContactDto{Email: "a@b.com"})
	if err == nil || err.Error() != "db error" {
		t.Errorf("expected db error, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Update tests
// ---------------------------------------------------------------------------

func TestUpdate_InvalidID(t *testing.T) {
	col := &mockCollection{}
	svc := NewContactService(col)

	_, err := svc.Update(model.UpdateContactDto{ID: "not-a-hex"}, &model.JwtCustomClaims{})
	if err == nil || err.Error() != "invalid ID format" {
		t.Errorf("expected 'invalid ID format', got %v", err)
	}
}

func TestUpdate_DBError(t *testing.T) {
	oid := primitive.NewObjectID()
	col := &mockCollection{
		findOneAndUpdateResult: &mockSingleResult{err: errors.New("update failed")},
	}
	svc := NewContactService(col)

	_, err := svc.Update(model.UpdateContactDto{ID: oid.Hex(), Email: "a@b.com"}, &model.JwtCustomClaims{})
	if err == nil || err.Error() != "update failed" {
		t.Errorf("expected 'update failed', got %v", err)
	}
}

func TestUpdate_Success(t *testing.T) {
	oid := primitive.NewObjectID()
	expected := model.Contact{Email: "updated@b.com", Name: "Updated"}
	col := &mockCollection{
		findOneAndUpdateResult: &mockSingleResult{
			decodeFn: func(v interface{}) error {
				*(v.(*model.Contact)) = expected
				return nil
			},
		},
	}
	svc := NewContactService(col)

	contact, err := svc.Update(
		model.UpdateContactDto{ID: oid.Hex(), Email: "updated@b.com", Name: "Updated"},
		&model.JwtCustomClaims{},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contact.Email != "updated@b.com" {
		t.Errorf("expected email updated@b.com, got %s", contact.Email)
	}
}

// ---------------------------------------------------------------------------
// Remove tests
// ---------------------------------------------------------------------------

func TestRemove_InvalidID(t *testing.T) {
	col := &mockCollection{}
	svc := NewContactService(col)

	_, err := svc.Remove("not-a-hex", &model.JwtCustomClaims{})
	if err == nil || err.Error() != "invalid ID format" {
		t.Errorf("expected 'invalid ID format', got %v", err)
	}
}

func TestRemove_DBError(t *testing.T) {
	oid := primitive.NewObjectID()
	col := &mockCollection{
		findOneAndUpdateResult: &mockSingleResult{err: errors.New("remove failed")},
	}
	svc := NewContactService(col)

	_, err := svc.Remove(oid.Hex(), &model.JwtCustomClaims{})
	if err == nil || err.Error() != "remove failed" {
		t.Errorf("expected 'remove failed', got %v", err)
	}
}

func TestRemove_Success(t *testing.T) {
	oid := primitive.NewObjectID()
	expected := model.Contact{Email: "x@b.com", Deleted: true}
	col := &mockCollection{
		findOneAndUpdateResult: &mockSingleResult{
			decodeFn: func(v interface{}) error {
				*(v.(*model.Contact)) = expected
				return nil
			},
		},
	}
	svc := NewContactService(col)

	contact, err := svc.Remove(oid.Hex(), &model.JwtCustomClaims{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contact.Deleted {
		t.Error("expected contact to be marked deleted")
	}
}

// ---------------------------------------------------------------------------
// Find tests
// ---------------------------------------------------------------------------

func TestFind_NoKeyword(t *testing.T) {
	contacts := []*model.Contact{{Email: "a@b.com"}, {Email: "c@d.com"}}
	col := &mockCollection{
		countResult: 2,
		findResult:  &mockCursor{contacts: contacts},
	}
	svc := NewContactService(col)

	result, count, err := svc.Find(model.QueryContactDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 contacts, got %d", len(result))
	}
}

func TestFind_WithKeyword_FilterContainsText(t *testing.T) {
	kw := "pizza"
	col := &mockCollection{
		countResult: 1,
		findResult:  &mockCursor{contacts: []*model.Contact{{Email: "a@b.com"}}},
	}
	svc := NewContactService(col)

	_, _, err := svc.Find(model.QueryContactDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}, Keyword: &kw})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify that $text was appended to the filter passed to CountDocuments.
	filter, ok := col.lastCountFilter.(bson.D)
	if !ok {
		t.Fatal("expected bson.D filter")
	}
	found := false
	for _, elem := range filter {
		if elem.Key == "$text" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected $text element in filter when keyword is provided")
	}
}

func TestFind_CountError(t *testing.T) {
	col := &mockCollection{countErr: errors.New("count failed")}
	svc := NewContactService(col)

	_, _, err := svc.Find(model.QueryContactDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}})
	if err == nil || err.Error() != "count failed" {
		t.Errorf("expected 'count failed', got %v", err)
	}
}

// ---------------------------------------------------------------------------
// FindOne tests
// ---------------------------------------------------------------------------

func TestFindOne_InvalidID(t *testing.T) {
	col := &mockCollection{}
	svc := NewContactService(col)

	_, err := svc.FindOne("not-a-hex")
	if err == nil || err.Error() != "invalid ID format" {
		t.Errorf("expected 'invalid ID format', got %v", err)
	}
}

func TestFindOne_Success(t *testing.T) {
	oid := primitive.NewObjectID()
	expected := model.Contact{Email: "found@b.com"}
	col := &mockCollection{
		findOneResult: &mockSingleResult{
			decodeFn: func(v interface{}) error {
				*(v.(*model.Contact)) = expected
				return nil
			},
		},
	}
	svc := NewContactService(col)

	contact, err := svc.FindOne(oid.Hex())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contact.Email != "found@b.com" {
		t.Errorf("expected email found@b.com, got %s", contact.Email)
	}
}

func TestFindOne_NotFound(t *testing.T) {
	oid := primitive.NewObjectID()
	col := &mockCollection{
		findOneResult: &mockSingleResult{err: mongo.ErrNoDocuments},
	}
	svc := NewContactService(col)

	_, err := svc.FindOne(oid.Hex())
	if err != mongo.ErrNoDocuments {
		t.Errorf("expected mongo.ErrNoDocuments, got %v", err)
	}
}
