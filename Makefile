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
	golangci-lint run ./...

.PHONY: build
build:
	go build -tags 'production' $(LDFLAGS) -o $(BINPATH)/dp-files-api

.PHONY: debug
debug:
	go build -tags 'debug' $(LDFLAGS) -o $(BINPATH)/dp-files-api
	HUMAN_LOG=1 DEBUG=1 $(BINPATH)/dp-files-api

.PHONY: generate-swagger
generate-swagger:
	swag i -g service/service.go

.PHONY: test
test:
	go test -race -cover ./...

.PHONY: convey
convey:
	goconvey ./...

.PHONY: test-component
test-component:
	go test -cover -coverpkg=github.com/ONSdigital/dp-files-api/... -component
	go test ./files -component

.PHONY: docker-test
docker-test-component:
	docker-compose -f docker-compose-services.yml -f docker-compose.yml down
	docker build -f Dockerfile . -t template_test --target=test
	docker-compose -f docker-compose-services.yml -f docker-compose.yml up -d
	docker-compose -f docker-compose-services.yml -f docker-compose.yml exec -T dp-files-api make test-component
	docker-compose -f docker-compose-services.yml -f docker-compose.yml down

.PHONY: test-coverage
test-coverage:
	rm combined-coverage.out component-coverage.out coverage.out
	go test -cover ./... -coverprofile=coverage.out
	go test -component -cover -coverpkg=github.com/ONSdigital/dp-files-api/... -coverprofile=component-coverage.out
	gocovmerge coverage.out component-coverage.out > combined-coverage.out
	go tool cover -html=combined-coverage.out

docker-local:
	docker-compose  -f docker-compose-services.yml -f docker-compose-local.yml down
	docker-compose  -f docker-compose-services.yml -f docker-compose-local.yml up -d
	docker-compose  -f docker-compose-services.yml -f docker-compose-local.yml exec dp-files-api bash