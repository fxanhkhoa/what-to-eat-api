package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"
	"what-to-eat/be/config"
	constants "what-to-eat/be/constants"
	"what-to-eat/be/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IngredientService struct{}

func (is *IngredientService) Collection() *mongo.Collection {
	dbName := config.GetDBInstance().GetDbName()
	col := config.GetDBInstance().GetClient().Database(dbName).Collection(constants.INGREDIENT_COLLECTION)
	return col
}

func (is *IngredientService) Create(createIngredientInput model.CreateIngredientDto, profile *model.JwtCustomClaims) (*model.Ingredient, error) {
	collection := is.Collection()

	now := time.Now()

	ingredient := model.Ingredient{
		Slug:               createIngredientInput.Slug,
		Title:              createIngredientInput.Title,
		Measure:            createIngredientInput.Measure,
		Calories:           createIngredientInput.Calories,
		Carbohydrate:       createIngredientInput.Carbohydrate,
		Fat:                createIngredientInput.Fat,
		IngredientCategory: createIngredientInput.IngredientCategory,
		Weight:             createIngredientInput.Weight,
		Protein:            createIngredientInput.Protein,
		Cholesterol:        createIngredientInput.Cholesterol,
		Sodium:             createIngredientInput.Sodium,
		Images:             createIngredientInput.Images,
		Deleted:            false,
		UpdatedAt:          &now,
		UpdatedBy:          new(string),
		CreatedAt:          &now,
		CreatedBy:          new(string),
	}
	filter := bson.M{"slug": createIngredientInput.Slug, "deleted": true}
	options := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	result := collection.FindOneAndUpdate(context.TODO(), filter, bson.M{"$set": ingredient}, options)
	if result.Err() != nil {
		return nil, result.Err()
	}
	decodeErr := result.Decode(&ingredient)
	return &ingredient, decodeErr
}

func (is *IngredientService) Update(updateIngredientInput model.UpdateIngredientDto, profile *model.JwtCustomClaims) (*model.Ingredient, error) {
	collection := is.Collection()
	now := time.Now()
	var ingredient model.Ingredient

	objectID, err := primitive.ObjectIDFromHex(updateIngredientInput.ID)
	if err != nil {
		return nil, errors.New("invalid ID format")
	}

	filter := bson.M{"_id": objectID, "deleted": false}
	options := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)
	result := collection.FindOneAndUpdate(context.TODO(), filter, bson.M{"$set": bson.M{
		"slug":               updateIngredientInput.Slug,
		"title":              updateIngredientInput.Title,
		"measure":            updateIngredientInput.Measure,
		"calories":           updateIngredientInput.Calories,
		"carbohydrate":       updateIngredientInput.Carbohydrate,
		"fat":                updateIngredientInput.Fat,
		"ingredientCategory": updateIngredientInput.IngredientCategory,
		"weight":             updateIngredientInput.Weight,
		"protein":            updateIngredientInput.Protein,
		"cholesterol":        updateIngredientInput.Cholesterol,
		"sodium":             updateIngredientInput.Sodium,
		"images":             updateIngredientInput.Images,
		"updatedAt":          &now,
		"updatedBy":          &profile.ID,
	}}, options)

	if result.Err() != nil {
		return nil, result.Err()
	}
	decodeErr := result.Decode(&ingredient)
	return &ingredient, decodeErr
}

func (is *IngredientService) Remove(id string, profile *model.JwtCustomClaims) (*model.Ingredient, error) {
	collection := is.Collection()
	now := time.Now()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid ID format")
	}

	filter := bson.M{"_id": objectID, "deleted": false}
	options := options.FindOneAndUpdate().SetReturnDocument(options.After)
	result := collection.FindOneAndUpdate(context.TODO(), filter, bson.M{"$set": bson.M{
		"deleted":   true,
		"deletedAt": now,
		"deletedBy": profile.ID,
	}}, options)
	if result.Err() != nil {
		return nil, result.Err()
	}
	ingredient := model.Ingredient{}
	decodeErr := result.Decode(&ingredient)
	return &ingredient, decodeErr
}

func (is *IngredientService) Find(query model.QueryIngredientDto) ([]*model.Ingredient, int64, error) {
	collection := is.Collection()
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetSkip((int64(query.Page) - 1) * int64(query.Limit)).SetLimit(int64(query.Limit))
	filter := bson.D{{Key: "deleted", Value: false}}
	if query.Keyword != nil && *query.Keyword != "" {
		filter = append(filter, bson.E{Key: "$text", Value: bson.D{{Key: "$search", Value: *query.Keyword}}})
	}

	if query.IngredientCategory != nil && len(*query.IngredientCategory) > 0 {
		filter = append(filter, bson.E{Key: "ingredientCategory", Value: bson.D{{Key: "$in", Value: query.IngredientCategory}}})
	}

	count, err := collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return nil, 0, err
	}

	cursor, err := collection.Find(context.TODO(), filter, opts)
	if err != nil {
		log.Println(err)
	}
	var ingredients []*model.Ingredient
	if err = cursor.All(context.TODO(), &ingredients); err != nil {
		log.Println(err)
	}
	defer cursor.Close(context.TODO())
	return ingredients, count, err
}

func (is *IngredientService) FindOne(id string) (*model.Ingredient, error) {
	collection := is.Collection()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid ID format")
	}

	filter := bson.M{"_id": objectID}
	result := collection.FindOne(context.TODO(), filter)
	if result.Err() != nil {
		return nil, result.Err()
	}
	ingredient := model.Ingredient{}
	decodeErr := result.Decode(&ingredient)
	return &ingredient, decodeErr
}

func (is *IngredientService) FindOneBySlug(slug string) (*model.Ingredient, error) {
	collection := is.Collection()
	filter := bson.M{"slug": slug}
	result := collection.FindOne(context.TODO(), filter)
	if result.Err() != nil {
		return nil, result.Err()
	}
	ingredient := model.Ingredient{}
	decodeErr := result.Decode(&ingredient)
	return &ingredient, decodeErr
}

func (is *IngredientService) FindTitleByLang(title string, lang string) (*model.Ingredient, error) {
	collection := is.Collection()
	titleFilter := fmt.Sprintf("title.%s", lang)
	filter := bson.M{titleFilter: title}
	result := collection.FindOne(context.TODO(), filter)
	if result.Err() != nil {
		return nil, result.Err()
	}
	ingredient := model.Ingredient{}
	decodeErr := result.Decode(&ingredient)
	if decodeErr != nil {
		return nil, decodeErr
	}
	return &ingredient, decodeErr
}

func (is *IngredientService) Random(limit *int, ingredientCategory *[]string) ([]*model.Ingredient, error) {
	collection := is.Collection()
	stages := []bson.D{}
	stages = append(stages, bson.D{{Key: "$match", Value: bson.D{{Key: "deleted", Value: false}}}})
	if ingredientCategory != nil && len(*ingredientCategory) > 0 {
		stages = append(stages, bson.D{{Key: "$match", Value: bson.D{{Key: "ingredientCategory", Value: bson.D{{Key: "$in", Value: *ingredientCategory}}}}}})
	}
	stages = append(stages, bson.D{{Key: "$sample", Value: bson.D{{Key: "size", Value: int64(*limit)}}}})

	cursor, err := collection.Aggregate(context.TODO(), stages)
	if err != nil {
		return nil, err
	}

	var records []*model.Ingredient
	if err = cursor.All(context.TODO(), &records); err != nil {
		log.Println(err)
	}
	defer cursor.Close(context.TODO())
	return records, err
}
