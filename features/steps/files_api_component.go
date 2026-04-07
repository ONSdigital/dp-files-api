package steps

import (
	"context"
	"crypto/rsa"
	"net/http"
	"time"

	permissionsAPISDK "github.com/ONSdigital/dp-permissions-api/sdk"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/ONSdigital/dp-authorisation/v2/authorisationtest"
	"github.com/ONSdigital/dp-files-api/files"

	kafka "github.com/ONSdigital/dp-kafka/v3"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-files-api/config"
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
	cg                      *kafka.ConsumerGroup
	msgs                    map[string]files.FilePublished
	isPublishing            bool
	isAuthorised            bool
	isViewerAllowed         bool
	isViewerNotAllowed      bool
	Config                  *config.Config
	AuthorisationMiddleware authorisation.Middleware
	viewerPrivKey           *rsa.PrivateKey
	viewerKID               string
}

func NewFilesAPIComponent(zebedeeURL string) (*FilesAPIComponent, error) {
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

	log.Namespace = "dp-files-api"

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
	r := &mux.Router{}
	c.svcList = &fakeServiceContainer{c.DpHTTPServer, r, c.isAuthorised, c.isViewerAllowed, c.isViewerNotAllowed}
	cfg, _ := config.Get()
	cfg.IsPublishing = c.isPublishing
	c.svc, _ = service.Run(context.Background(), c.svcList, c.errChan, cfg, r)

	return c.DpHTTPServer.Handler, nil
}

func (c *FilesAPIComponent) Reset() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, _ := mongo.Connect(
		context.Background(),
		options.Client().ApplyURI("mongodb://root:password@mongo:27017"))

	c.mongoClient = client
	if err := c.mongoClient.Database("files").Collection("metadata").Drop(ctx); err != nil {
		log.Error(ctx, "failed to drop metadata collection", err)
		panic(err)
	}
	if err := c.mongoClient.Database("files").CreateCollection(ctx, "metadata"); err != nil {
		log.Error(ctx, "failed to create metadata collection", err)
		panic(err)
	}
	if _, err := c.mongoClient.Database("files").Collection("metadata").Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "path", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	); err != nil {
		log.Error(ctx, "failed to create index on metadata collection", err)
		panic(err)
	}
	if err := c.mongoClient.Database("files").Collection("collections").Drop(ctx); err != nil {
		log.Error(ctx, "failed to drop collections collection", err)
		panic(err)
	}
	if err := c.mongoClient.Database("files").CreateCollection(ctx, "collections"); err != nil {
		log.Error(ctx, "failed to create collections collection", err)
		panic(err)
	}
	if _, err := c.mongoClient.Database("files").Collection("collections").Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	); err != nil {
		log.Error(ctx, "failed to create index on collections collection", err)
		panic(err)
	}
	if err := c.mongoClient.Database("files").Collection("bundles").Drop(ctx); err != nil {
		log.Error(ctx, "failed to drop bundles collection", err)
		panic(err)
	}
	if err := c.mongoClient.Database("files").CreateCollection(ctx, "bundles"); err != nil {
		log.Error(ctx, "failed to create bundles collection", err)
		panic(err)
	}
	if _, err := c.mongoClient.Database("files").Collection("bundles").Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	); err != nil {
		log.Error(ctx, "failed to create index on bundles collection", err)
		panic(err)
	}

	_ = c.mongoClient.Database("files").Collection("file_events").Drop(ctx)

	if err := c.mongoClient.Database("files").CreateCollection(ctx, "file_events"); err != nil {
		log.Error(ctx, "failed to create file_events collection", err)
		panic(err)
	}

	key, err := c.APIFeature.JWTFeature.EnsureKeys()
	if err != nil {
		panic(err)
	}
	c.Config.AuthConfig.JWTVerificationPublicKeys[key.KID] = key.PublicKeyB64

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
		return c.svc.Close(ctx, cfg.GracefulShutdownTimeout)
	}
	return nil
}
