package mongo

import (
	"context"

	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dpMongoLock "github.com/ONSdigital/dp-mongodb/v2/dplock"
	dpMongoHealth "github.com/ONSdigital/dp-mongodb/v2/health"
	dpmongo "github.com/ONSdigital/dp-mongodb/v2/mongodb"
)

const (
	connectTimeoutInSeconds = 5
	queryTimeoutInSeconds   = 15
)

// Mongo represents a simplistic MongoDB configuration.
type Mongo struct {
	datasetURL   string
	connection   *dpmongo.MongoConnection
	uri          string
	client       *dpMongoHealth.Client
	healthClient *dpMongoHealth.CheckMongoClient
	lockClient   *dpMongoLock.Lock
}

func New(ctx context.Context, cfg *config.Config) (*Mongo, error) {
	m := &Mongo{
		uri:        cfg.MongoConfig.URI,
	}

	connCfg := &dpmongo.MongoConnectionConfig{
		ConnectTimeoutInSeconds: connectTimeoutInSeconds,
		QueryTimeoutInSeconds:   queryTimeoutInSeconds,

		Username:                      cfg.MongoConfig.Username,
		Password:                      cfg.MongoConfig.Password,
		ClusterEndpoint:               cfg.MongoConfig.URI,
		Database:                      cfg.MongoConfig.Database,
		Collection:                    cfg.MongoConfig.Collection,
		IsWriteConcernMajorityEnabled: true,
		IsStrongReadConcernEnabled:    false,
	}

	conn, err := dpmongo.Open(connCfg)
	if err != nil {
		return nil, err
	}
	m.connection = conn

	// set up databaseCollectionBuilder here when collections are known

	// Create healthClient and healthclient from session
	m.client = dpMongoHealth.NewClient(m.connection)
	m.healthClient = &dpMongoHealth.CheckMongoClient{
		Client:      *m.client,
		Healthcheck: m.client.Healthcheck,
	}

	// create lock healthClient here when collections are known
	return m, nil
}

func (m *Mongo) URI() string {
	return m.uri
}

// Close represents mongo session closing within the context deadline
func (m *Mongo) Close(ctx context.Context) error {
	m.lockClient.Close(ctx)
	return m.connection.Close(ctx)
}

// Checker is called by the healthcheck library to check the health state of this mongoDB instance
func (m *Mongo) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	return m.healthClient.Checker(ctx, state)
}
