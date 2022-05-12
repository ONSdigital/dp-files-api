package store

import (
	"go.mongodb.org/mongo-driver/bson"
)

func createCollectionContainsNotUploadedFilesQuery(collectionID string) bson.M {
	return bson.M{"$and": []bson.M{
		{fieldCollectionID: collectionID},
		{fieldState: bson.M{"$ne": StateUploaded}},
	}}
}
