package mongo

import (
	"context"

	"github.com/ONSdigital/dp-mongodb/v3/mongodb"
	lock "github.com/square/mongo-lock"
)

//go:generate moq -out mock/collection.go -pkg mock . MongoCollection
// MongoCollection defines the required methods from the MongoDB Collection

type MongoCollection interface {
	Must() *mongodb.Must
	Distinct(ctx context.Context, fieldName string, filter interface{}) ([]interface{}, error)
	Count(ctx context.Context, filter interface{}, opts ...mongodb.FindOption) (int, error)
	Find(ctx context.Context, filter interface{}, results interface{}, opts ...mongodb.FindOption) (int, error)
	FindCursor(ctx context.Context, filter interface{}, opts ...mongodb.FindOption) (mongodb.Cursor, error)
	FindOne(ctx context.Context, filter interface{}, result interface{}, opts ...mongodb.FindOption) error
	Insert(ctx context.Context, document interface{}) (*mongodb.CollectionInsertResult, error)
	InsertMany(ctx context.Context, documents []interface{}) (*mongodb.CollectionInsertManyResult, error)
	Upsert(ctx context.Context, selector interface{}, update interface{}) (*mongodb.CollectionUpdateResult, error)
	UpsertById(ctx context.Context, id interface{}, update interface{}) (*mongodb.CollectionUpdateResult, error)
	UpdateById(ctx context.Context, id interface{}, update interface{}) (*mongodb.CollectionUpdateResult, error)
	Update(ctx context.Context, selector interface{}, update interface{}) (*mongodb.CollectionUpdateResult, error)
	UpdateMany(ctx context.Context, selector interface{}, update interface{}) (*mongodb.CollectionUpdateResult, error)
	Delete(ctx context.Context, selector interface{}) (*mongodb.CollectionDeleteResult, error)
	DeleteMany(ctx context.Context, selector interface{}) (*mongodb.CollectionDeleteResult, error)
	DeleteById(ctx context.Context, id interface{}) (*mongodb.CollectionDeleteResult, error)
	Aggregate(ctx context.Context, pipeline interface{}, results interface{}) error
	NewLockClient() *lock.Client
}

//go:generate moq -out mock/cursor.go -pkg mock . MongoCursor
type MongoCursor interface {
	mongodb.Cursor
}
