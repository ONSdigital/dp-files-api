package files

import "go.mongodb.org/mongo-driver/bson"

func createCollectionContainsNotUploadedFilesQuery(collectionID string) bson.M {
	return bson.M{"$and": []bson.M{
		{"collection_id": collectionID},
		{"state": bson.M{"$ne": StateUploaded}},
	}}
}
