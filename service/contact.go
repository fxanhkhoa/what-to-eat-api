package service

import (
	"context"
	"errors"
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

type ContactService struct{}

func (cs *ContactService) Collection() *mongo.Collection {
	dbName := config.GetDBInstance().GetDbName()
	col := config.GetDBInstance().GetClient().Database(dbName).Collection(constants.CONTACT_COLLECTION)
	return col
}

func (cs *ContactService) Create(createContactInput model.CreateContactDto) (*model.Contact, error) {
	collection := cs.Collection()

	now := time.Now()

	contact := model.Contact{
		Email:     createContactInput.Email,
		Name:      createContactInput.Name,
		Message:   createContactInput.Message,
		Deleted:   false,
		UpdatedAt: &now,
		UpdatedBy: nil,
		CreatedAt: &now,
		CreatedBy: nil,
	}

	result, err := collection.InsertOne(context.TODO(), contact)
	if err != nil {
		return nil, err
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		contact.ID = oid.Hex()
	}

	return &contact, nil
}

func (cs *ContactService) Update(updateContactInput model.UpdateContactDto, profile *model.JwtCustomClaims) (*model.Contact, error) {
	collection := cs.Collection()

	now := time.Now()

	var contact model.Contact

	objectID, err := primitive.ObjectIDFromHex(updateContactInput.ID)
	if err != nil {
		return nil, errors.New("invalid ID format")
	}

	filter := bson.M{"_id": objectID, "deleted": false}
	options := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)
	result := collection.FindOneAndUpdate(context.TODO(), filter, bson.M{"$set": bson.D{
		{Key: "email", Value: updateContactInput.Email},
		{Key: "name", Value: updateContactInput.Name},
		{Key: "message", Value: updateContactInput.Message},
		{Key: "updatedAt", Value: now},
		{Key: "updatedBy", Value: profile.ID},
	}}, options)
	if result.Err() != nil {
		return nil, result.Err()
	}
	decodeErr := result.Decode(&contact)
	return &contact, decodeErr
}

func (cs *ContactService) Remove(id string, profile *model.JwtCustomClaims) (*model.Contact, error) {
	collection := cs.Collection()
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
	contact := model.Contact{}
	decodeErr := result.Decode(&contact)
	return &contact, decodeErr
}

func (cs *ContactService) Find(query model.QueryContactDto) ([]*model.Contact, int64, error) {
	collection := cs.Collection()
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetSkip((int64(query.Page) - 1) * int64(query.Limit)).SetLimit(int64(query.Limit))
	filter := bson.D{{Key: "deleted", Value: false}}
	if query.Keyword != nil {
		filter = append(filter, bson.E{Key: "$text", Value: bson.D{{Key: "$search", Value: query.Keyword}}})
	}

	count, err := collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return nil, 0, err
	}

	cursor, err := collection.Find(context.TODO(), filter, opts)
	if err != nil {
		log.Println(err)
	}
	var contacts []*model.Contact
	if err = cursor.All(context.TODO(), &contacts); err != nil {
		log.Println(err)
	}
	defer cursor.Close(context.TODO())
	return contacts, count, err
}

func (cs *ContactService) FindOne(id string) (*model.Contact, error) {
	collection := cs.Collection()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid ID format")
	}

	filter := bson.M{"_id": objectID}
	result := collection.FindOne(context.TODO(), filter)
	if result.Err() != nil {
		return nil, result.Err()
	}
	contact := model.Contact{}
	decodeErr := result.Decode(&contact)
	return &contact, decodeErr
}
