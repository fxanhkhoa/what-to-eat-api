package service

import (
	"context"
	"log"
	"time"
	"what-to-eat/be/graph/model"
	"what-to-eat/be/shared"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IngredientService struct{}

func NewIngredientService() *IngredientService {
	return &IngredientService{}
}

func (is *IngredientService) Create(createIngredientInput model.CreateIngredientInput, profile *model.User) (*model.Ingredient, error) {
	_, collection := shared.Init("Ingredients")
	var title []*model.MultiLanguage
	for _, element := range createIngredientInput.Title {
		title = append(title, &model.MultiLanguage{Lang: element.Language, Data: element.Data})
	}
	now := time.Now()
	ingredient := model.Ingredient{
		Slug:               createIngredientInput.Slug,
		Title:              title,
		Measure:            createIngredientInput.Measure,
		Calories:           createIngredientInput.Calories,
		Carbohydrate:       createIngredientInput.Carbohydrate,
		Fat:                createIngredientInput.Fat,
		IngredientCategory: createIngredientInput.IngredientCategory,
		Weight:             createIngredientInput.Weight,
		Protein:            createIngredientInput.Protein,
		Cholesterol:        createIngredientInput.Cholesterol,
		Sodium:             createIngredientInput.Sodium,
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

func (is *IngredientService) Update(updateIngredientInput model.UpdateIngredientInput, profile *model.User) (*model.Ingredient, error) {
	_, collection := shared.Init("Ingredients")
	var title []*model.MultiLanguage
	for _, element := range updateIngredientInput.Title {
		title = append(title, &model.MultiLanguage{Lang: element.Language, Data: element.Data})
	}
	now := time.Now()
	ingredient := model.Ingredient{
		Slug:               updateIngredientInput.Slug,
		Title:              title,
		Measure:            updateIngredientInput.Measure,
		Calories:           updateIngredientInput.Calories,
		Carbohydrate:       updateIngredientInput.Carbohydrate,
		Fat:                updateIngredientInput.Fat,
		IngredientCategory: updateIngredientInput.IngredientCategory,
		Weight:             updateIngredientInput.Weight,
		Protein:            updateIngredientInput.Protein,
		Cholesterol:        updateIngredientInput.Cholesterol,
		Sodium:             updateIngredientInput.Sodium,
		UpdatedAt:          &now,
		UpdatedBy:          new(string),
	}
	filter := bson.M{"slug": updateIngredientInput.Slug, "deleted": false}
	options := options.FindOneAndUpdate().SetReturnDocument(options.After)
	result := collection.FindOneAndUpdate(context.TODO(), filter, bson.M{"$set": ingredient}, options)
	if result.Err() != nil {
		return nil, result.Err()
	}
	decodeErr := result.Decode(&ingredient)
	return &ingredient, decodeErr
}

func (is *IngredientService) Remove(slug string, profile *model.User) (*model.Ingredient, error) {
	_, collection := shared.Init("Ingredients")
	now := time.Now()
	filter := bson.M{"slug": slug, "deleted": false}
	options := options.FindOneAndUpdate().SetReturnDocument(options.After)
	result := collection.FindOneAndUpdate(context.TODO(), filter, bson.M{"$set": bson.M{
		"deleted":   true,
		"deletedAt": now,
		"deletedBy": "",
	}}, options)
	if result.Err() != nil {
		return nil, result.Err()
	}
	ingredient := model.Ingredient{}
	decodeErr := result.Decode(&ingredient)
	return &ingredient, decodeErr
}

func (is *IngredientService) Find(keyword *string, page *int, limit *int) ([]*model.Ingredient, error) {
	_, collection := shared.Init("Ingredients")
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetSkip((int64(*page) - 1) * int64(*limit)).SetLimit(int64(*limit))
	filter := bson.D{{Key: "deleted", Value: false}}
	if keyword != nil {
		filter = append(filter, bson.E{Key: "$text", Value: bson.D{{Key: "$search", Value: keyword}}})
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
	return ingredients, err
}

func (is *IngredientService) FindOne(slug string) (*model.Ingredient, error) {
	_, collection := shared.Init("Ingredients")
	filter := bson.M{"slug": slug}
	result := collection.FindOne(context.TODO(), filter)
	if result.Err() != nil {
		return nil, result.Err()
	}
	ingredient := model.Ingredient{}
	decodeErr := result.Decode(&ingredient)
	return &ingredient, decodeErr
}
