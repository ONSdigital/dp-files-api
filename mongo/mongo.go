package mongo

import (
	"context"
	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dpMongoHealth "github.com/ONSdigital/dp-mongodb/v3/health"
	dpmongo "github.com/ONSdigital/dp-mongodb/v3/mongodb"
)

const (
	connectTimeoutInSeconds = 5
	queryTimeoutInSeconds   = 5
)

//go:generate moq -out mock/Client.go -pkg mock . Client
type Client interface {
	URI() string
	Close(context.Context) error
	Checker(context.Context, *healthcheck.CheckState) error
	Connection() *dpmongo.MongoConnection
}

// Mongo represents a simplistic MongoDB configuration.
type Mongo struct {
	datasetURL string
	conn       *dpmongo.MongoConnection
	uri        string
	healthClient *dpMongoHealth.CheckMongoClient
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

		TLSConnectionConfig: dpmongo.TLSConnectionConfig{
			IsSSL: cfg.MongoConfig.IsSSL,
		},
	}

	conn, err := dpmongo.Open(connCfg)
	if err != nil {
		return nil, err
	}
	m.conn = conn

	m.healthClient = dpMongoHealth.NewClient(m.conn)

	// create lock healthClient here when collections are known
	return m, nil
}

func (m *Mongo) URI() string {
	return m.uri
}

// Close represents mongo session closing within the context deadline
func (m *Mongo) Close(ctx context.Context) error {
	return m.conn.Close(ctx)
}

// Checker is called by the healthcheck library to check the health state of this mongoDB instance
func (m *Mongo) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	return m.healthClient.Checker(ctx, state)
}

func (m *Mongo) Connection() *dpmongo.MongoConnection {
	return m.conn
}
