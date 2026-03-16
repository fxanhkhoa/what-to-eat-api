package service

import (
	"context"
	"errors"
	"log"
	"time"
	"what-to-eat/be/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ContactService struct {
	col CollectionInterface
}

func NewContactService(col CollectionInterface) *ContactService {
	return &ContactService{col: col}
}

func (cs *ContactService) Create(createContactInput model.CreateContactDto) (*model.Contact, error) {
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

	result, err := cs.col.InsertOne(context.TODO(), contact)
	if err != nil {
		return nil, err
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		contact.ID = oid.Hex()
	}

	return &contact, nil
}

func (cs *ContactService) Update(updateContactInput model.UpdateContactDto, profile *model.JwtCustomClaims) (*model.Contact, error) {
	now := time.Now()

	objectID, err := primitive.ObjectIDFromHex(updateContactInput.ID)
	if err != nil {
		return nil, errors.New("invalid ID format")
	}

	filter := bson.M{"_id": objectID, "deleted": false}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)
	result := cs.col.FindOneAndUpdate(context.TODO(), filter, bson.M{"$set": bson.D{
		{Key: "email", Value: updateContactInput.Email},
		{Key: "name", Value: updateContactInput.Name},
		{Key: "message", Value: updateContactInput.Message},
		{Key: "updatedAt", Value: now},
		{Key: "updatedBy", Value: profile.ID},
	}}, opts)
	if result.Err() != nil {
		return nil, result.Err()
	}
	var contact model.Contact
	decodeErr := result.Decode(&contact)
	return &contact, decodeErr
}

func (cs *ContactService) Remove(id string, profile *model.JwtCustomClaims) (*model.Contact, error) {
	now := time.Now()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid ID format")
	}

	filter := bson.M{"_id": objectID, "deleted": false}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	result := cs.col.FindOneAndUpdate(context.TODO(), filter, bson.M{"$set": bson.M{
		"deleted":   true,
		"deletedAt": now,
		"deletedBy": profile.ID,
	}}, opts)
	if result.Err() != nil {
		return nil, result.Err()
	}
	contact := model.Contact{}
	decodeErr := result.Decode(&contact)
	return &contact, decodeErr
}

func (cs *ContactService) Find(query model.QueryContactDto) ([]*model.Contact, int64, error) {
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetSkip((int64(query.Page) - 1) * int64(query.Limit)).SetLimit(int64(query.Limit))
	filter := bson.D{{Key: "deleted", Value: false}}
	if query.Keyword != nil {
		filter = append(filter, bson.E{Key: "$text", Value: bson.D{{Key: "$search", Value: query.Keyword}}})
	}

	count, err := cs.col.CountDocuments(context.TODO(), filter)
	if err != nil {
		return nil, 0, err
	}

	cursor, err := cs.col.Find(context.TODO(), filter, opts)
	if err != nil {
		log.Println(err)
		return nil, count, err
	}
	var contacts []*model.Contact
	if err = cursor.All(context.TODO(), &contacts); err != nil {
		log.Println(err)
	}
	defer cursor.Close(context.TODO())
	return contacts, count, err
}

func (cs *ContactService) FindOne(id string) (*model.Contact, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid ID format")
	}

	filter := bson.M{"_id": objectID}
	result := cs.col.FindOne(context.TODO(), filter)
	if result.Err() != nil {
		return nil, result.Err()
	}
	contact := model.Contact{}
	decodeErr := result.Decode(&contact)
	return &contact, decodeErr
}
