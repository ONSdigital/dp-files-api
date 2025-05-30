package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"

	"github.com/ONSdigital/dp-files-api/config"
	"github.com/ONSdigital/dp-files-api/service"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/pkg/errors"
)

const serviceName = "dp-files-api"

var (
	// BuildTime represents the time in which the service was built
	BuildTime string
	// GitCommit represents the commit (SHA-1) hash of the service that is running
	GitCommit string
	// Version represents the version of the service that is running
	Version string

	// TODO: remove below explainer before commiting
	/* NOTE: replace the above with the below to run code with for example vscode debugger.
	   BuildTime string = "1601119818"
	   GitCommit string = "6584b786caac36b6214ffe04bf62f058d4021538"
	   Version   string = "v0.1.0"
	*/
)

func main() {
	log.Namespace = serviceName
	ctx := context.Background()

	if err := run(ctx); err != nil {
		log.Fatal(ctx, "fatal runtime error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	// Read config
	cfg, err := config.Get()
	if err != nil {
		return errors.Wrap(err, "error getting configuration")
	}

	// Run the service, providing an error channel for fatal errors
	svcErrors := make(chan error, 1)
	r := mux.NewRouter().StrictSlash(true)

	svcList, err := service.NewServiceList(ctx, cfg, BuildTime, GitCommit, Version, r)
	if err != nil {
		return errors.Wrap(err, "initialising services failed")
	}
	log.Info(ctx, "dp-files-api version", log.Data{"version": Version})

	// Start service
	svc, err := service.Run(ctx, svcList, svcErrors, cfg, r)
	if err != nil {
		return errors.Wrap(err, "running service failed")
	}

	// blocks until an os interrupt or a fatal error occurs
	select {
	case err := <-svcErrors:
		return errors.Wrap(err, "service error received")
	case sig := <-signals:
		log.Info(ctx, "os signal received", log.Data{"signal": sig})
	}
	return svc.Close(ctx, cfg.GracefulShutdownTimeout)
}
