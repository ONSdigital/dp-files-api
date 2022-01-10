package mongo

import (
	"context"

	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	mongohealth "github.com/ONSdigital/dp-mongodb/v3/health"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
)

//go:generate moq -out mock/Client.go -pkg mock . Client
type Client interface {
	URI() string
	Close(context.Context) error
	Checker(context.Context, *healthcheck.CheckState) error
	Connection() *mongodriver.MongoConnection
}

// Mongo represents a simplistic MongoDB configuration.
type Mongo struct {
	mongodriver.MongoConnectionConfig

	conn         *mongodriver.MongoConnection
	healthClient *mongohealth.CheckMongoClient
}

func New(cfg config.MongoConfig) (m *Mongo, err error) {

	m = &Mongo{MongoConnectionConfig: cfg}
	m.conn, err = mongodriver.Open(&m.MongoConnectionConfig)
	if err != nil {
		return nil, err
	}

	m.healthClient = mongohealth.NewClient(m.conn)

	return m, nil
}

func (m *Mongo) URI() string {
	return m.ClusterEndpoint
}

// Close represents mongo session closing within the context deadline
func (m *Mongo) Close(ctx context.Context) error {
	return m.conn.Close(ctx)
}

// Checker is called by the healthcheck library to health the health state of this mongoDB instance
func (m *Mongo) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	return m.healthClient.Checker(ctx, state)
}

func (m *Mongo) Connection() *mongodriver.MongoConnection {
	return m.conn
}
