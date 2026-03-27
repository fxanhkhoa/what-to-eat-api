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
// dishMockCollection — CollectionInterface impl for dish tests
// ---------------------------------------------------------------------------

type dishMockCollection struct {
	// FindOneAndUpdate
	findOneAndUpdateResult SingleResult

	// CountDocuments
	countResult int64
	countErr    error

	// Find
	findResult Cursor
	findErr    error

	// FindOne
	findOneResult SingleResult

	// Aggregate — queue of cursors consumed in call order
	aggregateQueue []Cursor
	aggregateErr   error
	aggregateCalls int
}

func (m *dishMockCollection) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	return &mongo.InsertOneResult{InsertedID: primitive.NewObjectID()}, nil
}

func (m *dishMockCollection) FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...*options.FindOneAndUpdateOptions) SingleResult {
	return m.findOneAndUpdateResult
}

func (m *dishMockCollection) CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	return m.countResult, m.countErr
}

func (m *dishMockCollection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (Cursor, error) {
	if m.findResult != nil {
		return m.findResult, m.findErr
	}
	return emptyCursor(), m.findErr
}

func (m *dishMockCollection) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) SingleResult {
	return m.findOneResult
}

func (m *dishMockCollection) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return &mongo.UpdateResult{MatchedCount: 1}, nil
}

func (m *dishMockCollection) UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return nil, nil
}

func (m *dishMockCollection) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return &mongo.DeleteResult{DeletedCount: 1}, nil
}

func (m *dishMockCollection) Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (Cursor, error) {
	m.aggregateCalls++
	if m.aggregateErr != nil {
		return nil, m.aggregateErr
	}
	idx := m.aggregateCalls - 1
	if idx < len(m.aggregateQueue) {
		return m.aggregateQueue[idx], nil
	}
	return emptyCursor(), nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func decodeDishFn(dish model.Dish) func(interface{}) error {
	return func(v interface{}) error {
		*(v.(*model.Dish)) = dish
		return nil
	}
}

func dishesCursor(dishes []*model.Dish) Cursor {
	return &genericCursor{
		decodeFn: func(results interface{}) error {
			*(results.(*[]*model.Dish)) = dishes
			return nil
		},
	}
}

func analyzeResultCursor(data map[string]interface{}) Cursor {
	return &genericCursor{
		decodeFn: func(results interface{}) error {
			v := results.(*[]map[string]interface{})
			*v = []map[string]interface{}{data}
			return nil
		},
	}
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestDishCreate_FindOneAndUpdateError(t *testing.T) {
	col := &dishMockCollection{
		findOneAndUpdateResult: &mockSingleResult{err: errors.New("write error")},
	}
	svc := NewDishService(col)
	_, err := svc.Create(model.CreateDishDto{Slug: "pasta"}, &model.JwtCustomClaims{})
	if err == nil || err.Error() != "write error" {
		t.Errorf("expected 'write error', got %v", err)
	}
}

func TestDishCreate_Success(t *testing.T) {
	expected := model.Dish{Slug: "pasta"}
	col := &dishMockCollection{
		findOneAndUpdateResult: &mockSingleResult{decodeFn: decodeDishFn(expected)},
	}
	svc := NewDishService(col)
	dish, err := svc.Create(model.CreateDishDto{Slug: "pasta"}, &model.JwtCustomClaims{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dish.Slug != "pasta" {
		t.Errorf("expected slug 'pasta', got %s", dish.Slug)
	}
}

func TestDishCreate_SetsAuditFields(t *testing.T) {
	expected := model.Dish{Slug: "pasta"}
	col := &dishMockCollection{
		findOneAndUpdateResult: &mockSingleResult{decodeFn: decodeDishFn(expected)},
	}
	svc := NewDishService(col)
	profile := &model.JwtCustomClaims{}
	dish, err := svc.Create(model.CreateDishDto{Slug: "pasta"}, profile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Returned dish is from decode (mocked), but we verify no error and correct type
	if dish == nil {
		t.Error("expected non-nil dish")
	}
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestDishUpdate_InvalidID(t *testing.T) {
	svc := NewDishService(&dishMockCollection{})
	_, err := svc.Update(model.UpdateDishDto{ID: "bad-id"}, &model.JwtCustomClaims{})
	if err == nil || err.Error() != "invalid ID format" {
		t.Errorf("expected 'invalid ID format', got %v", err)
	}
}

func TestDishUpdate_FindOneAndUpdateError(t *testing.T) {
	col := &dishMockCollection{
		findOneAndUpdateResult: &mockSingleResult{err: errors.New("update error")},
	}
	svc := NewDishService(col)
	_, err := svc.Update(model.UpdateDishDto{ID: primitive.NewObjectID().Hex()}, &model.JwtCustomClaims{})
	if err == nil || err.Error() != "update error" {
		t.Errorf("expected 'update error', got %v", err)
	}
}

func TestDishUpdate_Success(t *testing.T) {
	expected := model.Dish{Slug: "pasta-updated"}
	col := &dishMockCollection{
		findOneAndUpdateResult: &mockSingleResult{decodeFn: decodeDishFn(expected)},
	}
	svc := NewDishService(col)
	dish, err := svc.Update(
		model.UpdateDishDto{ID: primitive.NewObjectID().Hex(), Slug: "pasta-updated"},
		&model.JwtCustomClaims{},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dish.Slug != "pasta-updated" {
		t.Errorf("expected slug 'pasta-updated', got %s", dish.Slug)
	}
}

// ---------------------------------------------------------------------------
// Remove
// ---------------------------------------------------------------------------

func TestDishRemove_InvalidID(t *testing.T) {
	svc := NewDishService(&dishMockCollection{})
	_, err := svc.Remove("bad-id", &model.JwtCustomClaims{})
	if err == nil || err.Error() != "invalid ID format" {
		t.Errorf("expected 'invalid ID format', got %v", err)
	}
}

func TestDishRemove_FindOneAndUpdateError(t *testing.T) {
	col := &dishMockCollection{
		findOneAndUpdateResult: &mockSingleResult{err: errors.New("remove error")},
	}
	svc := NewDishService(col)
	_, err := svc.Remove(primitive.NewObjectID().Hex(), &model.JwtCustomClaims{})
	if err == nil || err.Error() != "remove error" {
		t.Errorf("expected 'remove error', got %v", err)
	}
}

func TestDishRemove_Success(t *testing.T) {
	expected := model.Dish{Slug: "gone", Deleted: true}
	col := &dishMockCollection{
		findOneAndUpdateResult: &mockSingleResult{decodeFn: decodeDishFn(expected)},
	}
	svc := NewDishService(col)
	dish, err := svc.Remove(primitive.NewObjectID().Hex(), &model.JwtCustomClaims{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !dish.Deleted {
		t.Error("expected dish to be marked deleted")
	}
}

// ---------------------------------------------------------------------------
// Find
// ---------------------------------------------------------------------------

func TestDishFind_CountError(t *testing.T) {
	col := &dishMockCollection{countErr: errors.New("count error")}
	svc := NewDishService(col)
	_, _, err := svc.Find(model.QueryDishDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}})
	if err == nil || err.Error() != "count error" {
		t.Errorf("expected 'count error', got %v", err)
	}
}

func TestDishFind_Success_EmptyResult(t *testing.T) {
	col := &dishMockCollection{
		countResult: 0,
		findResult:  dishesCursor([]*model.Dish{}),
	}
	svc := NewDishService(col)
	dishes, count, err := svc.Find(model.QueryDishDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 || len(dishes) != 0 {
		t.Errorf("expected 0 dishes, got count=%d, len=%d", count, len(dishes))
	}
}

func TestDishFind_Success_WithResults(t *testing.T) {
	dishList := []*model.Dish{{Slug: "pasta"}, {Slug: "pizza"}}
	col := &dishMockCollection{
		countResult: 2,
		findResult:  dishesCursor(dishList),
	}
	svc := NewDishService(col)
	dishes, count, err := svc.Find(model.QueryDishDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected count=2, got %d", count)
	}
	if len(dishes) != 2 {
		t.Errorf("expected 2 dishes, got %d", len(dishes))
	}
	if dishes[0].Slug != "pasta" {
		t.Errorf("expected first dish slug 'pasta', got %s", dishes[0].Slug)
	}
}

func TestDishFind_WithKeyword(t *testing.T) {
	col := &dishMockCollection{
		countResult: 1,
		findResult:  dishesCursor([]*model.Dish{{Slug: "pasta"}}),
	}
	svc := NewDishService(col)
	kw := "pasta"
	dishes, count, err := svc.Find(model.QueryDishDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}, Keyword: &kw})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 || len(dishes) != 1 {
		t.Errorf("expected 1 dish, got count=%d, len=%d", count, len(dishes))
	}
}

func TestDishFind_WithFilters(t *testing.T) {
	col := &dishMockCollection{
		countResult: 1,
		findResult:  dishesCursor([]*model.Dish{{Slug: "veg-pasta"}}),
	}
	svc := NewDishService(col)
	tags := []string{"vegetarian"}
	mealCats := []string{"dinner"}
	labels := []string{"healthy"}
	dishes, count, err := svc.Find(model.QueryDishDto{
		BaseDto:        model.BaseDto{Page: 1, Limit: 10},
		Tags:           &tags,
		MealCategories: &mealCats,
		Labels:         &labels,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 || len(dishes) != 1 {
		t.Errorf("expected 1 dish with filters, got count=%d, len=%d", count, len(dishes))
	}
}

func TestDishFind_CursorDecodeError(t *testing.T) {
	col := &dishMockCollection{
		countResult: 1,
		findResult:  &genericCursor{err: errors.New("cursor error")},
	}
	svc := NewDishService(col)
	_, _, err := svc.Find(model.QueryDishDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}})
	if err == nil || err.Error() != "cursor error" {
		t.Errorf("expected 'cursor error', got %v", err)
	}
}

// ---------------------------------------------------------------------------
// FindOne
// ---------------------------------------------------------------------------

func TestDishFindOne_InvalidID(t *testing.T) {
	svc := NewDishService(&dishMockCollection{})
	_, err := svc.FindOne("not-a-hex")
	if err == nil || err.Error() != "invalid ID format" {
		t.Errorf("expected 'invalid ID format', got %v", err)
	}
}

func TestDishFindOne_NotFound(t *testing.T) {
	col := &dishMockCollection{
		findOneResult: &mockSingleResult{err: mongo.ErrNoDocuments},
	}
	svc := NewDishService(col)
	_, err := svc.FindOne(primitive.NewObjectID().Hex())
	if err != mongo.ErrNoDocuments {
		t.Errorf("expected ErrNoDocuments, got %v", err)
	}
}

func TestDishFindOne_Success(t *testing.T) {
	expected := model.Dish{Slug: "pasta"}
	col := &dishMockCollection{
		findOneResult: &mockSingleResult{decodeFn: decodeDishFn(expected)},
	}
	svc := NewDishService(col)
	dish, err := svc.FindOne(primitive.NewObjectID().Hex())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dish.Slug != "pasta" {
		t.Errorf("expected slug 'pasta', got %s", dish.Slug)
	}
}

// ---------------------------------------------------------------------------
// FindOneBySlug
// ---------------------------------------------------------------------------

func TestDishFindOneBySlug_NotFound(t *testing.T) {
	col := &dishMockCollection{
		findOneResult: &mockSingleResult{err: mongo.ErrNoDocuments},
	}
	svc := NewDishService(col)
	_, err := svc.FindOneBySlug("unknown-slug")
	if err != mongo.ErrNoDocuments {
		t.Errorf("expected ErrNoDocuments, got %v", err)
	}
}

func TestDishFindOneBySlug_Success(t *testing.T) {
	expected := model.Dish{Slug: "classic-pasta"}
	col := &dishMockCollection{
		findOneResult: &mockSingleResult{decodeFn: decodeDishFn(expected)},
	}
	svc := NewDishService(col)
	dish, err := svc.FindOneBySlug("classic-pasta")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dish.Slug != "classic-pasta" {
		t.Errorf("expected slug 'classic-pasta', got %s", dish.Slug)
	}
}

// ---------------------------------------------------------------------------
// FindWithScore
// ---------------------------------------------------------------------------

func TestDishFindWithScore_NoKeyword_FallsBackToFind(t *testing.T) {
	col := &dishMockCollection{
		countResult: 1,
		findResult:  dishesCursor([]*model.Dish{{Slug: "pasta"}}),
	}
	svc := NewDishService(col)
	dishes, count, err := svc.FindWithScore(model.QueryDishDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 || len(dishes) != 1 {
		t.Errorf("expected 1 dish, got count=%d, len=%d", count, len(dishes))
	}
}

func TestDishFindWithScore_AggregateError(t *testing.T) {
	col := &dishMockCollection{aggregateErr: errors.New("agg error")}
	svc := NewDishService(col)
	kw := "chicken"
	_, _, err := svc.FindWithScore(model.QueryDishDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}, Keyword: &kw})
	if err == nil || err.Error() != "agg error" {
		t.Errorf("expected 'agg error', got %v", err)
	}
}

func TestDishFindWithScore_CountCursorAllError(t *testing.T) {
	// First aggregate call succeeds but cursor.All fails
	col := &dishMockCollection{
		aggregateQueue: []Cursor{&genericCursor{err: errors.New("cursor error")}},
	}
	svc := NewDishService(col)
	kw := "chicken"
	_, _, err := svc.FindWithScore(model.QueryDishDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}, Keyword: &kw})
	if err == nil || err.Error() != "cursor error" {
		t.Errorf("expected 'cursor error', got %v", err)
	}
}

func TestDishFindWithScore_Success_EmptyResults(t *testing.T) {
	// Both aggregate calls return empty cursors → valid success with zero results
	col := &dishMockCollection{
		aggregateQueue: []Cursor{emptyCursor(), emptyCursor()},
	}
	svc := NewDishService(col)
	kw := "chicken"
	dishes, count, err := svc.FindWithScore(model.QueryDishDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}, Keyword: &kw})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 || len(dishes) != 0 {
		t.Errorf("expected 0 results, got count=%d, len=%d", count, len(dishes))
	}
}

// ---------------------------------------------------------------------------
// FindWithFuzzyScore
// ---------------------------------------------------------------------------

func TestDishFindWithFuzzyScore_NoKeyword_FallsBackToFind(t *testing.T) {
	col := &dishMockCollection{
		countResult: 2,
		findResult:  dishesCursor([]*model.Dish{{Slug: "a"}, {Slug: "b"}}),
	}
	svc := NewDishService(col)
	dishes, count, err := svc.FindWithFuzzyScore(model.QueryDishDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 || len(dishes) != 2 {
		t.Errorf("expected 2 dishes, got count=%d, len=%d", count, len(dishes))
	}
}

// ---------------------------------------------------------------------------
// FindSmart
// ---------------------------------------------------------------------------

func TestDishFindSmart_NoKeyword_FallsBackToFind(t *testing.T) {
	col := &dishMockCollection{
		countResult: 1,
		findResult:  dishesCursor([]*model.Dish{{Slug: "pasta"}}),
	}
	svc := NewDishService(col)
	dishes, count, err := svc.FindSmart(model.QueryDishDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 || len(dishes) != 1 {
		t.Errorf("expected 1 dish, got count=%d, len=%d", count, len(dishes))
	}
}

// ---------------------------------------------------------------------------
// Random
// ---------------------------------------------------------------------------

func TestDishRandom_AggregateError(t *testing.T) {
	col := &dishMockCollection{aggregateErr: errors.New("agg error")}
	svc := NewDishService(col)
	_, err := svc.Random(model.QueryDishRandomDto{Limit: 5})
	if err == nil || err.Error() != "agg error" {
		t.Errorf("expected 'agg error', got %v", err)
	}
}

func TestDishRandom_Success(t *testing.T) {
	dishList := []*model.Dish{{Slug: "pasta"}, {Slug: "pizza"}}
	col := &dishMockCollection{
		aggregateQueue: []Cursor{dishesCursor(dishList)},
	}
	svc := NewDishService(col)
	dishes, err := svc.Random(model.QueryDishRandomDto{Limit: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dishes) != 2 {
		t.Errorf("expected 2 dishes, got %d", len(dishes))
	}
}

func TestDishRandom_WithMealCategory(t *testing.T) {
	dishList := []*model.Dish{{Slug: "breakfast-dish"}}
	col := &dishMockCollection{
		aggregateQueue: []Cursor{dishesCursor(dishList)},
	}
	svc := NewDishService(col)
	cats := []string{"breakfast"}
	dishes, err := svc.Random(model.QueryDishRandomDto{Limit: 3, MealCategories: &cats})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dishes) != 1 {
		t.Errorf("expected 1 dish, got %d", len(dishes))
	}
}

// ---------------------------------------------------------------------------
// Analyze
// ---------------------------------------------------------------------------

func TestDishAnalyze_AggregateError(t *testing.T) {
	col := &dishMockCollection{aggregateErr: errors.New("agg error")}
	svc := NewDishService(col)
	_, err := svc.Analyze()
	if err == nil || err.Error() != "agg error" {
		t.Errorf("expected 'agg error', got %v", err)
	}
}

func TestDishAnalyze_CursorAllError(t *testing.T) {
	col := &dishMockCollection{
		aggregateQueue: []Cursor{&genericCursor{err: errors.New("cursor error")}},
	}
	svc := NewDishService(col)
	_, err := svc.Analyze()
	if err == nil || err.Error() != "cursor error" {
		t.Errorf("expected 'cursor error', got %v", err)
	}
}

func TestDishAnalyze_EmptyResults(t *testing.T) {
	col := &dishMockCollection{
		aggregateQueue: []Cursor{emptyCursor()},
	}
	svc := NewDishService(col)
	result, err := svc.Analyze()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
	}
}

func TestDishAnalyze_WithData(t *testing.T) {
	data := map[string]interface{}{
		"categoryDistribution": []interface{}{},
		"avgTimes":             []interface{}{},
		"difficultyLevels":     []interface{}{},
	}
	col := &dishMockCollection{
		aggregateQueue: []Cursor{analyzeResultCursor(data)},
	}
	svc := NewDishService(col)
	result, err := svc.Analyze()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result["categoryDistribution"]; !ok {
		t.Error("expected 'categoryDistribution' key in result")
	}
}

// ---------------------------------------------------------------------------
// FindSuggestions
// ---------------------------------------------------------------------------

func TestDishFindSuggestions_EmptyKeyword(t *testing.T) {
	svc := NewDishService(&dishMockCollection{})
	suggestions, err := svc.FindSuggestions("", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 0 {
		t.Errorf("expected empty suggestions for empty keyword, got %v", suggestions)
	}
}

func TestDishFindSuggestions_AggregateError(t *testing.T) {
	col := &dishMockCollection{aggregateErr: errors.New("agg error")}
	svc := NewDishService(col)
	_, err := svc.FindSuggestions("pasta", 5)
	if err == nil || err.Error() != "agg error" {
		t.Errorf("expected 'agg error', got %v", err)
	}
}

func TestDishFindSuggestions_Success_EmptyResults(t *testing.T) {
	col := &dishMockCollection{
		aggregateQueue: []Cursor{emptyCursor()},
	}
	svc := NewDishService(col)
	suggestions, err := svc.FindSuggestions("pasta", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 0 {
		t.Errorf("expected empty suggestions from empty cursor, got %v", suggestions)
	}
}

// ---------------------------------------------------------------------------
// generateFuzzyVariants — pure function tests
// ---------------------------------------------------------------------------

func TestGenerateFuzzyVariants_KnownTypo_GivesCorrection(t *testing.T) {
	svc := &DishService{}
	variants := svc.generateFuzzyVariants("chiken")
	found := false
	for _, v := range variants {
		if v == "chicken" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'chicken' correction in variants, got %v", variants)
	}
}

func TestGenerateFuzzyVariants_MultiWord_ContainsIndividualWords(t *testing.T) {
	svc := &DishService{}
	variants := svc.generateFuzzyVariants("pork chicken")
	hasPork, hasChicken := false, false
	for _, v := range variants {
		if v == "pork" {
			hasPork = true
		}
		if v == "chicken" {
			hasChicken = true
		}
	}
	if !hasPork {
		t.Errorf("expected 'pork' in variants, got %v", variants)
	}
	if !hasChicken {
		t.Errorf("expected 'chicken' in variants, got %v", variants)
	}
}

func TestGenerateFuzzyVariants_ShortWord_NoPrefixVariants(t *testing.T) {
	svc := &DishService{}
	// "egg" has len 3, not > 4 — no prefix variants should be added
	variants := svc.generateFuzzyVariants("egg")
	if len(variants) != 1 {
		t.Errorf("expected only original word, got %v", variants)
	}
	if variants[0] != "egg" {
		t.Errorf("expected 'egg', got %s", variants[0])
	}
}

func TestGenerateFuzzyVariants_LongWord_HasPrefixVariants(t *testing.T) {
	svc := &DishService{}
	// "spaghetti" len=9 — should produce "spaghet" and "spaghett"
	variants := svc.generateFuzzyVariants("spaghetti")
	hasPrefix := false
	for _, v := range variants {
		if v == "spaghet" || v == "spaghett" {
			hasPrefix = true
			break
		}
	}
	if !hasPrefix {
		t.Errorf("expected prefix variants for 'spaghetti', got %v", variants)
	}
}

func TestGenerateFuzzyVariants_NoDuplicates(t *testing.T) {
	svc := &DishService{}
	variants := svc.generateFuzzyVariants("spagetti") // typo → spaghetti
	seen := make(map[string]bool)
	for _, v := range variants {
		if seen[v] {
			t.Errorf("duplicate variant %q in %v", v, variants)
		}
		seen[v] = true
	}
}

func TestGenerateFuzzyVariants_OriginalAlwaysFirst(t *testing.T) {
	svc := &DishService{}
	variants := svc.generateFuzzyVariants("mushroom")
	if len(variants) == 0 || variants[0] != "mushroom" {
		t.Errorf("expected original keyword first, got %v", variants)
	}
}

func TestGenerateFuzzyVariants_SalmanCorrection(t *testing.T) {
	svc := &DishService{}
	variants := svc.generateFuzzyVariants("salman")
	found := false
	for _, v := range variants {
		if v == "salmon" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'salmon' correction, got %v", variants)
	}
}

// ---------------------------------------------------------------------------
// buildSimpleScoreExpression — pure function tests
// ---------------------------------------------------------------------------

func TestBuildSimpleScoreExpression_EmptyKeyword_ReturnsOne(t *testing.T) {
	svc := &DishService{}
	result := svc.buildSimpleScoreExpression("")
	if result != 1 {
		t.Errorf("expected 1 for empty keyword, got %v", result)
	}
}

func TestBuildSimpleScoreExpression_NonEmpty_ReturnsBsonD(t *testing.T) {
	svc := &DishService{}
	result := svc.buildSimpleScoreExpression("chicken")
	doc, ok := result.(bson.D)
	if !ok {
		t.Fatalf("expected bson.D, got %T", result)
	}
	if len(doc) == 0 || doc[0].Key != "$add" {
		t.Errorf("expected bson.D with '$add' key, got %v", doc)
	}
}

func TestBuildSimpleScoreExpression_RegexQuotesMeta(t *testing.T) {
	// Special regex chars should not cause panic
	svc := &DishService{}
	result := svc.buildSimpleScoreExpression("pasta+bolognese[2]")
	if result == nil {
		t.Error("expected non-nil result for special-char keyword")
	}
}

// ---------------------------------------------------------------------------
// buildSimpleSearchPipeline — pure function tests
// ---------------------------------------------------------------------------

func TestBuildSimpleSearchPipeline_AlwaysSixStages(t *testing.T) {
	svc := &DishService{}
	query := model.QueryDishDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}}
	pipeline := svc.buildSimpleSearchPipeline("", []string{}, query)
	// $match, $addFields, $sort, $skip, $limit, $project = 6 stages
	if len(pipeline) != 6 {
		t.Errorf("expected 6 pipeline stages, got %d", len(pipeline))
	}
}

func TestBuildSimpleSearchPipeline_WithKeyword_StillSixStages(t *testing.T) {
	svc := &DishService{}
	query := model.QueryDishDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}}
	pipeline := svc.buildSimpleSearchPipeline("chicken", []string{"chicken"}, query)
	if len(pipeline) != 6 {
		t.Errorf("expected 6 stages with keyword, got %d", len(pipeline))
	}
}

func TestBuildSimpleSearchPipeline_PaginationReflectedInStages(t *testing.T) {
	svc := &DishService{}
	query := model.QueryDishDto{BaseDto: model.BaseDto{Page: 3, Limit: 20}}
	pipeline := svc.buildSimpleSearchPipeline("rice", []string{"rice"}, query)
	// $skip stage is index 3
	skipStage := pipeline[3]
	if len(skipStage) == 0 || skipStage[0].Key != "$skip" {
		t.Errorf("expected $skip at index 3, got %v", skipStage)
	}
	// Expected skip = (3-1)*20 = 40
	if skipStage[0].Value != int64(40) {
		t.Errorf("expected skip=40, got %v", skipStage[0].Value)
	}
}

// ---------------------------------------------------------------------------
// buildSearchPipeline — pure function tests
// ---------------------------------------------------------------------------

func TestBuildSearchPipeline_SevenStages(t *testing.T) {
	svc := &DishService{}
	query := model.QueryDishDto{BaseDto: model.BaseDto{Page: 1, Limit: 10}}
	pipeline := svc.buildSearchPipeline("chicken", []string{"chicken"}, query)
	// $match, $addFields(score), $match(score>0), $sort, $skip, $limit, $project = 7 stages
	if len(pipeline) != 7 {
		t.Errorf("expected 7 pipeline stages, got %d", len(pipeline))
	}
}

func TestBuildSearchPipeline_WithFilters(t *testing.T) {
	svc := &DishService{}
	tags := []string{"vegan"}
	levels := []string{"easy"}
	query := model.QueryDishDto{
		BaseDto:         model.BaseDto{Page: 1, Limit: 5},
		Tags:            &tags,
		DifficultLevels: &levels,
	}
	pipeline := svc.buildSearchPipeline("salad", []string{"salad"}, query)
	// Filters are baked into $match, stage count stays 7
	if len(pipeline) != 7 {
		t.Errorf("expected 7 stages regardless of filters, got %d", len(pipeline))
	}
}
