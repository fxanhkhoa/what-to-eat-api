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
// Feedback-specific mock collection
// ---------------------------------------------------------------------------

type feedbackMockCollection struct {
	insertOneResult *mongo.InsertOneResult
	insertOneErr    error

	// Sequence of FindOne results consumed in order (for Update/Delete which
	// call FindOne first, then FindOneAndUpdate / DeleteOne).
	findOneQueue []SingleResult
	findOneIdx   int

	findOneAndUpdateResult SingleResult

	countResult int64
	countErr    error

	findResult Cursor
	findErr    error

	deleteOneResult *mongo.DeleteResult
	deleteOneErr    error

	insertOneCalls int
	deleteOneCalls int
}

func (m *feedbackMockCollection) InsertOne(ctx context.Context, doc interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	m.insertOneCalls++
	return m.insertOneResult, m.insertOneErr
}

func (m *feedbackMockCollection) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) SingleResult {
	if len(m.findOneQueue) > 0 {
		idx := m.findOneIdx
		if idx >= len(m.findOneQueue) {
			idx = len(m.findOneQueue) - 1
		}
		m.findOneIdx++
		return m.findOneQueue[idx]
	}
	return &errSingleResult{err: errors.New("no result configured")}
}

func (m *feedbackMockCollection) FindOneAndUpdate(ctx context.Context, filter, update interface{}, opts ...*options.FindOneAndUpdateOptions) SingleResult {
	return m.findOneAndUpdateResult
}

func (m *feedbackMockCollection) CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	return m.countResult, m.countErr
}

func (m *feedbackMockCollection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (Cursor, error) {
	if m.findResult != nil {
		return m.findResult, m.findErr
	}
	return emptyCursor(), m.findErr
}

func (m *feedbackMockCollection) UpdateOne(ctx context.Context, filter, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return &mongo.UpdateResult{ModifiedCount: 1}, nil
}

func (m *feedbackMockCollection) UpdateMany(ctx context.Context, filter, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return nil, nil
}

func (m *feedbackMockCollection) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	m.deleteOneCalls++
	if m.deleteOneErr != nil {
		return nil, m.deleteOneErr
	}
	if m.deleteOneResult != nil {
		return m.deleteOneResult, nil
	}
	return &mongo.DeleteResult{DeletedCount: 1}, nil
}

func (m *feedbackMockCollection) Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (Cursor, error) {
	return emptyCursor(), nil
}

// ---------------------------------------------------------------------------
// SingleResult helpers
// ---------------------------------------------------------------------------

// feedbackSingleResult decodes a Feedback.
type feedbackSingleResult struct {
	fb  model.Feedback
	err error
}

func (r *feedbackSingleResult) Decode(v interface{}) error {
	if r.err != nil {
		return r.err
	}
	*(v.(*model.Feedback)) = r.fb
	return nil
}
func (r *feedbackSingleResult) Err() error { return r.err }

// errSingleResult always errors on Decode/Err.
type errSingleResult struct{ err error }

func (r *errSingleResult) Decode(v interface{}) error { return r.err }
func (r *errSingleResult) Err() error                 { return r.err }

// ---------------------------------------------------------------------------
// Cursor helper
// ---------------------------------------------------------------------------

func feedbacksCursor(fbs []model.Feedback) Cursor {
	return &genericCursor{
		decodeFn: func(results interface{}) error {
			*(results.(*[]model.Feedback)) = fbs
			return nil
		},
	}
}

// ---------------------------------------------------------------------------
// Fixtures
// ---------------------------------------------------------------------------

func newUserID() primitive.ObjectID { return primitive.NewObjectID() }

func sampleFeedback(userID primitive.ObjectID) model.Feedback {
	now := time.Now()
	comment := "Great app!"
	_ = comment
	return model.Feedback{
		ID:        primitive.NewObjectID(),
		UserID:    &userID,
		UserName:  "Alice",
		Email:     "alice@example.com",
		Rating:    5,
		Comment:   "Great app!",
		Page:      "/home",
		UserAgent: "Mozilla/5.0",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func feedbackCreateDto() model.CreateFeedbackDto {
	return model.CreateFeedbackDto{
		UserName:  "Alice",
		Email:     "alice@example.com",
		Rating:    5,
		Comment:   "Great app!",
		Page:      "/home",
		UserAgent: "Mozilla/5.0",
	}
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestFeedbackService_Create_Success(t *testing.T) {
	oid := primitive.NewObjectID()
	userID := newUserID()
	col := &feedbackMockCollection{
		insertOneResult: &mongo.InsertOneResult{InsertedID: oid},
	}
	svc := NewFeedbackService(col)

	fb, err := svc.Create(feedbackCreateDto(), &userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.ID != oid {
		t.Errorf("expected ID %s, got %s", oid.Hex(), fb.ID.Hex())
	}
	if fb.Email != "alice@example.com" {
		t.Errorf("unexpected email: %s", fb.Email)
	}
	if col.insertOneCalls != 1 {
		t.Errorf("expected 1 InsertOne call, got %d", col.insertOneCalls)
	}
}

func TestFeedbackService_Create_NilUserID(t *testing.T) {
	oid := primitive.NewObjectID()
	col := &feedbackMockCollection{
		insertOneResult: &mongo.InsertOneResult{InsertedID: oid},
	}
	svc := NewFeedbackService(col)

	fb, err := svc.Create(feedbackCreateDto(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if fb.UserID != nil {
		t.Error("expected UserID to be nil")
	}
}

func TestFeedbackService_Create_SetsTimestamps(t *testing.T) {
	oid := primitive.NewObjectID()
	before := time.Now().Add(-time.Second)
	col := &feedbackMockCollection{insertOneResult: &mongo.InsertOneResult{InsertedID: oid}}
	svc := NewFeedbackService(col)

	fb, err := svc.Create(feedbackCreateDto(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if fb.CreatedAt.Before(before) {
		t.Error("CreatedAt should be set to approximately now")
	}
	if fb.UpdatedAt.Before(before) {
		t.Error("UpdatedAt should be set to approximately now")
	}
}

func TestFeedbackService_Create_InsertError(t *testing.T) {
	col := &feedbackMockCollection{insertOneErr: errors.New("db error")}
	svc := NewFeedbackService(col)

	_, err := svc.Create(feedbackCreateDto(), nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// GetAll
// ---------------------------------------------------------------------------

func TestFeedbackService_GetAll_ReturnsResults(t *testing.T) {
	uid := newUserID()
	fb1 := sampleFeedback(uid)
	fb2 := sampleFeedback(uid)
	col := &feedbackMockCollection{
		findResult:  feedbacksCursor([]model.Feedback{fb1, fb2}),
		countResult: 2,
	}
	svc := NewFeedbackService(col)

	resp, err := svc.GetAll(model.FeedbackListDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}})
	if err != nil {
		t.Fatal(err)
	}
	data := resp.Data.([]model.Feedback)
	if len(data) != 2 {
		t.Errorf("expected 2 feedbacks, got %d", len(data))
	}
	if resp.Metadata.TotalItems != 2 {
		t.Errorf("expected TotalItems=2, got %d", resp.Metadata.TotalItems)
	}
}

func TestFeedbackService_GetAll_EmptyResult(t *testing.T) {
	col := &feedbackMockCollection{
		findResult:  feedbacksCursor([]model.Feedback{}),
		countResult: 0,
	}
	svc := NewFeedbackService(col)

	resp, err := svc.GetAll(model.FeedbackListDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}})
	if err != nil {
		t.Fatal(err)
	}
	data := resp.Data.([]model.Feedback)
	if len(data) != 0 {
		t.Errorf("expected 0 feedbacks, got %d", len(data))
	}
}

func TestFeedbackService_GetAll_DefaultsPagination(t *testing.T) {
	col := &feedbackMockCollection{
		findResult:  feedbacksCursor([]model.Feedback{}),
		countResult: 0,
	}
	svc := NewFeedbackService(col)

	// Page=0, Limit=0 should default to page=1, limit=10
	resp, err := svc.GetAll(model.FeedbackListDto{})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Metadata.CurrentPage != 1 {
		t.Errorf("expected CurrentPage=1, got %d", resp.Metadata.CurrentPage)
	}
	if resp.Metadata.ItemsPerPage != 10 {
		t.Errorf("expected ItemsPerPage=10, got %d", resp.Metadata.ItemsPerPage)
	}
}

func TestFeedbackService_GetAll_FindError(t *testing.T) {
	col := &feedbackMockCollection{findErr: errors.New("find failed")}
	svc := NewFeedbackService(col)

	_, err := svc.GetAll(model.FeedbackListDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFeedbackService_GetAll_CursorDecodeError(t *testing.T) {
	col := &feedbackMockCollection{
		findResult: &genericCursor{err: errors.New("decode error")},
	}
	svc := NewFeedbackService(col)

	_, err := svc.GetAll(model.FeedbackListDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFeedbackService_GetAll_CountError(t *testing.T) {
	col := &feedbackMockCollection{
		findResult: feedbacksCursor([]model.Feedback{}),
		countErr:   errors.New("count failed"),
	}
	svc := NewFeedbackService(col)

	_, err := svc.GetAll(model.FeedbackListDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFeedbackService_GetAll_FilterByRating(t *testing.T) {
	uid := newUserID()
	fb := sampleFeedback(uid)
	fb.Rating = 5
	col := &feedbackMockCollection{
		findResult:  feedbacksCursor([]model.Feedback{fb}),
		countResult: 1,
	}
	svc := NewFeedbackService(col)

	rating := 5
	resp, err := svc.GetAll(model.FeedbackListDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}, Rating: &rating})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Metadata.TotalItems != 1 {
		t.Errorf("expected 1 item, got %d", resp.Metadata.TotalItems)
	}
}

func TestFeedbackService_GetAll_FilterByEmail(t *testing.T) {
	uid := newUserID()
	fb := sampleFeedback(uid)
	col := &feedbackMockCollection{
		findResult:  feedbacksCursor([]model.Feedback{fb}),
		countResult: 1,
	}
	svc := NewFeedbackService(col)

	email := "alice"
	resp, err := svc.GetAll(model.FeedbackListDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}, Email: &email})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Metadata.TotalItems != 1 {
		t.Errorf("expected 1 item, got %d", resp.Metadata.TotalItems)
	}
}

func TestFeedbackService_GetAll_TotalPagesAtLeastOne(t *testing.T) {
	col := &feedbackMockCollection{
		findResult:  feedbacksCursor([]model.Feedback{}),
		countResult: 0,
	}
	svc := NewFeedbackService(col)

	resp, err := svc.GetAll(model.FeedbackListDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Metadata.TotalPages < 1 {
		t.Errorf("expected TotalPages >= 1, got %v", resp.Metadata.TotalPages)
	}
}

// ---------------------------------------------------------------------------
// GetById
// ---------------------------------------------------------------------------

func TestFeedbackService_GetById_Success(t *testing.T) {
	uid := newUserID()
	fb := sampleFeedback(uid)
	col := &feedbackMockCollection{
		findOneQueue: []SingleResult{&feedbackSingleResult{fb: fb}},
	}
	svc := NewFeedbackService(col)

	got, err := svc.GetById(fb.ID.Hex())
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != fb.ID {
		t.Errorf("expected ID %s, got %s", fb.ID.Hex(), got.ID.Hex())
	}
}

func TestFeedbackService_GetById_InvalidID(t *testing.T) {
	svc := NewFeedbackService(&feedbackMockCollection{})
	_, err := svc.GetById("not-an-objectid")
	if err == nil {
		t.Fatal("expected error for invalid ID")
	}
}

func TestFeedbackService_GetById_NotFound(t *testing.T) {
	col := &feedbackMockCollection{
		findOneQueue: []SingleResult{&errSingleResult{err: mongo.ErrNoDocuments}},
	}
	svc := NewFeedbackService(col)

	_, err := svc.GetById(primitive.NewObjectID().Hex())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "feedback not found" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestFeedbackService_GetById_DBError(t *testing.T) {
	col := &feedbackMockCollection{
		findOneQueue: []SingleResult{&errSingleResult{err: errors.New("db error")}},
	}
	svc := NewFeedbackService(col)

	_, err := svc.GetById(primitive.NewObjectID().Hex())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestFeedbackService_Update_Success(t *testing.T) {
	uid := newUserID()
	fb := sampleFeedback(uid)
	updated := fb
	updated.Rating = 4

	col := &feedbackMockCollection{
		// First FindOne returns the existing record; FindOneAndUpdate returns updated record.
		findOneQueue:           []SingleResult{&feedbackSingleResult{fb: fb}},
		findOneAndUpdateResult: &feedbackSingleResult{fb: updated},
	}
	svc := NewFeedbackService(col)

	rating := 4
	got, err := svc.Update(fb.ID.Hex(), model.UpdateFeedbackDto{Rating: &rating}, uid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Rating != 4 {
		t.Errorf("expected rating 4, got %d", got.Rating)
	}
}

func TestFeedbackService_Update_InvalidID(t *testing.T) {
	svc := NewFeedbackService(&feedbackMockCollection{})
	_, err := svc.Update("bad-id", model.UpdateFeedbackDto{}, primitive.NewObjectID())
	if err == nil {
		t.Fatal("expected error for invalid ID")
	}
}

func TestFeedbackService_Update_NotFound(t *testing.T) {
	col := &feedbackMockCollection{
		findOneQueue: []SingleResult{&errSingleResult{err: mongo.ErrNoDocuments}},
	}
	svc := NewFeedbackService(col)

	_, err := svc.Update(primitive.NewObjectID().Hex(), model.UpdateFeedbackDto{}, primitive.NewObjectID())
	if err == nil || err.Error() != "feedback not found" {
		t.Fatalf("expected 'feedback not found', got %v", err)
	}
}

func TestFeedbackService_Update_Unauthorized(t *testing.T) {
	uid := newUserID()
	fb := sampleFeedback(uid)

	col := &feedbackMockCollection{
		findOneQueue: []SingleResult{&feedbackSingleResult{fb: fb}},
	}
	svc := NewFeedbackService(col)

	otherUser := primitive.NewObjectID()
	_, err := svc.Update(fb.ID.Hex(), model.UpdateFeedbackDto{}, otherUser)
	if err == nil || err.Error() != "unauthorized to update this feedback" {
		t.Fatalf("expected unauthorized error, got %v", err)
	}
}

func TestFeedbackService_Update_NilUserIDOnExisting(t *testing.T) {
	uid := newUserID()
	fb := sampleFeedback(uid)
	fb.UserID = nil // simulate feedback with no userID

	col := &feedbackMockCollection{
		findOneQueue: []SingleResult{&feedbackSingleResult{fb: fb}},
	}
	svc := NewFeedbackService(col)

	_, err := svc.Update(fb.ID.Hex(), model.UpdateFeedbackDto{}, uid)
	if err == nil || err.Error() != "unauthorized to update this feedback" {
		t.Fatalf("expected unauthorized error, got %v", err)
	}
}

func TestFeedbackService_Update_FindOneAndUpdateError(t *testing.T) {
	uid := newUserID()
	fb := sampleFeedback(uid)

	col := &feedbackMockCollection{
		findOneQueue:           []SingleResult{&feedbackSingleResult{fb: fb}},
		findOneAndUpdateResult: &errSingleResult{err: errors.New("update failed")},
	}
	svc := NewFeedbackService(col)

	rating := 3
	_, err := svc.Update(fb.ID.Hex(), model.UpdateFeedbackDto{Rating: &rating}, uid)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFeedbackService_Update_CommentOnly(t *testing.T) {
	uid := newUserID()
	fb := sampleFeedback(uid)
	updated := fb
	updated.Comment = "Updated comment"

	col := &feedbackMockCollection{
		findOneQueue:           []SingleResult{&feedbackSingleResult{fb: fb}},
		findOneAndUpdateResult: &feedbackSingleResult{fb: updated},
	}
	svc := NewFeedbackService(col)

	comment := "Updated comment"
	got, err := svc.Update(fb.ID.Hex(), model.UpdateFeedbackDto{Comment: &comment}, uid)
	if err != nil {
		t.Fatal(err)
	}
	if got.Comment != "Updated comment" {
		t.Errorf("unexpected comment: %s", got.Comment)
	}
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestFeedbackService_Delete_Success(t *testing.T) {
	uid := newUserID()
	fb := sampleFeedback(uid)

	col := &feedbackMockCollection{
		findOneQueue:    []SingleResult{&feedbackSingleResult{fb: fb}},
		deleteOneResult: &mongo.DeleteResult{DeletedCount: 1},
	}
	svc := NewFeedbackService(col)

	err := svc.Delete(fb.ID.Hex(), uid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if col.deleteOneCalls != 1 {
		t.Errorf("expected 1 DeleteOne call, got %d", col.deleteOneCalls)
	}
}

func TestFeedbackService_Delete_InvalidID(t *testing.T) {
	svc := NewFeedbackService(&feedbackMockCollection{})
	err := svc.Delete("bad-id", primitive.NewObjectID())
	if err == nil {
		t.Fatal("expected error for invalid ID")
	}
}

func TestFeedbackService_Delete_NotFound(t *testing.T) {
	col := &feedbackMockCollection{
		findOneQueue: []SingleResult{&errSingleResult{err: mongo.ErrNoDocuments}},
	}
	svc := NewFeedbackService(col)

	err := svc.Delete(primitive.NewObjectID().Hex(), primitive.NewObjectID())
	if err == nil || err.Error() != "feedback not found" {
		t.Fatalf("expected 'feedback not found', got %v", err)
	}
}

func TestFeedbackService_Delete_Unauthorized(t *testing.T) {
	uid := newUserID()
	fb := sampleFeedback(uid)

	col := &feedbackMockCollection{
		findOneQueue: []SingleResult{&feedbackSingleResult{fb: fb}},
	}
	svc := NewFeedbackService(col)

	err := svc.Delete(fb.ID.Hex(), primitive.NewObjectID())
	if err == nil || err.Error() != "unauthorized to delete this feedback" {
		t.Fatalf("expected unauthorized error, got %v", err)
	}
}

func TestFeedbackService_Delete_NilUserIDOnExisting(t *testing.T) {
	uid := newUserID()
	fb := sampleFeedback(uid)
	fb.UserID = nil

	col := &feedbackMockCollection{
		findOneQueue: []SingleResult{&feedbackSingleResult{fb: fb}},
	}
	svc := NewFeedbackService(col)

	err := svc.Delete(fb.ID.Hex(), uid)
	if err == nil || err.Error() != "unauthorized to delete this feedback" {
		t.Fatalf("expected unauthorized error, got %v", err)
	}
}

func TestFeedbackService_Delete_DeleteError(t *testing.T) {
	uid := newUserID()
	fb := sampleFeedback(uid)

	col := &feedbackMockCollection{
		findOneQueue: []SingleResult{&feedbackSingleResult{fb: fb}},
		deleteOneErr: errors.New("delete failed"),
	}
	svc := NewFeedbackService(col)

	err := svc.Delete(fb.ID.Hex(), uid)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFeedbackService_Delete_ZeroDeletedCount(t *testing.T) {
	uid := newUserID()
	fb := sampleFeedback(uid)

	col := &feedbackMockCollection{
		findOneQueue:    []SingleResult{&feedbackSingleResult{fb: fb}},
		deleteOneResult: &mongo.DeleteResult{DeletedCount: 0},
	}
	svc := NewFeedbackService(col)

	err := svc.Delete(fb.ID.Hex(), uid)
	if err == nil || err.Error() != "feedback not found" {
		t.Fatalf("expected 'feedback not found', got %v", err)
	}
}
