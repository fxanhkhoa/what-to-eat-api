package service

import (
	"context"
	"errors"
	"testing"
	"time"
	"what-to-eat/be/model"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ---------------------------------------------------------------------------
// DishVote-specific mock collection
// ---------------------------------------------------------------------------

type dishVoteMockCollection struct {
	insertOneResult *mongo.InsertOneResult
	insertOneErr    error

	findOneResult          SingleResult
	findOneAndUpdateResult SingleResult

	countResult int64
	countErr    error

	findResult Cursor
	findErr    error

	insertOneCalls int
}

func (m *dishVoteMockCollection) InsertOne(ctx context.Context, doc interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	m.insertOneCalls++
	return m.insertOneResult, m.insertOneErr
}

func (m *dishVoteMockCollection) FindOneAndUpdate(ctx context.Context, filter, update interface{}, opts ...*options.FindOneAndUpdateOptions) SingleResult {
	return m.findOneAndUpdateResult
}

func (m *dishVoteMockCollection) CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	return m.countResult, m.countErr
}

func (m *dishVoteMockCollection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (Cursor, error) {
	if m.findResult != nil {
		return m.findResult, m.findErr
	}
	return emptyCursor(), m.findErr
}

func (m *dishVoteMockCollection) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) SingleResult {
	return m.findOneResult
}

func (m *dishVoteMockCollection) UpdateOne(ctx context.Context, filter, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return &mongo.UpdateResult{ModifiedCount: 1}, nil
}

func (m *dishVoteMockCollection) UpdateMany(ctx context.Context, filter, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return nil, nil
}

func (m *dishVoteMockCollection) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return &mongo.DeleteResult{DeletedCount: 1}, nil
}

func (m *dishVoteMockCollection) Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (Cursor, error) {
	return emptyCursor(), nil
}

// ---------------------------------------------------------------------------
// SingleResult helpers
// ---------------------------------------------------------------------------

type dishVoteSingleResult struct {
	vote model.DishVote
	err  error
}

func (r *dishVoteSingleResult) Decode(v interface{}) error {
	if r.err != nil {
		return r.err
	}
	*(v.(*model.DishVote)) = r.vote
	return nil
}
func (r *dishVoteSingleResult) Err() error { return r.err }

// ---------------------------------------------------------------------------
// Cursor helper
// ---------------------------------------------------------------------------

func dishVotesCursor(votes []*model.DishVote) Cursor {
	return &genericCursor{
		decodeFn: func(results interface{}) error {
			*(results.(*[]*model.DishVote)) = votes
			return nil
		},
	}
}

// ---------------------------------------------------------------------------
// Shared fixtures
// ---------------------------------------------------------------------------

func sampleProfile() *model.JwtCustomClaims {
	return &model.JwtCustomClaims{RegisteredClaims: jwt.RegisteredClaims{ID: "user123"}, RoleName: "USER"}
}

func adminProfile() *model.JwtCustomClaims {
	return &model.JwtCustomClaims{RegisteredClaims: jwt.RegisteredClaims{ID: "admin1"}, RoleName: "ADMIN"}
}

func strPtr(s string) *string { return &s }

func sampleCreateDto() model.CreateDishVoteDto {
	return model.CreateDishVoteDto{
		Title:       strPtr("Lunch vote"),
		Description: strPtr("Pick your lunch"),
		DishVoteItems: []*model.DishVoteItem{
			{Slug: "pho", VoteUser: []*string{}, VoteAnonymous: []*string{}},
		},
	}
}

func sampleDishVote() model.DishVote {
	now := time.Now()
	id := primitive.NewObjectID().Hex()
	uid := "user123"
	title := "Lunch vote"
	return model.DishVote{
		ID:          id,
		Title:       &title,
		Description: strPtr("Pick your lunch"),
		DishVoteItems: []*model.DishVoteItem{
			{Slug: "pho"},
		},
		Deleted:   false,
		CreatedBy: &uid,
		CreatedAt: &now,
		UpdatedAt: &now,
	}
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestDishVoteService_Create_Success(t *testing.T) {
	oid := primitive.NewObjectID()
	col := &dishVoteMockCollection{
		insertOneResult: &mongo.InsertOneResult{InsertedID: oid},
	}
	svc := NewDishVoteService(col)

	vote, err := svc.Create(sampleCreateDto(), sampleProfile())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vote.ID != oid.Hex() {
		t.Errorf("expected ID %s, got %s", oid.Hex(), vote.ID)
	}
	if *vote.Title != "Lunch vote" {
		t.Errorf("unexpected title: %s", *vote.Title)
	}
	if col.insertOneCalls != 1 {
		t.Errorf("expected 1 InsertOne call, got %d", col.insertOneCalls)
	}
}

func TestDishVoteService_Create_SetsDeletedFalse(t *testing.T) {
	oid := primitive.NewObjectID()
	col := &dishVoteMockCollection{insertOneResult: &mongo.InsertOneResult{InsertedID: oid}}
	svc := NewDishVoteService(col)

	vote, err := svc.Create(sampleCreateDto(), sampleProfile())
	if err != nil {
		t.Fatal(err)
	}
	if vote.Deleted {
		t.Error("expected Deleted=false on new vote")
	}
}

func TestDishVoteService_Create_SetsCreatedBy(t *testing.T) {
	oid := primitive.NewObjectID()
	col := &dishVoteMockCollection{insertOneResult: &mongo.InsertOneResult{InsertedID: oid}}
	svc := NewDishVoteService(col)

	profile := sampleProfile()
	vote, err := svc.Create(sampleCreateDto(), profile)
	if err != nil {
		t.Fatal(err)
	}
	if vote.CreatedBy == nil || *vote.CreatedBy != profile.ID {
		t.Errorf("expected CreatedBy=%s", profile.ID)
	}
}

func TestDishVoteService_Create_InsertError(t *testing.T) {
	col := &dishVoteMockCollection{insertOneErr: errors.New("db error")}
	svc := NewDishVoteService(col)

	_, err := svc.Create(sampleCreateDto(), sampleProfile())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDishVoteService_Create_WithItems(t *testing.T) {
	oid := primitive.NewObjectID()
	col := &dishVoteMockCollection{insertOneResult: &mongo.InsertOneResult{InsertedID: oid}}
	svc := NewDishVoteService(col)

	dto := sampleCreateDto()
	dto.DishVoteItems = append(dto.DishVoteItems, &model.DishVoteItem{Slug: "bun-bo", IsCustom: false})

	vote, err := svc.Create(dto, sampleProfile())
	if err != nil {
		t.Fatal(err)
	}
	if len(vote.DishVoteItems) != 2 {
		t.Errorf("expected 2 items, got %d", len(vote.DishVoteItems))
	}
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestDishVoteService_Update_Success(t *testing.T) {
	expected := sampleDishVote()
	col := &dishVoteMockCollection{
		findOneAndUpdateResult: &dishVoteSingleResult{vote: expected},
	}
	svc := NewDishVoteService(col)

	dto := model.UpdateDishVoteDto{
		ID:    expected.ID,
		Title: strPtr("Updated title"),
	}
	vote, err := svc.Update(dto, sampleProfile())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vote == nil {
		t.Fatal("expected non-nil vote")
	}
}

func TestDishVoteService_Update_InvalidID(t *testing.T) {
	col := &dishVoteMockCollection{}
	svc := NewDishVoteService(col)

	dto := model.UpdateDishVoteDto{ID: "not-an-objectid"}
	_, err := svc.Update(dto, sampleProfile())
	if err == nil {
		t.Fatal("expected error for invalid ID")
	}
}

func TestDishVoteService_Update_FindOneAndUpdateError(t *testing.T) {
	oid := primitive.NewObjectID()
	col := &dishVoteMockCollection{
		findOneAndUpdateResult: &dishVoteSingleResult{err: errors.New("not found")},
	}
	svc := NewDishVoteService(col)

	dto := model.UpdateDishVoteDto{ID: oid.Hex(), Title: strPtr("x")}
	_, err := svc.Update(dto, sampleProfile())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDishVoteService_Update_NilProfile(t *testing.T) {
	expected := sampleDishVote()
	col := &dishVoteMockCollection{
		findOneAndUpdateResult: &dishVoteSingleResult{vote: expected},
	}
	svc := NewDishVoteService(col)

	dto := model.UpdateDishVoteDto{ID: expected.ID, Title: strPtr("no profile")}
	// Should not panic with nil profile (updatedBy defaults to "")
	vote, err := svc.Update(dto, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vote == nil {
		t.Fatal("expected non-nil vote")
	}
}

// ---------------------------------------------------------------------------
// Remove
// ---------------------------------------------------------------------------

func TestDishVoteService_Remove_Success(t *testing.T) {
	expected := sampleDishVote()
	col := &dishVoteMockCollection{
		findOneAndUpdateResult: &dishVoteSingleResult{vote: expected},
	}
	svc := NewDishVoteService(col)

	vote, err := svc.Remove(expected.ID, sampleProfile())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vote == nil {
		t.Fatal("expected non-nil vote")
	}
}

func TestDishVoteService_Remove_NotFound(t *testing.T) {
	col := &dishVoteMockCollection{
		findOneAndUpdateResult: &dishVoteSingleResult{err: errors.New("not found")},
	}
	svc := NewDishVoteService(col)

	_, err := svc.Remove("someid", sampleProfile())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// Find
// ---------------------------------------------------------------------------

func TestDishVoteService_Find_ReturnsResults(t *testing.T) {
	v1, v2 := sampleDishVote(), sampleDishVote()
	col := &dishVoteMockCollection{
		countResult: 2,
		findResult:  dishVotesCursor([]*model.DishVote{&v1, &v2}),
	}
	svc := NewDishVoteService(col)

	keyword := ""
	results, count, err := svc.Find(model.QueryDishVoteDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}, Keyword: &keyword}, sampleProfile())
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestDishVoteService_Find_EmptyResult(t *testing.T) {
	col := &dishVoteMockCollection{
		countResult: 0,
		findResult:  dishVotesCursor([]*model.DishVote{}),
	}
	svc := NewDishVoteService(col)

	results, count, err := svc.Find(model.QueryDishVoteDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}}, sampleProfile())
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 || len(results) != 0 {
		t.Error("expected empty result")
	}
}

func TestDishVoteService_Find_CountError(t *testing.T) {
	col := &dishVoteMockCollection{countErr: errors.New("count failed")}
	svc := NewDishVoteService(col)

	_, _, err := svc.Find(model.QueryDishVoteDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}}, sampleProfile())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDishVoteService_Find_AdminSeesAll(t *testing.T) {
	v := sampleDishVote()
	col := &dishVoteMockCollection{
		countResult: 1,
		findResult:  dishVotesCursor([]*model.DishVote{&v}),
	}
	svc := NewDishVoteService(col)

	// Admin profile — no createdBy filter applied
	results, count, err := svc.Find(model.QueryDishVoteDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}}, adminProfile())
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 || len(results) != 1 {
		t.Errorf("expected 1 result for admin, got count=%d len=%d", count, len(results))
	}
}

func TestDishVoteService_Find_NilProfile(t *testing.T) {
	v := sampleDishVote()
	col := &dishVoteMockCollection{
		countResult: 1,
		findResult:  dishVotesCursor([]*model.DishVote{&v}),
	}
	svc := NewDishVoteService(col)

	// nil profile should not panic — no createdBy filter added
	results, _, err := svc.Find(model.QueryDishVoteDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestDishVoteService_Find_WithKeyword(t *testing.T) {
	v := sampleDishVote()
	col := &dishVoteMockCollection{
		countResult: 1,
		findResult:  dishVotesCursor([]*model.DishVote{&v}),
	}
	svc := NewDishVoteService(col)

	keyword := "pho"
	results, count, err := svc.Find(model.QueryDishVoteDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}, Keyword: &keyword}, sampleProfile())
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 || len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

// ---------------------------------------------------------------------------
// FindOne
// ---------------------------------------------------------------------------

func TestDishVoteService_FindOne_Success(t *testing.T) {
	expected := sampleDishVote()
	oid := primitive.NewObjectID()
	expected.ID = oid.Hex()
	col := &dishVoteMockCollection{
		findOneResult: &dishVoteSingleResult{vote: expected},
	}
	svc := NewDishVoteService(col)

	vote, err := svc.FindOne(oid.Hex())
	if err != nil {
		t.Fatal(err)
	}
	if vote.ID != expected.ID {
		t.Errorf("expected ID %s, got %s", expected.ID, vote.ID)
	}
}

func TestDishVoteService_FindOne_InvalidID(t *testing.T) {
	svc := NewDishVoteService(&dishVoteMockCollection{})
	_, err := svc.FindOne("not-an-objectid")
	if err == nil {
		t.Fatal("expected error for invalid ID")
	}
}

func TestDishVoteService_FindOne_NotFound(t *testing.T) {
	oid := primitive.NewObjectID()
	col := &dishVoteMockCollection{
		findOneResult: &dishVoteSingleResult{err: errors.New("not found")},
	}
	svc := NewDishVoteService(col)

	_, err := svc.FindOne(oid.Hex())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
