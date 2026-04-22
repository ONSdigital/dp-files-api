package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/cucumber/godog/colors"

	"github.com/ONSdigital/log.go/v2/log"

	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-files-api/features/steps"
	"github.com/cucumber/godog"
)

var (
	componentFlag = flag.Bool("component", false, "perform component tests")
	loggingFlag   = flag.Bool("logging", false, "print logging")
)

const mongoVersion = "4.4.8"
const databaseName = "files"
const replicaSetName = "rs0"

type ComponentTest struct {
	MongoFeature *componenttest.MongoFeature
}

func (f *ComponentTest) InitializeScenario(ctx *godog.ScenarioContext) {
	authorizationFeature := componenttest.NewAuthorizationFeature()

	mongoURI, err := f.MongoFeature.GetConnectionString()
	if err != nil {
		panic(err)
	}

	component, err := steps.NewFilesAPIComponent(mongoURI, authorizationFeature.FakeAuthService.Server.URL)
	if err != nil {
		panic(err)
	}

	apiFeature := componenttest.NewAPIFeature(component.Initialiser)
	component.APIFeature = apiFeature

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		component.Reset()
		return ctx, nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, e error) (context.Context, error) {
		err := component.Close()
		if err != nil {
			log.Error(ctx, "error closing service", err)
		}
		return ctx, nil
	})

	apiFeature.RegisterSteps(ctx)
	component.RegisterSteps(ctx)
	authorizationFeature.RegisterSteps(ctx)
}

func (f *ComponentTest) InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
		if !*loggingFlag {
			buf := bytes.NewBufferString("")
			log.SetDestination(buf, buf)
		}
		f.MongoFeature = componenttest.NewMongoFeature(componenttest.MongoOptions{MongoVersion: mongoVersion, DatabaseName: databaseName, ReplicaSetName: replicaSetName})
	})
	ctx.AfterSuite(func() {
		if err := f.MongoFeature.Close(); err != nil {
			log.Error(context.Background(), "failed to close mongo feature", err)
		}
	})
}

func TestComponent(t *testing.T) {
	if *componentFlag {
		f := &ComponentTest{}

		status := godog.TestSuite{
			Name:                 "feature_tests",
			ScenarioInitializer:  f.InitializeScenario,
			TestSuiteInitializer: f.InitializeTestSuite,
			Options: &godog.Options{
				Output:      colors.Colored(os.Stdout),
				Format:      "pretty",
				Paths:       flag.Args(),
				Concurrency: 1,
			}}.Run()

		fmt.Println("=================================")
		fmt.Printf("Component test coverage: %.2f%%\n", testing.Coverage()*100)
		fmt.Println("=================================")

		if status > 0 {
			t.Fail()
		}
	} else {
		t.Skip("component flag required to run component tests")
	}
}
