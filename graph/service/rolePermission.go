package service

import (
	"context"
	"log"
	"time"
	"what-to-eat/be/graph/model"
	"what-to-eat/be/shared"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RolePermissionService struct{}

func NewRolePermissionService() *RolePermissionService {
	return &RolePermissionService{}
}

func (r *RolePermissionService) Create(input model.CreateRolePermissionInput, profile *model.User) (*model.RolePermission, error) {
	collection := shared.Init("RolePermissions")
	now := time.Now()

	rolePermission := model.RolePermission{
		Name:        input.Name,
		Permission:  input.Permission,
		Description: input.Description,
		Deleted:     false,
		UpdatedAt:   &now,
		UpdatedBy:   &profile.ID,
		CreatedAt:   &now,
		CreatedBy:   &profile.ID,
	}

	filter := bson.M{"name": input.Name, "deleted": true}
	options := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	result := collection.FindOneAndUpdate(context.TODO(), filter, bson.M{"$set": rolePermission}, options)
	if result.Err() != nil {
		return nil, result.Err()
	}
	decodeErr := result.Decode(&rolePermission)
	return &rolePermission, decodeErr
}

func (r *RolePermissionService) Update(input model.UpdateRolePermissionInput, profile *model.User) (*model.RolePermission, error) {
	collection := shared.Init("RolePermissions")
	now := time.Now()

	rolePermission := model.RolePermission{
		Name:        input.Name,
		Permission:  input.Permission,
		Description: input.Description,
		UpdatedAt:   &now,
		UpdatedBy:   &profile.ID,
	}

	objectID, err := primitive.ObjectIDFromHex(input.ID)
	if err != nil {
		log.Printf("Update user error: %s \n", err.Error())
	}
	filter := bson.M{"_id": objectID, "deleted": false}
	options := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	result := collection.FindOneAndUpdate(context.TODO(), filter, bson.M{"$set": rolePermission}, options)
	if result.Err() != nil {
		return nil, result.Err()
	}
	decodeErr := result.Decode(&rolePermission)
	return &rolePermission, decodeErr
}

func (r *RolePermissionService) FindOne(id string) (*model.RolePermission, error) {
	collection := shared.Init("RolePermissions")
	var rolePermission model.RolePermission
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("Update user error: %s \n", err.Error())
	}
	filter := bson.M{"_id": objectID, "deleted": false}
	result := collection.FindOne(context.TODO(), filter)
	if result.Err() != nil {
		return nil, result.Err()
	}
	decodeErr := result.Decode(&rolePermission)
	return &rolePermission, decodeErr
}

func (r *RolePermissionService) Find(page *int, limit *int) ([]*model.RolePermission, error) {
	collection := shared.Init("RolePermissions")
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetSkip((int64(*page) - 1) * int64(*limit)).SetLimit(int64(*limit))
	filter := bson.D{{Key: "deleted", Value: false}}
	cursor, err := collection.Find(context.TODO(), filter, opts)
	if err != nil {
		log.Println(err)
	}
	var rolePermissions []*model.RolePermission
	if err = cursor.All(context.TODO(), &rolePermissions); err != nil {
		log.Println(err)
	}
	defer cursor.Close(context.TODO())
	return rolePermissions, err
}

func (r *RolePermissionService) Remove(id string, profile *model.User) (*model.RolePermission, error) {
	collection := shared.Init("RolePermissions")
	now := time.Now()
	var rolePermission model.RolePermission
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

	decodeErr := result.Decode(&rolePermission)
	return &rolePermission, decodeErr
}

func (r *RolePermissionService) FindByName(name string) (*model.RolePermission, error) {
	collection := shared.Init("RolePermissions")
	var rolePermission model.RolePermission
	filter := bson.M{"name": name, "deleted": false}
	result := collection.FindOne(context.TODO(), filter)
	if result.Err() != nil {
		return nil, result.Err()
	}
	decodeErr := result.Decode(&rolePermission)
	return &rolePermission, decodeErr
}
