package service

import (
	"context"
	"errors"
	"time"
	"what-to-eat/be/config"
	"what-to-eat/be/constants"
	"what-to-eat/be/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserDietaryPreferenceService struct{}

func NewUserDietaryPreferenceService() *UserDietaryPreferenceService {
	return &UserDietaryPreferenceService{}
}

func (s *UserDietaryPreferenceService) Collection() *mongo.Collection {
	dbName := config.GetDBInstance().GetDbName()
	col := config.GetDBInstance().GetClient().Database(dbName).Collection(constants.USER_DIETARY_PREFERENCE_COLLECTION)
	return col
}

func (s *UserDietaryPreferenceService) CreateOrUpdate(dto model.CreateUserDietaryPreferenceDto, profile *model.JwtCustomClaims) (*model.UserDietaryPreference, error) {
	collection := s.Collection()
	now := time.Now()

	preference := model.UserDietaryPreference{
		UserID:              dto.UserID,
		DietaryRestrictions: dto.DietaryRestrictions,
		Allergies:           dto.Allergies,
		DislikedIngredients: dto.DislikedIngredients,
		PreferredCuisines:   dto.PreferredCuisines,
		PreferredLabels:     dto.PreferredLabels,
		SpiceLevel:          dto.SpiceLevel,
		DifficultLevel:      dto.DifficultLevel,
		MaxCookingTime:      dto.MaxCookingTime,
		MaxPreparationTime:  dto.MaxPreparationTime,
		Deleted:             false,
		UpdatedAt:           &now,
		UpdatedBy:           &profile.ID,
	}

	filter := bson.M{"userId": dto.UserID, "deleted": false}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)
	
	update := bson.M{
		"$set": bson.M{
			"userId":              preference.UserID,
			"dietaryRestrictions": preference.DietaryRestrictions,
			"allergies":           preference.Allergies,
			"dislikedIngredients": preference.DislikedIngredients,
			"preferredCuisines":   preference.PreferredCuisines,
			"preferredLabels":     preference.PreferredLabels,
			"spiceLevel":          preference.SpiceLevel,
			"difficultLevel":      preference.DifficultLevel,
			"maxCookingTime":      preference.MaxCookingTime,
			"maxPreparationTime":  preference.MaxPreparationTime,
			"deleted":             false,
			"updatedAt":           now,
			"updatedBy":           profile.ID,
		},
		"$setOnInsert": bson.M{
			"createdAt": now,
			"createdBy": profile.ID,
		},
	}

	result := collection.FindOneAndUpdate(context.TODO(), filter, update, opts)
	if result.Err() != nil && result.Err() != mongo.ErrNoDocuments {
		return nil, result.Err()
	}

	decodeErr := result.Decode(&preference)
	if decodeErr != nil && decodeErr != mongo.ErrNoDocuments {
		return nil, decodeErr
	}

	return &preference, nil
}

func (s *UserDietaryPreferenceService) Update(dto model.UpdateUserDietaryPreferenceDto, profile *model.JwtCustomClaims) (*model.UserDietaryPreference, error) {
	collection := s.Collection()
	now := time.Now()

	objectID, err := primitive.ObjectIDFromHex(dto.ID)
	if err != nil {
		return nil, errors.New("invalid ID format")
	}

	filter := bson.M{"_id": objectID, "userId": dto.UserID, "deleted": false}
	update := bson.M{
		"$set": bson.M{
			"dietaryRestrictions": dto.DietaryRestrictions,
			"allergies":           dto.Allergies,
			"dislikedIngredients": dto.DislikedIngredients,
			"preferredCuisines":   dto.PreferredCuisines,
			"preferredLabels":     dto.PreferredLabels,
			"spiceLevel":          dto.SpiceLevel,
			"difficultLevel":      dto.DifficultLevel,
			"maxCookingTime":      dto.MaxCookingTime,
			"maxPreparationTime":  dto.MaxPreparationTime,
			"updatedAt":           now,
			"updatedBy":           profile.ID,
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var preference model.UserDietaryPreference
	result := collection.FindOneAndUpdate(context.TODO(), filter, update, opts)
	if result.Err() != nil {
		return nil, result.Err()
	}

	if err := result.Decode(&preference); err != nil {
		return nil, err
	}

	return &preference, nil
}

func (s *UserDietaryPreferenceService) Get(userID string) (*model.UserDietaryPreference, error) {
	collection := s.Collection()
	
	filter := bson.M{"userId": userID, "deleted": false}
	var preference model.UserDietaryPreference
	
	result := collection.FindOne(context.TODO(), filter)
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, result.Err()
	}

	if err := result.Decode(&preference); err != nil {
		return nil, err
	}

	return &preference, nil
}

func (s *UserDietaryPreferenceService) Delete(userID string, profile *model.JwtCustomClaims) error {
	collection := s.Collection()
	now := time.Now()

	filter := bson.M{"userId": userID, "deleted": false}
	update := bson.M{
		"$set": bson.M{
			"deleted":   true,
			"deletedAt": now,
			"deletedBy": profile.ID,
		},
	}

	result, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("dietary preference not found")
	}

	return nil
}

func (s *UserDietaryPreferenceService) HasPreferences(userID string) (bool, error) {
	collection := s.Collection()
	
	filter := bson.M{"userId": userID, "deleted": false}
	count, err := collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
