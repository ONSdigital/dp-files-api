package steps

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	permissionsAPISDK "github.com/ONSdigital/dp-permissions-api/sdk"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/ONSdigital/dp-authorisation/v2/authorisationtest"
	"github.com/ONSdigital/dp-files-api/files"

	kafka "github.com/ONSdigital/dp-kafka/v3"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-files-api/config"
	filesmongo "github.com/ONSdigital/dp-files-api/mongo"
	"github.com/ONSdigital/dp-files-api/service"
	dphttp "github.com/ONSdigital/dp-net/v3/http"
	"github.com/ONSdigital/log.go/v2/log"
)

type FilesAPIComponent struct {
	DpHTTPServer            *dphttp.Server
	svc                     *service.Service
	svcList                 service.ServiceContainer
	APIFeature              *componenttest.APIFeature
	errChan                 chan error
	mongoClient             *mongo.Client
	mongoStoreClient        filesmongo.Client
	mongoCfg                config.MongoConfig
	cg                      *kafka.ConsumerGroup
	msgs                    map[string]files.FilePublished
	msgsMu                  sync.RWMutex
	isPublishing            bool
	isAuthorised            bool
	isViewerAllowed         bool
	isViewerNotAllowed      bool
	Config                  *config.Config
	AuthorisationMiddleware authorisation.Middleware
	mongoURI                string

	initMu      sync.Mutex
	initHandler http.Handler
}

func buildMongoConfigFromURI(mongoURI string, cfg *config.Config) (config.MongoConfig, error) {
	mongoCfg := cfg.MongoConfig

	u, err := url.Parse(mongoURI)
	if err != nil {
		return mongoCfg, err
	}

	mongoCfg.ClusterEndpoint = u.Host
	if db := strings.TrimPrefix(u.Path, "/"); db != "" {
		mongoCfg.Database = db
	}

	return mongoCfg, nil
}

func (c *FilesAPIComponent) rebuildMongoStoreClient() error {
	mongoCfg, err := buildMongoConfigFromURI(c.mongoURI, c.Config)
	if err != nil {
		return err
	}

	mongoStoreClient, err := filesmongo.New(mongoCfg)
	if err != nil {
		return err
	}

	c.mongoCfg = mongoCfg
	c.mongoStoreClient = mongoStoreClient
	return nil
}

func waitForPrimary(ctx context.Context, c *mongo.Client) error {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	for {
		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		err := c.Ping(pingCtx, readpref.Primary())
		cancel()
		if err == nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return err
		case <-ticker.C:
		}
	}
}

func NewFilesAPIComponent(mongoURI, zebedeeURL string) (*FilesAPIComponent, error) {
	s := dphttp.NewServer("", http.NewServeMux())
	s.HandleOSSignals = false

	c := &FilesAPIComponent{
		DpHTTPServer: s,
		errChan:      make(chan error),
	}

	var err error
	c.Config, err = config.Get()
	if err != nil {
		return nil, err
	}

	fakePermissionsAPI := setupFakePermissionsAPI()

	c.Config.PermissionsAPIURL = fakePermissionsAPI.URL()
	c.Config.ZebedeeURL = zebedeeURL
	c.mongoURI = appendMongoOptions(mongoURI)

	c.isPublishing = true

	return c, nil
}

func getPermissionsBundle() *permissionsAPISDK.Bundle {
	return &permissionsAPISDK.Bundle{
		"static-files:read": {
			"users/service": {
				{
					ID: "1",
				},
			},
			"groups/role-publisher": {
				{
					ID: "1",
				},
			},
			"groups/role-admin": {
				{
					ID: "1",
				},
			},
			"groups/role-viewer-allowed": {
				{
					ID: "1",
					Condition: permissionsAPISDK.Condition{
						Values:    []string{"cpih01/feb-2026"},
						Attribute: "dataset_edition",
						Operator:  "StringEquals",
					},
				},
			},
			"groups/role-viewer-not-allowed": {
				{
					ID: "1",
					Condition: permissionsAPISDK.Condition{
						Values:    []string{"1/45"},
						Attribute: "dataset_edition",
						Operator:  "StringEquals",
					},
				},
			},
		},
		"static-files:update": {
			"groups/role-publisher": {
				{
					ID: "1",
				},
			},
			"users/service": {
				{
					ID: "1",
				},
			},
			"groups/role-admin": {
				{
					ID: "1",
				},
			},
		},
		"static-files:create": {
			"groups/role-publisher": {
				{
					ID: "1",
				},
			},
		},
		"static-files:delete": {
			"groups/role-publisher": {
				{
					ID: "1",
				},
			},
		},
	}
}

func setupFakePermissionsAPI() *authorisationtest.FakePermissionsAPI {
	fakePermissionsAPI := authorisationtest.NewFakePermissionsAPI()
	bundle := getPermissionsBundle()
	fakePermissionsAPI.Reset()
	if err := fakePermissionsAPI.UpdatePermissionsBundleResponse(bundle); err != nil {
		log.Error(context.Background(), "failed to update permissions bundle response", err)
	}
	return fakePermissionsAPI
}

func (c *FilesAPIComponent) DoGetAuthorisationMiddleware(ctx context.Context, cfg *authorisation.Config) (authorisation.Middleware, error) {
	middleware, err := authorisation.NewMiddlewareFromConfig(ctx, cfg, cfg.JWTVerificationPublicKeys)
	if err != nil {
		return nil, err
	}

	c.AuthorisationMiddleware = middleware
	return c.AuthorisationMiddleware, nil
}

func (c *FilesAPIComponent) Initialiser() (http.Handler, error) {
	c.initMu.Lock()
	defer c.initMu.Unlock()

	if c.initHandler != nil {
		return c.initHandler, nil
	}

	r := &mux.Router{}
	c.svcList = &fakeServiceContainer{c.DpHTTPServer, r, c.isAuthorised, c.isViewerAllowed, c.isViewerNotAllowed, c.mongoStoreClient}
	cfg, err := config.Get()
	if err != nil {
		return nil, err
	}
	cfg.IsPublishing = c.isPublishing

	c.svc, err = service.Run(context.Background(), c.svcList, c.errChan, cfg, r)
	if err != nil {
		return nil, err
	}

	c.initHandler = r
	return c.initHandler, nil
}

func (c *FilesAPIComponent) Reset() {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	client, err := mongo.Connect(
		ctx,
		options.Client().ApplyURI(c.mongoURI),
	)
	if err != nil {
		log.Error(ctx, "failed to connect to mongo", err)
		panic(err)
	}
	c.mongoClient = client

	if err = waitForPrimary(ctx, c.mongoClient); err != nil {
		log.Error(ctx, "mongo primary not ready", err, log.Data{"mongo_uri": c.mongoURI})
		panic(err)
	}

	if err = c.rebuildMongoStoreClient(); err != nil {
		log.Error(ctx, "failed to rebuild mongo store client", err, log.Data{"mongo_uri": c.mongoURI})
		panic(err)
	}

	db := c.mongoClient.Database("files")

	// Reset whole database
	if err = db.Drop(ctx); err != nil {
		log.Error(ctx, "failed to drop files database", err)
		panic(err)
	}

	// Recreate collections
	for _, name := range []string{"metadata", "collections", "bundles", "file_events"} {
		if err = db.CreateCollection(ctx, name); err != nil {
			log.Error(ctx, "failed to create collection", err, log.Data{"collection": name})
			panic(err)
		}
	}

	// Recreate indexes
	if _, err = db.Collection("metadata").Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "path", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	); err != nil {
		log.Error(ctx, "failed to create index on metadata collection", err)
		panic(err)
	}

	if _, err = db.Collection("collections").Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	); err != nil {
		log.Error(ctx, "failed to create index on collections collection", err)
		panic(err)
	}

	if _, err = db.Collection("bundles").Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	); err != nil {
		log.Error(ctx, "failed to create index on bundles collection", err)
		panic(err)
	}

	key, err := c.APIFeature.JWTFeature.EnsureKeys()
	if err != nil {
		panic(err)
	}
	c.Config.JWTVerificationPublicKeys[key.KID] = key.PublicKeyB64

	c.isPublishing = true
}

func (c *FilesAPIComponent) Close() error {
	if c.cg != nil {
		if err := c.cg.Stop(); err != nil {
			return err
		}
	}

	cfg, _ := config.Get()

	if c.svc != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := c.svc.Close(ctx, cfg.GracefulShutdownTimeout); err != nil {
			return err
		}
	}

	c.initMu.Lock()
	c.initHandler = nil
	c.initMu.Unlock()
	return nil
}

func appendMongoOptions(uri string) string {
	opts := "directConnection=true&serverSelectionTimeoutMS=30000"
	if strings.Contains(uri, "?") {
		return uri + "&" + opts
	}
	return uri + "?" + opts
}

func (c *FilesAPIComponent) SetMongoURI(mongoURI string) {
	c.mongoURI = appendMongoOptions(mongoURI)
}
