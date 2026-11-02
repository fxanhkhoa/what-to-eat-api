package service

import (
	"context"
	"errors"
	"time"
	"what-to-eat/be/config"
	constants "what-to-eat/be/constants"
	"what-to-eat/be/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type FeedbackService struct {
	col CollectionInterface
}

// NewFeedbackService creates a FeedbackService with an injected collection (useful for testing).
func NewFeedbackService(col CollectionInterface) *FeedbackService {
	return &FeedbackService{col: col}
}

func (fs *FeedbackService) Collection() *mongo.Collection {
	dbName := config.GetDBInstance().GetDbName()
	col := config.GetDBInstance().GetClient().Database(dbName).Collection(constants.FEEDBACK_COLLECTION)
	return col
}

func (fs *FeedbackService) getCol() CollectionInterface {
	if fs.col != nil {
		return fs.col
	}
	return NewMongoCollectionAdapter(fs.Collection())
}

func (fs *FeedbackService) Create(createFeedbackInput model.CreateFeedbackDto, userId *primitive.ObjectID) (*model.Feedback, error) {
	collection := fs.getCol()

	now := time.Now()

	feedback := model.Feedback{
		UserID:    userId,
		UserName:  createFeedbackInput.UserName,
		Email:     createFeedbackInput.Email,
		Rating:    createFeedbackInput.Rating,
		Comment:   createFeedbackInput.Comment,
		Page:      createFeedbackInput.Page,
		UserAgent: createFeedbackInput.UserAgent,
		CreatedAt: now,
		UpdatedAt: now,
	}

	result, err := collection.InsertOne(context.TODO(), feedback)
	if err != nil {
		return nil, err
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		feedback.ID = oid
	}

	return &feedback, nil
}

func (fs *FeedbackService) GetAll(feedbackListDto model.FeedbackListDto) (*model.PaginationResponse, error) {
	collection := fs.getCol()

	filter := bson.M{}

	if feedbackListDto.Rating != nil {
		filter["rating"] = *feedbackListDto.Rating
	}

	if feedbackListDto.Email != nil && *feedbackListDto.Email != "" {
		filter["email"] = bson.M{"$regex": *feedbackListDto.Email, "$options": "i"}
	}

	page := feedbackListDto.Page
	limit := feedbackListDto.Limit

	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 10
	}

	skip := (page - 1) * limit

	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(limit))
	findOptions.SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := collection.Find(context.TODO(), filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var feedbacks []model.Feedback
	if err = cursor.All(context.TODO(), &feedbacks); err != nil {
		return nil, err
	}

	totalItems, err := collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	totalPages := float64(totalItems) / float64(limit)
	if totalPages < 1 {
		totalPages = 1
	}

	metadata := model.CountMetaData{
		TotalItems:   totalItems,
		ItemCount:    len(feedbacks),
		ItemsPerPage: limit,
		TotalPages:   totalPages,
		CurrentPage:  page,
	}

	return &model.PaginationResponse{
		Data:     feedbacks,
		Metadata: metadata,
	}, nil
}

func (fs *FeedbackService) GetById(id string) (*model.Feedback, error) {
	collection := fs.getCol()

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid feedback ID")
	}

	var feedback model.Feedback
	err = collection.FindOne(context.TODO(), bson.M{"_id": objectId}).Decode(&feedback)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("feedback not found")
		}
		return nil, err
	}

	return &feedback, nil
}

func (fs *FeedbackService) Update(id string, updateFeedbackDto model.UpdateFeedbackDto, userId primitive.ObjectID) (*model.Feedback, error) {
	collection := fs.getCol()

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid feedback ID")
	}

	// Check if feedback exists and belongs to user
	var existingFeedback model.Feedback
	err = collection.FindOne(context.TODO(), bson.M{"_id": objectId}).Decode(&existingFeedback)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("feedback not found")
		}
		return nil, err
	}

	// Verify user owns this feedback
	if existingFeedback.UserID == nil || *existingFeedback.UserID != userId {
		return nil, errors.New("unauthorized to update this feedback")
	}

	updateData := bson.M{
		"updatedAt": time.Now(),
	}

	if updateFeedbackDto.Rating != nil {
		updateData["rating"] = *updateFeedbackDto.Rating
	}

	if updateFeedbackDto.Comment != nil {
		updateData["comment"] = *updateFeedbackDto.Comment
	}

	update := bson.M{"$set": updateData}

	var updatedFeedback model.Feedback
	err = collection.FindOneAndUpdate(
		context.TODO(),
		bson.M{"_id": objectId},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updatedFeedback)

	if err != nil {
		return nil, err
	}

	return &updatedFeedback, nil
}

func (fs *FeedbackService) Delete(id string, userId primitive.ObjectID) error {
	collection := fs.getCol()

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid feedback ID")
	}

	// Check if feedback exists and belongs to user
	var existingFeedback model.Feedback
	err = collection.FindOne(context.TODO(), bson.M{"_id": objectId}).Decode(&existingFeedback)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("feedback not found")
		}
		return err
	}

	// Verify user owns this feedback
	if existingFeedback.UserID == nil || *existingFeedback.UserID != userId {
		return errors.New("unauthorized to delete this feedback")
	}

	result, err := collection.DeleteOne(context.TODO(), bson.M{"_id": objectId})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("feedback not found")
	}

	return nil
}
