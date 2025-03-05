package service

import (
	"context"
	"net/http"
	"time"

	auth "github.com/ONSdigital/dp-authorisation/v2/authorisation"

	"github.com/ONSdigital/dp-files-api/aws"
	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-files-api/store"

	"github.com/ONSdigital/dp-files-api/health"
	kafka "github.com/ONSdigital/dp-kafka/v3"

	"github.com/ONSdigital/dp-files-api/api"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/mongo"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

//	@title			dp-files-api
//	@version		1.0.0
//	@description	Digital Publishing API for handling file metadata.

//	@license.name	Open Government Licence v3.0
//	@license.url	http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/

//	@host		    localhost:26900
//  @schemes		http
//  @basePath		/

//  @tag.name       private

// Service contains all the configs, server and clients to run the API
type Service struct {
	Server         files.HTTPServer
	Router         *mux.Router
	ServiceList    ServiceContainer
	HealthCheck    health.Checker
	MongoClient    mongo.Client
	KafkaProducer  kafka.IProducer
	AuthMiddleware auth.Middleware
	S3Client       aws.S3Clienter
}

// Run the service
func Run(ctx context.Context, serviceList ServiceContainer, svcErrors chan error, cfg *config.Config, r *mux.Router) (*Service, error) {
	log.Info(ctx, "running service")

	mongoClient := serviceList.GetMongoDB()
	kafkaProducer := serviceList.GetKafkaProducer()
	hc := serviceList.GetHealthCheck()
	authMiddleware := serviceList.GetAuthMiddleware()
	s3Client := serviceList.GetS3Clienter()
	store := store.NewStore(
		mongoClient.Collection(config.MetadataCollection),
		mongoClient.Collection(config.CollectionsCollection),
		kafkaProducer,
		serviceList.GetClock(),
		s3Client,
		cfg,
	)

	getSingleFile := api.HandleGetFileMetadata(store.GetFileMetadata)

	const filesURI = "/files/{path:.*}"
	if cfg.IsPublishing {
		register := api.HandlerRegisterUploadStarted(store.RegisterFileUpload, cfg.MongoConfig.QueryTimeout)
		getMultipleFiles := api.HandlerGetFilesMetadata(store.GetFilesMetadata)
		collectionPublished := api.HandleMarkCollectionPublished(store.MarkCollectionPublished)

		r.Path("/files").HandlerFunc(authMiddleware.Require("static-files:create", register)).Methods(http.MethodPost)
		r.Path("/files").HandlerFunc(authMiddleware.Require("static-files:read", getMultipleFiles)).Methods(http.MethodGet)
		r.Path("/collection/{collectionID}").HandlerFunc(authMiddleware.Require("static-files:update", collectionPublished)).Methods(http.MethodPatch)
		r.Path(filesURI).HandlerFunc(authMiddleware.Require("static-files:read", getSingleFile)).Methods(http.MethodGet)

		patchRequestHandlers := api.PatchRequestHandlers{
			UploadComplete:   authMiddleware.Require("static-files:update", api.HandleMarkUploadComplete(store.MarkUploadComplete)),
			Published:        authMiddleware.Require("static-files:update", api.HandleMarkFilePublished(store.MarkFilePublished)),
			Moved:            authMiddleware.Require("static-files:update", api.HandleMarkFileMoved(store.MarkFileMoved)),
			CollectionUpdate: authMiddleware.Require("static-files:update", api.HandlerUpdateCollectionID(store.UpdateCollectionID)),
		}

		r.Path(filesURI).HandlerFunc(api.PatchRequestToHandler(patchRequestHandlers)).Methods(http.MethodPatch)
	} else {
		forbiddenHandler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}

		r.Path("/files").HandlerFunc(forbiddenHandler).Methods(http.MethodPost)
		r.Path(filesURI).HandlerFunc(forbiddenHandler).Methods(http.MethodPatch)
		r.Path("/collection/{collectionID}").HandlerFunc(forbiddenHandler).Methods(http.MethodPatch)

		// simple scenario - web mode where users are not authenticated - allowed based on publishing status
		r.Path(filesURI).HandlerFunc(getSingleFile).Methods(http.MethodGet)
	}
	r.Path("/health").HandlerFunc(hc.Handler)

	s := serviceList.GetHTTPServer()

	svc := &Service{
		Router:         r,
		HealthCheck:    hc,
		ServiceList:    serviceList,
		Server:         s,
		MongoClient:    mongoClient,
		KafkaProducer:  kafkaProducer,
		AuthMiddleware: authMiddleware,
		S3Client:       s3Client,
	}

	if err := svc.registerCheckers(ctx, hc, cfg.IsPublishing); err != nil {
		return nil, errors.Wrap(err, "unable to register checkers")
	}

	hc.Start(ctx)

	// Run the http server in a new go-routine
	go func() {
		if err := s.ListenAndServe(); err != nil {
			svcErrors <- errors.Wrap(err, "failure in http listen and serve")
		}
	}()

	return svc, nil
}

// Close gracefully shuts the service down in the required order, with timeout
func (svc *Service) Close(ctx context.Context, timeout time.Duration) error {
	log.Info(ctx, "commencing graceful shutdown")
	ctx, cancel := context.WithTimeout(ctx, timeout)

	var err error

	go func() {
		defer cancel()
		err = svc.ServiceList.Shutdown(ctx)
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	// timeout expired
	if ctx.Err() == context.DeadlineExceeded {
		log.Error(ctx, "shutdown timed out", ctx.Err())
		return ctx.Err()
	}

	// other error
	if err != nil {
		log.Error(ctx, "failed to shutdown gracefully ", err)
		return err
	}

	log.Info(ctx, "graceful shutdown was successful")
	return nil
}

func (svc *Service) registerCheckers(ctx context.Context, hc health.Checker, isPublishing bool) (err error) {
	hasErrors := false

	if err = hc.AddCheck("Mongo DB", svc.MongoClient.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding health for mongo db", err)
	}

	if isPublishing {
		if err = hc.AddCheck("Authorization Middleware", svc.AuthMiddleware.HealthCheck); err != nil {
			hasErrors = true
			log.Error(ctx, "error adding health for authorization middleware", err)
		}

		if err := hc.AddCheck("jwt keys state health check", svc.AuthMiddleware.IdentityHealthCheck); err != nil {
			hasErrors = true
			log.Error(ctx, "error getting jwt keys from identity service", err)
		}

		if err = hc.AddCheck("Kafka Producer", svc.KafkaProducer.Checker); err != nil {
			hasErrors = true
			log.Error(ctx, "error adding health for kafka producer", err)
		}

		if err = hc.AddCheck("S3 Client", svc.S3Client.Checker); err != nil {
			hasErrors = true
			log.Error(ctx, "error adding health for s3 client", err)
		}
	}

	if hasErrors {
		return errors.New("Error(s) registering checkers for healthcheck")
	}

	return nil
}
