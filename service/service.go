package service

import (
	"context"
	"github.com/ONSdigital/dp-files-api/health"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"time"

	"github.com/ONSdigital/dp-files-api/api"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/mongo"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// Service contains all the configs, server and clients to run the API
type Service struct {
	Server        files.HTTPServer
	Router        *mux.Router
	ServiceList   ServiceContainer
	HealthCheck   health.Checker
	MongoClient   mongo.Client
	KafkaProducer kafka.IProducer
}

// Run the service
func Run(ctx context.Context, serviceList ServiceContainer, svcErrors chan error) (*Service, error) {

	log.Info(ctx, "running service")

	//log.Info(ctx, "using service configuration", log.Data{"config": cfg})

	mongoClient, err := serviceList.GetMongoDB(ctx)
	if err != nil {
		log.Error(ctx, "could not obtain mongo session", err)
		return nil, err
	}

	kafkaProducer, err := serviceList.GetKafkaProducer(ctx)
	if err != nil {
		log.Error(ctx, "could not obtain kafka connection", err)
		return nil, err
	}

	store := files.NewStore(mongoClient, serviceList.GetClock(ctx))

	r := mux.NewRouter() // TODO: Add any middleware that your service requires
	r.StrictSlash(true).Path("/v1/files/register").HandlerFunc(api.CreateFileUploadStartedHandler(store.RegisterFileUpload))
	r.StrictSlash(true).Path("/v1/files/upload-complete").HandlerFunc(api.CreateMarkUploadCompleteHandler(store.MarkUploadComplete))
	r.StrictSlash(true).Path("/v1/files/publish").HandlerFunc(api.CreatePublishHandler(store.PublishCollection))

	// The path below is the catchall route and MUST be the last one
	r.StrictSlash(true).Path("/v1/files/{path:[a-zA-Z0-9\\.\\-\\/]+}").HandlerFunc(api.CreateGetFileMetadataHandler(store.GetFileMetadata))

	s := serviceList.GetHTTPServer(r)

	// TODO: Add other(s) to serviceList here

	hc, err := serviceList.GetHealthCheck()

	if err != nil {
		log.Fatal(ctx, "could not instantiate healthcheck", err)
		return nil, err
	}

	svc := &Service{
		Router:        r,
		HealthCheck:   hc,
		ServiceList:   serviceList,
		Server:        s,
		MongoClient:   mongoClient,
		KafkaProducer: kafkaProducer,
	}

	if err := svc.registerCheckers(ctx, hc); err != nil {
		return nil, errors.Wrap(err, "unable to register checkers")
	}

	r.StrictSlash(true).Path("/health").HandlerFunc(hc.Handler)
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

func (svc *Service) registerCheckers(ctx context.Context, hc health.Checker) (err error) {
	hasErrors := false

	if err = hc.AddCheck("Mongo DB", svc.MongoClient.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding health for mongo db", err)
	}

	if err = hc.AddCheck("Kafka Producer", svc.KafkaProducer.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding health for kafka producer", err)
	}

	if hasErrors {
		return errors.New("Error(s) registering checkers for healthcheck")
	}

	return nil
}
