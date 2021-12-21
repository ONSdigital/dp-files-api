BINPATH ?= build

BUILD_TIME=$(shell date +%s)
GIT_COMMIT=$(shell git rev-parse HEAD)
VERSION ?= $(shell git tag --points-at HEAD | grep ^v | head -n 1)

LDFLAGS = -ldflags "-X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(VERSION)"

.PHONY: all
all: audit test build

.PHONY: audit
audit:
	go list -json -m all | nancy sleuth

.PHONY: lint
lint:
	exit

.PHONY: build
build:
	go build -tags 'production' $(LDFLAGS) -o $(BINPATH)/dp-files-api

.PHONY: debug
debug:
	go build -tags 'debug' $(LDFLAGS) -o $(BINPATH)/dp-files-api
	HUMAN_LOG=1 DEBUG=1 $(BINPATH)/dp-files-api

.PHONY: test
test:
	go test -race -cover ./...

.PHONY: convey
convey:
	goconvey ./...

.PHONY: test-component
test-component:
	go test -cover -coverpkg=github.com/ONSdigital/dp-files-api/... -component

.PHONY: docker-test
docker-test-component:
	docker-compose -f docker-compose.yml down
	docker build -f Dockerfile . -t template_test --target=test
	docker-compose -f docker-compose.yml up -d
	docker-compose -f docker-compose.yml exec -T http go test -component
	docker-compose -f docker-compose.yml down

# Enabling components to run on an M1 chip as Mongo cannot be installed on Apple Silicon
m1-docker-test:
	docker-compose -f docker-compose.yml down
	docker buildx build --platform linux/amd64 -f Dockerfile . -t template_test --target=test
	docker-compose -f docker-compose.yml up -d
	docker-compose -f docker-compose.yml exec -T http go test ./...
	docker-compose -f docker-compose.yml exec -T http go test -component
	docker-compose -f docker-compose.yml down

.PHONY: test-coverage
test-coverage:
	rm combined-coverage.out component-coverage.out coverage.out
	go test -cover ./... -coverprofile=coverage.out
	go test -component -cover -coverpkg=github.com/ONSdigital/dp-files-api/... -coverprofile=component-coverage.out
	gocovmerge coverage.out component-coverage.out > combined-coverage.out
	go tool cover -html=combined-coverage.out
