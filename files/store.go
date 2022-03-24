package files

import (
	"github.com/ONSdigital/dp-files-api/clock"
	kafka "github.com/ONSdigital/dp-kafka/v3"
)

type Store struct {
	mongoCollection MongoCollection
	kafka           kafka.IProducer
	clock           clock.Clock
}

func NewStore(collection MongoCollection, kafkaProducer kafka.IProducer, clk clock.Clock) *Store {
	return &Store{collection, kafkaProducer, clk}
}
