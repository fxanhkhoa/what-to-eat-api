package service

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SingleResult abstracts *mongo.SingleResult for testability.
type SingleResult interface {
	Decode(v interface{}) error
	Err() error
}

// Cursor abstracts *mongo.Cursor for testability.
type Cursor interface {
	All(ctx context.Context, results interface{}) error
	Close(ctx context.Context) error
}

// CollectionInterface abstracts the MongoDB collection operations used by services.
type CollectionInterface interface {
	InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
	FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...*options.FindOneAndUpdateOptions) SingleResult
	CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error)
	Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (Cursor, error)
	FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) SingleResult
	UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)
}

// mongoCollectionAdapter wraps *mongo.Collection to implement CollectionInterface.
type mongoCollectionAdapter struct {
	col *mongo.Collection
}

// NewMongoCollectionAdapter returns a CollectionInterface backed by a real *mongo.Collection.
func NewMongoCollectionAdapter(col *mongo.Collection) CollectionInterface {
	return &mongoCollectionAdapter{col: col}
}

func (a *mongoCollectionAdapter) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	return a.col.InsertOne(ctx, document, opts...)
}

func (a *mongoCollectionAdapter) FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...*options.FindOneAndUpdateOptions) SingleResult {
	return a.col.FindOneAndUpdate(ctx, filter, update, opts...)
}

func (a *mongoCollectionAdapter) CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	return a.col.CountDocuments(ctx, filter, opts...)
}

func (a *mongoCollectionAdapter) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (Cursor, error) {
	return a.col.Find(ctx, filter, opts...)
}

func (a *mongoCollectionAdapter) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) SingleResult {
	return a.col.FindOne(ctx, filter, opts...)
}

func (a *mongoCollectionAdapter) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return a.col.UpdateOne(ctx, filter, update, opts...)
}

func (a *mongoCollectionAdapter) UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return a.col.UpdateMany(ctx, filter, update, opts...)
}

func (a *mongoCollectionAdapter) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return a.col.DeleteOne(ctx, filter, opts...)
}
