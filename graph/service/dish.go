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

type DishService struct{}

func NewDishService() *DishService {
	return &DishService{}
}

func (ds *DishService) Create(createDishInput model.CreateDishInput, profile *model.User) (*model.Dish, error) {
	_, collection := shared.Init("Dishes")

	var title []*model.MultiLanguageD
	for _, element := range createDishInput.Title {
		title = append(title, &model.MultiLanguageD{Lang: element.Lang, Data: element.Data})
	}

	var shortDescription []*model.MultiLanguageD
	for _, element := range createDishInput.ShortDescription {
		shortDescription = append(shortDescription, &model.MultiLanguageD{Lang: element.Lang, Data: element.Data})
	}

	var content []*model.MultiLanguageD
	for _, element := range createDishInput.Content {
		content = append(content, &model.MultiLanguageD{Lang: element.Lang, Data: element.Data})
	}

	var ingredients []*model.IngredientsInDish
	for _, element := range createDishInput.Ingredients {
		ingredients = append(ingredients, &model.IngredientsInDish{Quantity: element.Quantity, Slug: element.Slug, Note: element.Note})
	}

	now := time.Now()

	dish := model.Dish{
		Slug:                 createDishInput.Slug,
		Title:                title,
		ShortDescription:     shortDescription,
		Content:              content,
		Tags:                 createDishInput.Tags,
		PreparationTime:      createDishInput.PreparationTime,
		CookingTime:          createDishInput.CookingTime,
		DifficultLevel:       createDishInput.DifficultLevel,
		MealCategories:       createDishInput.MealCategories,
		IngredientCategories: createDishInput.IngredientCategories,
		Thumbnail:            createDishInput.Thumbnail,
		Videos:               createDishInput.Videos,
		Ingredients:          ingredients,
		RelatedDishes:        createDishInput.RelatedDishes,
		Deleted:              false,
		UpdatedAt:            &now,
		UpdatedBy:            &profile.ID,
		CreatedAt:            &now,
		CreatedBy:            &profile.ID,
	}

	filter := bson.M{"slug": createDishInput.Slug, "deleted": true}
	options := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)
	result := collection.FindOneAndUpdate(context.TODO(), filter, bson.M{"$set": dish}, options)
	if result.Err() != nil {
		return nil, result.Err()
	}
	decodeErr := result.Decode(&dish)
	return &dish, decodeErr
}

func (ds *DishService) Update(updateDishInput model.UpdateDishInput, profile *model.User) (*model.Dish, error) {
	_, collection := shared.Init("Dishes")

	var title []*model.MultiLanguageD
	for _, element := range updateDishInput.Title {
		title = append(title, &model.MultiLanguageD{Lang: element.Lang, Data: element.Data})
	}

	var shortDescription []*model.MultiLanguageD
	for _, element := range updateDishInput.ShortDescription {
		shortDescription = append(shortDescription, &model.MultiLanguageD{Lang: element.Lang, Data: element.Data})
	}

	var content []*model.MultiLanguageD
	for _, element := range updateDishInput.Content {
		content = append(content, &model.MultiLanguageD{Lang: element.Lang, Data: element.Data})
	}

	var ingredients []*model.IngredientsInDish
	for _, element := range updateDishInput.Ingredients {
		ingredients = append(ingredients, &model.IngredientsInDish{Quantity: element.Quantity, Slug: element.Slug, Note: element.Note})
	}

	now := time.Now()

	dish := model.Dish{
		Slug:                 updateDishInput.Slug,
		Title:                title,
		ShortDescription:     shortDescription,
		Content:              content,
		Tags:                 updateDishInput.Tags,
		PreparationTime:      updateDishInput.PreparationTime,
		CookingTime:          updateDishInput.CookingTime,
		DifficultLevel:       updateDishInput.DifficultLevel,
		MealCategories:       updateDishInput.MealCategories,
		IngredientCategories: updateDishInput.IngredientCategories,
		Thumbnail:            updateDishInput.Thumbnail,
		Videos:               updateDishInput.Videos,
		Ingredients:          ingredients,
		RelatedDishes:        updateDishInput.RelatedDishes,
		UpdatedAt:            &now,
		UpdatedBy:            &profile.ID,
	}

	filter := bson.M{"slug": updateDishInput.Slug, "deleted": false}
	options := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)
	result := collection.FindOneAndUpdate(context.TODO(), filter, bson.M{"$set": dish}, options)
	if result.Err() != nil {
		return nil, result.Err()
	}
	decodeErr := result.Decode(&dish)
	return &dish, decodeErr
}

func (ds *DishService) Remove(slug string, profile *model.User) (*model.Dish, error) {
	_, collection := shared.Init("Dishes")
	now := time.Now()
	filter := bson.M{"slug": slug, "deleted": false}
	options := options.FindOneAndUpdate().SetReturnDocument(options.After)
	result := collection.FindOneAndUpdate(context.TODO(), filter, bson.M{"$set": bson.M{
		"deleted":   true,
		"deletedAt": now,
		"deletedBy": profile.ID,
	}}, options)
	if result.Err() != nil {
		return nil, result.Err()
	}
	dish := model.Dish{}
	decodeErr := result.Decode(&dish)
	return &dish, decodeErr
}

func (ds *DishService) Find(keyword *string, page *int, limit *int) ([]*model.Dish, error) {
	_, collection := shared.Init("Dishes")
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetSkip((int64(*page) - 1) * int64(*limit)).SetLimit(int64(*limit))
	filter := bson.D{{Key: "deleted", Value: false}}
	if keyword != nil {
		filter = append(filter, bson.E{Key: "$text", Value: bson.D{{Key: "$search", Value: keyword}}})
	}
	cursor, err := collection.Find(context.TODO(), filter, opts)
	if err != nil {
		log.Println(err)
	}
	var dishes []*model.Dish
	if err = cursor.All(context.TODO(), &dishes); err != nil {
		log.Println(err)
	}
	defer cursor.Close(context.TODO())
	return dishes, err
}

func (ds *DishService) FindOne(slug string) (*model.Dish, error) {
	_, collection := shared.Init("Dishes")
	filter := bson.M{"slug": slug}
	result := collection.FindOne(context.TODO(), filter)
	if result.Err() != nil {
		return nil, result.Err()
	}
	dish := model.Dish{}
	decodeErr := result.Decode(&dish)
	return &dish, decodeErr
}

func (ds *DishService) Count(keyword *string) (int64, error) {
	_, collection := shared.Init("Dishes")
	filter := bson.D{{Key: "deleted", Value: false}}
	if keyword != nil {
		filter = append(filter, bson.E{Key: "$text", Value: bson.D{{Key: "$search", Value: keyword}}})
	}
	total, err := collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return 0, err
	}

	return total, err
}

func (ds *DishService) Random(limit *int) ([]*model.Dish, error) {
	_, collection := shared.Init("Dishes")
	stages := []bson.D{}
	stages = append(stages, bson.D{{Key: "$match", Value: bson.D{{Key: "deleted", Value: false}}}})
	stages = append(stages, bson.D{{Key: "$sample", Value: bson.D{{Key: "size", Value: int64(*limit)}}}})
	cursor, err := collection.Aggregate(context.TODO(), stages)
	if err != nil {
		return nil, err
	}

	var dishes []*model.Dish
	if err = cursor.All(context.TODO(), &dishes); err != nil {
		log.Println(err)
	}
	defer cursor.Close(context.TODO())
	return dishes, err
}
