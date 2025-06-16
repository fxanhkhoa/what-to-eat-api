package service

import (
	"context"
	"log"
	"time"
	"what-to-eat/be/config"
	"what-to-eat/be/constants"
	"what-to-eat/be/model"

	"firebase.google.com/go/v4/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

func (s *UserService) Collection() *mongo.Collection {
	dbName := config.GetDBInstance().GetDbName()
	col := config.GetDBInstance().GetClient().Database(dbName).Collection(constants.USER_COLLECTION)
	return col
}

func (u *UserService) FindUserByUID(googleID string) (*model.User, error) {
	collection := u.Collection()
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
	collection := u.Collection()
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
	collection := u.Collection()
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

func (u *UserService) Create(createUserInput model.CreateUserDto, profile *model.JwtCustomClaims) (*model.User, error) {
	collection := u.Collection()
	now := time.Now()
	user := model.User{
		Email:       createUserInput.Email,
		Password:    createUserInput.Password,
		Name:        createUserInput.Name,
		DateOfBirth: createUserInput.DateOfBirth,
		Address:     createUserInput.Address,
		Phone:       createUserInput.Phone,
		GoogleID:    createUserInput.GoogleID,
		FacebookID:  new(string),
		GithubID:    new(string),
		Avatar:      createUserInput.Avatar,
		Deleted:     false,
		UpdatedAt:   &now,
		UpdatedBy:   &profile.ID,
		CreatedAt:   &now,
		CreatedBy:   &profile.ID,
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

func (u *UserService) Update(updateUserInput model.UpdateUserDto, profile *model.JwtCustomClaims) (*model.User, error) {
	collection := u.Collection()
	now := time.Now()

	var user model.User

	objectID, err := primitive.ObjectIDFromHex(updateUserInput.ID)
	if err != nil {
		log.Printf("Update user error: %s \n", err.Error())
	}
	filter := bson.M{"_id": objectID, "deleted": false}

	options := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	result := collection.FindOneAndUpdate(context.TODO(), filter, bson.M{"$set": bson.D{
		{Key: "email", Value: updateUserInput.Email},
		{Key: "name", Value: updateUserInput.Name},
		{Key: "dateOfBirth", Value: updateUserInput.DateOfBirth},
		{Key: "address", Value: updateUserInput.Address},
		{Key: "phone", Value: updateUserInput.Phone},
		{Key: "googleID", Value: updateUserInput.GoogleID},
		{Key: "facebookID", Value: updateUserInput.FacebookID},
		{Key: "githubID", Value: updateUserInput.GithubID},
		{Key: "avatar", Value: updateUserInput.Avatar},
		{Key: "updatedAt", Value: &now},
		{Key: "updatedBy", Value: &profile.ID},
	}}, options)
	if result.Err() != nil {
		return nil, result.Err()
	}
	decodeErr := result.Decode(&user)
	return &user, decodeErr
}

func (u *UserService) Remove(id string, profile *model.JwtCustomClaims) (*model.User, error) {
	collection := u.Collection()
	now := time.Now()
	var user model.User
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("Remove Role error: %s \n", err.Error())
	}
	filter := bson.M{"_id": objectID, "deleted": false}
	options := options.FindOneAndUpdate().SetReturnDocument(options.After)
	result := collection.FindOneAndUpdate(context.TODO(), filter, bson.M{"$set": bson.M{
		"deleted":   true,
		"deletedAt": now,
		"deletedBy": &profile.ID,
	}}, options)

	if result.Err() != nil {
		log.Println(err)
		return nil, err
	}

	decodeErr := result.Decode(&user)
	return &user, decodeErr
}

func (u *UserService) UpdateRole(id string, roleName string, profile *model.User) (*model.User, error) {
	collection := u.Collection()
	now := time.Now()
	var user model.User
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("Update Role error: %s \n", err.Error())
	}
	filter := bson.M{"_id": objectID, "deleted": false}
	options := options.FindOneAndUpdate().SetReturnDocument(options.After)
	result := collection.FindOneAndUpdate(context.TODO(), filter, bson.M{"$set": bson.M{
		"roleName":  roleName,
		"updatedAt": now,
		"updatedBy": &profile.ID,
	}}, options)

	if result.Err() != nil {
		log.Println(err)
		return nil, err
	}

	decodeErr := result.Decode(&user)
	return &user, decodeErr
}

func (u *UserService) FindAll(dto model.QueryUserDto) ([]*model.User, int64, error) {
	collection := u.Collection()
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetSkip((int64(dto.Page) - 1) * int64(dto.Limit)).SetLimit(int64(dto.Limit))
	filter := bson.D{{Key: "deleted", Value: false}}

	if dto.Keyword != "" {
		filter = append(filter, bson.E{Key: "$text", Value: bson.D{{Key: "$search", Value: dto.Keyword}}})
	}

	count, err := u.Collection().CountDocuments(context.TODO(), filter)
	if err != nil {
		return nil, 0, err
	}

	cursor, err := collection.Find(context.TODO(), filter, opts)
	if err != nil {
		log.Printf("Users error: %s \n", err.Error())
	}

	var users []*model.User
	if err = cursor.All(context.TODO(), &users); err != nil {
		log.Printf("Users error: %s \n", err.Error())
	}

	defer cursor.Close(context.TODO())
	return users, count, err
}

func (u *UserService) FindOne(id string) (*model.User, error) {
	collection := u.Collection()
	var user model.User
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("Get user error: %s \n", err.Error())
	}
	filter := bson.M{"_id": objectID, "deleted": false}
	result := collection.FindOne(context.TODO(), filter)
	if result.Err() != nil {
		return nil, result.Err()
	}
	decodeErr := result.Decode(&user)
	return &user, decodeErr
}

func (u *UserService) FindByEmail(email string) (*model.User, error) {
	collection := u.Collection()
	var user model.User
	filter := bson.M{"email": email, "deleted": false}
	result := collection.FindOne(context.TODO(), filter)
	if result.Err() != nil {
		return nil, result.Err()
	}
	decodeErr := result.Decode(&user)
	return &user, decodeErr
}
