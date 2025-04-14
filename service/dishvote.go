package service

import (
	"context"
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

type DishVoteService struct{}

func (dvs *DishVoteService) Collection() *mongo.Collection {
	dbName := config.GetDBInstance().GetDbName()
	col := config.GetDBInstance().GetClient().Database(dbName).Collection(constants.DISH_VOTE_COLLECTION)
	return col
}

func (dvs *DishVoteService) Create(createDishVoteInput model.CreateDishVoteDto, profile *model.JwtCustomClaims) (*mongo.InsertOneResult, error) {
	collection := dvs.Collection()

	now := time.Now()

	dishVote := model.DishVote{
		Title:         createDishVoteInput.Title,
		Description:   createDishVoteInput.Description,
		DishVoteItems: createDishVoteInput.DishVoteItems,
		Deleted:       false,
		UpdatedAt:     &now,
		UpdatedBy:     &profile.ID,
		CreatedAt:     &now,
		CreatedBy:     &profile.ID,
	}

	result, err := collection.InsertOne(context.TODO(), dishVote)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (dvs *DishVoteService) Update(updateDishVoteInput model.UpdateDishVoteDto, profile *model.JwtCustomClaims) (*model.DishVote, error) {
	collection := dvs.Collection()

	now := time.Now()

	var dishVote model.DishVote

	filter := bson.M{"_id": updateDishVoteInput.ID, "deleted": false}
	options := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)
	result := collection.FindOneAndUpdate(context.TODO(), filter, bson.M{"$set": bson.D{
		{Key: "title", Value: updateDishVoteInput.Title},
		{Key: "description", Value: updateDishVoteInput.Description},
		{Key: "dishVoteItems", Value: updateDishVoteInput.DishVoteItems},
		{Key: "updatedAt", Value: now},
		{Key: "updatedBy", Value: profile.ID},
	}}, options)
	if result.Err() != nil {
		return nil, result.Err()
	}

	decodeErr := result.Decode(&dishVote)
	return &dishVote, decodeErr
}

func (dvs *DishVoteService) Remove(id string, profile *model.User) (*model.DishVote, error) {
	collection := dvs.Collection()
	now := time.Now()
	filter := bson.M{"_id": id, "deleted": false}
	options := options.FindOneAndUpdate().SetReturnDocument(options.After)
	result := collection.FindOneAndUpdate(context.TODO(), filter, bson.M{"$set": bson.M{
		"deleted":   true,
		"deletedAt": now,
		"deletedBy": profile.ID,
	}}, options)
	if result.Err() != nil {
		return nil, result.Err()
	}
	dishVote := model.DishVote{}
	decodeErr := result.Decode(&dishVote)
	return &dishVote, decodeErr
}

func (dvs *DishVoteService) Find(query model.QueryDishVoteDto) ([]*model.DishVote, int64, error) {
	collection := dvs.Collection()
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetSkip((int64(query.Page) - 1) * int64(query.Limit)).SetLimit(int64(query.Limit))
	filter := bson.D{{Key: "deleted", Value: false}}
	if query.Keyword != nil && *query.Keyword != "" {
		filter = append(filter, bson.E{Key: "$text", Value: bson.D{{Key: "$search", Value: *query.Keyword}}})
	}

	count, err := collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return nil, 0, err
	}

	cursor, err := collection.Find(context.TODO(), filter, opts)
	if err != nil {
		log.Println(err)
	}
	var dishVotes []*model.DishVote
	if err = cursor.All(context.TODO(), &dishVotes); err != nil {
		log.Println(err)
	}
	defer cursor.Close(context.TODO())
	return dishVotes, count, err
}

func (dvs *DishVoteService) FindOne(id string) (*model.DishVote, error) {
	collection := dvs.Collection()
	objID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return nil, err
	}
	filter := bson.M{"_id": objID}
	result := collection.FindOne(context.TODO(), filter)
	if result.Err() != nil {
		return nil, result.Err()
	}
	dishVote := model.DishVote{}
	decodeErr := result.Decode(&dishVote)
	return &dishVote, decodeErr
}
