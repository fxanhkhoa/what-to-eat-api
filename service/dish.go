package service

import (
	"context"
	"log"
	"time"
	"what-to-eat/be/config"
	constants "what-to-eat/be/constants"
	"what-to-eat/be/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DishService struct{}

func (s *DishService) Collection() *mongo.Collection {
	dbName := config.GetDBInstance().GetDbName()
	col := config.GetDBInstance().GetClient().Database(dbName).Collection(constants.DISH_COLLECTION)
	return col
}

func (ds *DishService) Create(createDishInput model.CreateDishDto, profile *model.JwtCustomClaims) (*model.Dish, error) {
	collection := ds.Collection()

	now := time.Now()

	dish := model.Dish{
		Slug:                 createDishInput.Slug,
		Title:                createDishInput.Title,
		ShortDescription:     createDishInput.ShortDescription,
		Content:              createDishInput.Content,
		Tags:                 createDishInput.Tags,
		PreparationTime:      createDishInput.PreparationTime,
		CookingTime:          createDishInput.CookingTime,
		DifficultLevel:       createDishInput.DifficultLevel,
		MealCategories:       createDishInput.MealCategories,
		IngredientCategories: createDishInput.IngredientCategories,
		Thumbnail:            createDishInput.Thumbnail,
		Videos:               createDishInput.Videos,
		Ingredients:          createDishInput.Ingredients,
		RelatedDishes:        createDishInput.RelatedDishes,
		Labels:               createDishInput.Labels,
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

func (ds *DishService) Update(updateDishInput model.UpdateDishDto, profile *model.JwtCustomClaims) (*model.Dish, error) {
	collection := ds.Collection()

	now := time.Now()

	var dish model.Dish

	filter := bson.M{"slug": updateDishInput.Slug, "deleted": false}
	options := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)
	result := collection.FindOneAndUpdate(context.TODO(), filter, bson.M{"$set": bson.D{
		{Key: "slug", Value: updateDishInput.Slug},
		{Key: "title", Value: updateDishInput.Title},
		{Key: "shortDescription", Value: updateDishInput.ShortDescription},
		{Key: "content", Value: updateDishInput.Content},
		{Key: "tags", Value: updateDishInput.Tags},
		{Key: "preparationTime", Value: updateDishInput.PreparationTime},
		{Key: "cookingTime", Value: updateDishInput.CookingTime},
		{Key: "difficultLevel", Value: updateDishInput.DifficultLevel},
		{Key: "mealCategories", Value: updateDishInput.MealCategories},
		{Key: "ingredientCategories", Value: updateDishInput.IngredientCategories},
		{Key: "thumbnail", Value: updateDishInput.Thumbnail},
		{Key: "videos", Value: updateDishInput.Videos},
		{Key: "ingredients", Value: updateDishInput.Ingredients},
		{Key: "relatedDishes", Value: updateDishInput.RelatedDishes},
		{Key: "labels", Value: updateDishInput.Labels},
		{Key: "updatedAt", Value: now},
		{Key: "updatedBy", Value: profile.ID},
	}}, options)
	if result.Err() != nil {
		return nil, result.Err()
	}
	decodeErr := result.Decode(&dish)
	return &dish, decodeErr
}

func (ds *DishService) Remove(slug string, profile *model.User) (*model.Dish, error) {
	collection := ds.Collection()
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

func (ds *DishService) Find(query model.QueryDishDto) ([]*model.Dish, int64, error) {
	collection := ds.Collection()
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetSkip((int64(query.Page) - 1) * int64(query.Limit)).SetLimit(int64(query.Limit))
	filter := bson.D{{Key: "deleted", Value: false}}
	if query.Keyword != nil {
		filter = append(filter, bson.E{Key: "$text", Value: bson.D{{Key: "$search", Value: query.Keyword}}})
	}
	if query.Tags != nil && len(*query.Tags) > 0 {
		filter = append(filter, bson.E{Key: "tags", Value: bson.D{{Key: "$in", Value: query.Tags}}})
	}
	if query.PreparationTimeFrom != nil && query.PreparationTimeTo != nil {
		filter = append(filter, bson.E{Key: "preparationTime", Value: bson.D{{Key: "$lte", Value: query.PreparationTimeTo}, {Key: "$gte", Value: query.PreparationTimeFrom}}})
	}
	if query.CookingTimeFrom != nil && query.CookingTimeTo != nil {
		filter = append(filter, bson.E{Key: "cookingTime", Value: bson.D{{Key: "$lte", Value: query.CookingTimeTo}, {Key: "$gte", Value: query.CookingTimeFrom}}})
	}
	if query.DifficultLevels != nil && len(*query.DifficultLevels) > 0 {
		filter = append(filter, bson.E{Key: "difficultLevel", Value: bson.D{{Key: "$in", Value: query.DifficultLevels}}})
	}
	if query.MealCategories != nil && len(*query.MealCategories) > 0 {
		filter = append(filter, bson.E{Key: "mealCategories", Value: bson.D{{Key: "$in", Value: query.MealCategories}}})
	}
	if query.IngredientCategories != nil && len(*query.IngredientCategories) > 0 {
		filter = append(filter, bson.E{Key: "ingredientCategories", Value: bson.D{{Key: "$in", Value: query.IngredientCategories}}})
	}
	if query.Ingredients != nil && len(*query.Ingredients) > 0 {
		filter = append(filter, bson.E{Key: "ingredients.slug", Value: bson.D{{Key: "$in", Value: query.Ingredients}}})
	}
	if query.Labels != nil && len(*query.Labels) > 0 {
		filter = append(filter, bson.E{Key: "labels", Value: bson.D{{Key: "$in", Value: query.Labels}}})
	}

	count, err := collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return nil, 0, err
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
	return dishes, count, err
}

func (ds *DishService) FindOne(id string) (*model.Dish, error) {
	collection := ds.Collection()
	filter := bson.M{"_id": id}
	result := collection.FindOne(context.TODO(), filter)
	if result.Err() != nil {
		return nil, result.Err()
	}
	dish := model.Dish{}
	decodeErr := result.Decode(&dish)
	return &dish, decodeErr
}

func (ds *DishService) FindOneBySlug(slug string) (*model.Dish, error) {
	collection := ds.Collection()
	filter := bson.M{"slug": slug}
	result := collection.FindOne(context.TODO(), filter)
	if result.Err() != nil {
		return nil, result.Err()
	}
	dish := model.Dish{}
	decodeErr := result.Decode(&dish)
	return &dish, decodeErr
}

func (ds *DishService) Random(limit *int) ([]*model.Dish, error) {
	collection := ds.Collection()
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
