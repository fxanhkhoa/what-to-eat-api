package service

import (
	"context"
	"time"
	"what-to-eat/be/graph/model"
	"what-to-eat/be/shared"

	"firebase.google.com/go/v4/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

func (u *UserService) FindUserByUID(googleID string) (*model.User, error) {
	_, collection := shared.Init(shared.DatabaseName, "Users")
	var user *model.User
	filter := bson.M{"googleID": googleID}
	result := collection.FindOne(context.TODO(), filter)
	if result.Err() != nil {
		return nil, result.Err()
	}
	err := result.Decode(&user)
	return user, err
}

func (u *UserService) CreateUserWithGoogle(queriedUser *auth.UserRecord) (*model.User, error) {
	_, collection := shared.Init(shared.DatabaseName, "Users")
	now := time.Now()
	user := model.User{
		Email:       queriedUser.Email,
		Name:        &queriedUser.DisplayName,
		DateOfBirth: &now,
		Phone:       &queriedUser.PhoneNumber,
		GoogleID:    &queriedUser.UID,
		Avatar:      &queriedUser.PhotoURL,
		Deleted:     false,
		UpdatedAt:   &now,
		CreatedAt:   &now,
		RoleName:    "USER",
	}

	filter := bson.M{"GoogleID": user.GoogleID, "deleted": true}
	options := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	result := collection.FindOneAndUpdate(context.TODO(), filter, bson.M{"$set": user}, options)
	if result.Err() != nil {
		return nil, result.Err()
	}
	decodeErr := result.Decode(&user)
	return &user, decodeErr
}

func (u *UserService) FindByID(id string) (*model.User, error) {
	_, collection := shared.Init(shared.DatabaseName, "Users")
	var user *model.User
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"_id": objectID}
	result := collection.FindOne(context.TODO(), filter)
	if result.Err() != nil {
		return nil, result.Err()
	}
	decodeErr := result.Decode(&user)
	return user, decodeErr
}
