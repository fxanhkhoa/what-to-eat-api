package service

import (
	"context"
	"errors"
	"testing"
	"what-to-eat/be/model"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ---------------------------------------------------------------------------
// ingredientMockCollection — CollectionInterface impl for ingredient tests
// ---------------------------------------------------------------------------

type ingredientMockCollection struct {
	findOneAndUpdateResult SingleResult

	countResult int64
	countErr    error

	findResult Cursor
	findErr    error

	findOneResult SingleResult

	aggregateResult Cursor
	aggregateErr    error
}

func (m *ingredientMockCollection) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	return &mongo.InsertOneResult{InsertedID: primitive.NewObjectID()}, nil
}

func (m *ingredientMockCollection) FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...*options.FindOneAndUpdateOptions) SingleResult {
	return m.findOneAndUpdateResult
}

func (m *ingredientMockCollection) CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	return m.countResult, m.countErr
}

func (m *ingredientMockCollection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (Cursor, error) {
	if m.findResult != nil {
		return m.findResult, m.findErr
	}
	return emptyCursor(), m.findErr
}

func (m *ingredientMockCollection) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) SingleResult {
	return m.findOneResult
}

func (m *ingredientMockCollection) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return &mongo.UpdateResult{MatchedCount: 1}, nil
}

func (m *ingredientMockCollection) UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return nil, nil
}

func (m *ingredientMockCollection) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return &mongo.DeleteResult{DeletedCount: 1}, nil
}

func (m *ingredientMockCollection) Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (Cursor, error) {
	if m.aggregateErr != nil {
		return nil, m.aggregateErr
	}
	if m.aggregateResult != nil {
		return m.aggregateResult, nil
	}
	return emptyCursor(), nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func decodeIngredientFn(ing model.Ingredient) func(interface{}) error {
	return func(v interface{}) error {
		*(v.(*model.Ingredient)) = ing
		return nil
	}
}

func ingredientsCursor(ings []*model.Ingredient) Cursor {
	return &genericCursor{
		decodeFn: func(results interface{}) error {
			*(results.(*[]*model.Ingredient)) = ings
			return nil
		},
	}
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestIngredientCreate_FindOneAndUpdateError(t *testing.T) {
	col := &ingredientMockCollection{
		findOneAndUpdateResult: &mockSingleResult{err: errors.New("write error")},
	}
	svc := NewIngredientService(col)
	_, err := svc.Create(model.CreateIngredientDto{Slug: "salt"}, &model.JwtCustomClaims{})
	if err == nil || err.Error() != "write error" {
		t.Errorf("expected 'write error', got %v", err)
	}
}

func TestIngredientCreate_Success(t *testing.T) {
	expected := model.Ingredient{Slug: "salt"}
	col := &ingredientMockCollection{
		findOneAndUpdateResult: &mockSingleResult{decodeFn: decodeIngredientFn(expected)},
	}
	svc := NewIngredientService(col)
	ing, err := svc.Create(model.CreateIngredientDto{Slug: "salt"}, &model.JwtCustomClaims{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ing.Slug != "salt" {
		t.Errorf("expected slug 'salt', got %s", ing.Slug)
	}
}

func TestIngredientCreate_DecodeError(t *testing.T) {
	col := &ingredientMockCollection{
		findOneAndUpdateResult: &mockSingleResult{decodeFn: func(v interface{}) error {
			return errors.New("decode error")
		}},
	}
	svc := NewIngredientService(col)
	_, err := svc.Create(model.CreateIngredientDto{Slug: "salt"}, &model.JwtCustomClaims{})
	if err == nil || err.Error() != "decode error" {
		t.Errorf("expected 'decode error', got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestIngredientUpdate_InvalidID(t *testing.T) {
	svc := NewIngredientService(&ingredientMockCollection{})
	_, err := svc.Update(model.UpdateIngredientDto{ID: "bad-id"}, &model.JwtCustomClaims{})
	if err == nil || err.Error() != "invalid ID format" {
		t.Errorf("expected 'invalid ID format', got %v", err)
	}
}

func TestIngredientUpdate_FindOneAndUpdateError(t *testing.T) {
	col := &ingredientMockCollection{
		findOneAndUpdateResult: &mockSingleResult{err: errors.New("update error")},
	}
	svc := NewIngredientService(col)
	_, err := svc.Update(model.UpdateIngredientDto{ID: primitive.NewObjectID().Hex()}, &model.JwtCustomClaims{})
	if err == nil || err.Error() != "update error" {
		t.Errorf("expected 'update error', got %v", err)
	}
}

func TestIngredientUpdate_Success(t *testing.T) {
	expected := model.Ingredient{Slug: "pepper-updated"}
	col := &ingredientMockCollection{
		findOneAndUpdateResult: &mockSingleResult{decodeFn: decodeIngredientFn(expected)},
	}
	svc := NewIngredientService(col)
	ing, err := svc.Update(
		model.UpdateIngredientDto{ID: primitive.NewObjectID().Hex(), Slug: "pepper-updated"},
		&model.JwtCustomClaims{},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ing.Slug != "pepper-updated" {
		t.Errorf("expected slug 'pepper-updated', got %s", ing.Slug)
	}
}

// ---------------------------------------------------------------------------
// Remove
// ---------------------------------------------------------------------------

func TestIngredientRemove_InvalidID(t *testing.T) {
	svc := NewIngredientService(&ingredientMockCollection{})
	_, err := svc.Remove("bad-id", &model.JwtCustomClaims{})
	if err == nil || err.Error() != "invalid ID format" {
		t.Errorf("expected 'invalid ID format', got %v", err)
	}
}

func TestIngredientRemove_FindOneAndUpdateError(t *testing.T) {
	col := &ingredientMockCollection{
		findOneAndUpdateResult: &mockSingleResult{err: errors.New("remove error")},
	}
	svc := NewIngredientService(col)
	_, err := svc.Remove(primitive.NewObjectID().Hex(), &model.JwtCustomClaims{})
	if err == nil || err.Error() != "remove error" {
		t.Errorf("expected 'remove error', got %v", err)
	}
}

func TestIngredientRemove_Success(t *testing.T) {
	expected := model.Ingredient{Slug: "gone", Deleted: true}
	col := &ingredientMockCollection{
		findOneAndUpdateResult: &mockSingleResult{decodeFn: decodeIngredientFn(expected)},
	}
	svc := NewIngredientService(col)
	ing, err := svc.Remove(primitive.NewObjectID().Hex(), &model.JwtCustomClaims{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ing.Deleted {
		t.Error("expected ingredient to be marked deleted")
	}
}

// ---------------------------------------------------------------------------
// Find
// ---------------------------------------------------------------------------

func TestIngredientFind_CountError(t *testing.T) {
	col := &ingredientMockCollection{countErr: errors.New("count error")}
	svc := NewIngredientService(col)
	_, _, err := svc.Find(model.QueryIngredientDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}})
	if err == nil || err.Error() != "count error" {
		t.Errorf("expected 'count error', got %v", err)
	}
}

func TestIngredientFind_FindError(t *testing.T) {
	// Find errors are only logged; cursor.All on the empty fallback cursor succeeds.
	col := &ingredientMockCollection{
		countResult: 1,
		findErr:     errors.New("find error"),
	}
	svc := NewIngredientService(col)
	_, count, err := svc.Find(model.QueryIngredientDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}})
	if err != nil {
		t.Errorf("expected no propagated error, got %v", err)
	}
	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}
}

func TestIngredientFind_CursorError(t *testing.T) {
	col := &ingredientMockCollection{
		countResult: 1,
		findResult:  &genericCursor{err: errors.New("cursor error")},
	}
	svc := NewIngredientService(col)
	_, _, err := svc.Find(model.QueryIngredientDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}})
	if err == nil || err.Error() != "cursor error" {
		t.Errorf("expected 'cursor error', got %v", err)
	}
}

func TestIngredientFind_Success(t *testing.T) {
	ings := []*model.Ingredient{{Slug: "salt"}, {Slug: "pepper"}}
	col := &ingredientMockCollection{
		countResult: 2,
		findResult:  ingredientsCursor(ings),
	}
	svc := NewIngredientService(col)
	keyword := "salt"
	cats := []string{"spice"}
	result, count, err := svc.Find(model.QueryIngredientDto{
		BaseDto:            model.BaseDto{Page: 1, Limit: 10},
		Keyword:            &keyword,
		IngredientCategory: &cats,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 ingredients, got %d", len(result))
	}
}

func TestIngredientFind_NoFilters(t *testing.T) {
	ings := []*model.Ingredient{{Slug: "salt"}}
	col := &ingredientMockCollection{
		countResult: 1,
		findResult:  ingredientsCursor(ings),
	}
	svc := NewIngredientService(col)
	result, count, err := svc.Find(model.QueryIngredientDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 || len(result) != 1 {
		t.Errorf("unexpected result: count=%d len=%d", count, len(result))
	}
}

// ---------------------------------------------------------------------------
// FindOne
// ---------------------------------------------------------------------------

func TestIngredientFindOne_InvalidID(t *testing.T) {
	svc := NewIngredientService(&ingredientMockCollection{})
	_, err := svc.FindOne("bad-id")
	if err == nil || err.Error() != "invalid ID format" {
		t.Errorf("expected 'invalid ID format', got %v", err)
	}
}

func TestIngredientFindOne_FindOneError(t *testing.T) {
	col := &ingredientMockCollection{
		findOneResult: &mockSingleResult{err: mongo.ErrNoDocuments},
	}
	svc := NewIngredientService(col)
	_, err := svc.FindOne(primitive.NewObjectID().Hex())
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestIngredientFindOne_Success(t *testing.T) {
	expected := model.Ingredient{Slug: "salt"}
	col := &ingredientMockCollection{
		findOneResult: &mockSingleResult{decodeFn: decodeIngredientFn(expected)},
	}
	svc := NewIngredientService(col)
	ing, err := svc.FindOne(primitive.NewObjectID().Hex())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ing.Slug != "salt" {
		t.Errorf("expected slug 'salt', got %s", ing.Slug)
	}
}

// ---------------------------------------------------------------------------
// FindOneBySlug
// ---------------------------------------------------------------------------

func TestIngredientFindOneBySlug_Error(t *testing.T) {
	col := &ingredientMockCollection{
		findOneResult: &mockSingleResult{err: mongo.ErrNoDocuments},
	}
	svc := NewIngredientService(col)
	_, err := svc.FindOneBySlug("missing")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestIngredientFindOneBySlug_Success(t *testing.T) {
	expected := model.Ingredient{Slug: "pepper"}
	col := &ingredientMockCollection{
		findOneResult: &mockSingleResult{decodeFn: decodeIngredientFn(expected)},
	}
	svc := NewIngredientService(col)
	ing, err := svc.FindOneBySlug("pepper")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ing.Slug != "pepper" {
		t.Errorf("expected slug 'pepper', got %s", ing.Slug)
	}
}

// ---------------------------------------------------------------------------
// FindTitleByLang
// ---------------------------------------------------------------------------

func TestIngredientFindTitleByLang_FindOneError(t *testing.T) {
	col := &ingredientMockCollection{
		findOneResult: &mockSingleResult{err: mongo.ErrNoDocuments},
	}
	svc := NewIngredientService(col)
	_, err := svc.FindTitleByLang("Muối", "vi")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestIngredientFindTitleByLang_DecodeError(t *testing.T) {
	col := &ingredientMockCollection{
		findOneResult: &mockSingleResult{decodeFn: func(v interface{}) error {
			return errors.New("decode error")
		}},
	}
	svc := NewIngredientService(col)
	_, err := svc.FindTitleByLang("Muối", "vi")
	if err == nil || err.Error() != "decode error" {
		t.Errorf("expected 'decode error', got %v", err)
	}
}

func TestIngredientFindTitleByLang_Success(t *testing.T) {
	expected := model.Ingredient{Slug: "salt"}
	col := &ingredientMockCollection{
		findOneResult: &mockSingleResult{decodeFn: decodeIngredientFn(expected)},
	}
	svc := NewIngredientService(col)
	ing, err := svc.FindTitleByLang("Salt", "en")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ing.Slug != "salt" {
		t.Errorf("expected slug 'salt', got %s", ing.Slug)
	}
}

// ---------------------------------------------------------------------------
// Random
// ---------------------------------------------------------------------------

func TestIngredientRandom_AggregateError(t *testing.T) {
	col := &ingredientMockCollection{
		aggregateErr: errors.New("agg error"),
	}
	svc := NewIngredientService(col)
	limit := 5
	_, err := svc.Random(&limit, nil)
	if err == nil || err.Error() != "agg error" {
		t.Errorf("expected 'agg error', got %v", err)
	}
}

func TestIngredientRandom_CursorError(t *testing.T) {
	col := &ingredientMockCollection{
		aggregateResult: &genericCursor{err: errors.New("cursor error")},
	}
	svc := NewIngredientService(col)
	limit := 5
	_, err := svc.Random(&limit, nil)
	if err == nil || err.Error() != "cursor error" {
		t.Errorf("expected 'cursor error', got %v", err)
	}
}

func TestIngredientRandom_Success(t *testing.T) {
	ings := []*model.Ingredient{{Slug: "salt"}, {Slug: "pepper"}}
	col := &ingredientMockCollection{
		aggregateResult: ingredientsCursor(ings),
	}
	svc := NewIngredientService(col)
	limit := 2
	result, err := svc.Random(&limit, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 ingredients, got %d", len(result))
	}
}

func TestIngredientRandom_WithCategory(t *testing.T) {
	ings := []*model.Ingredient{{Slug: "chili"}}
	col := &ingredientMockCollection{
		aggregateResult: ingredientsCursor(ings),
	}
	svc := NewIngredientService(col)
	limit := 1
	cats := []string{"spice"}
	result, err := svc.Random(&limit, &cats)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 ingredient, got %d", len(result))
	}
}
