module github.com/ONSdigital/dp-files-api

go 1.17

replace github.com/coreos/etcd => github.com/coreos/etcd v3.3.24+incompatible

require (
	github.com/ONSdigital/dp-component-test v0.6.3
	github.com/ONSdigital/dp-healthcheck v1.2.1
	github.com/ONSdigital/dp-mongodb/v3 v3.0.0-beta.5
	github.com/ONSdigital/dp-net v1.2.0
	github.com/ONSdigital/dp-net/v2 v2.2.0-beta
	github.com/ONSdigital/log.go/v2 v2.0.9
	github.com/cucumber/godog v0.12.2
	github.com/gorilla/mux v1.8.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/pkg/errors v0.9.1
	github.com/rdumont/assistdog v0.0.0-20201106100018-168b06230d14
	github.com/smartystreets/goconvey v1.7.2
	github.com/stretchr/testify v1.7.0
	go.mongodb.org/mongo-driver v1.8.1
)

require (
	github.com/ONSdigital/dp-api-clients-go v1.43.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.0
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/go-playground/validator v9.31.0+incompatible
	github.com/leodido/go-urn v1.2.1 // indirect
	golang.org/x/crypto v0.0.0-20211209193657-4570a0811e8b // indirect
	golang.org/x/net v0.0.0-20211209124913-491a49abca63 // indirect
	golang.org/x/sys v0.0.0-20211210111614-af8b64212486 // indirect
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
)
