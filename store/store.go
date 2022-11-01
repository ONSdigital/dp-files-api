package store

import (
	"github.com/ONSdigital/dp-files-api/aws"
	"github.com/ONSdigital/dp-files-api/clock"
	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-files-api/mongo"
	kafka "github.com/ONSdigital/dp-kafka/v3"
)

type Store struct {
	mongoCollection mongo.MongoCollection
	kafka           kafka.IProducer
	clock           clock.Clock
	s3client        aws.S3Clienter
	cfg             *config.Config
}

func NewStore(collection mongo.MongoCollection, kafkaProducer kafka.IProducer, clk clock.Clock, c aws.S3Clienter, cfg *config.Config) *Store {
	return &Store{collection, kafkaProducer, clk, c, cfg}
}
