package config

import (
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfig(t *testing.T) {
	os.Clearenv()
	testCfg, err := Get()
	Convey("Given an environment with no environment variables set", t, func() {
		Convey("When the config values are retrieved", func() {
			Convey("Then testCfg should not be nil", func() {
				So(testCfg, ShouldNotBeNil)
			})

			Convey("Then there should be no error returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the values should be set to the expected defaults", func() {
				So(testCfg.BindAddr, ShouldEqual, "localhost:26900")
				So(testCfg.AwsRegion, ShouldEqual, "eu-west-2")
				So(testCfg.PrivateBucketName, ShouldEqual, "testing")
				So(testCfg.LocalstackHost, ShouldEqual, "http://127.0.0.1:4566")
				So(testCfg.GracefulShutdownTimeout, ShouldEqual, 5*time.Second)
				So(testCfg.HealthCheckInterval, ShouldEqual, 30*time.Second)
				So(testCfg.HealthCheckCriticalTimeout, ShouldEqual, 90*time.Second)
				So(testCfg.IsPublishing, ShouldBeFalse)
				So(testCfg.MaxNumBatches, ShouldEqual, 5)
				So(testCfg.MinBatchSize, ShouldEqual, 20)
				So(testCfg.MongoConfig.ClusterEndpoint, ShouldEqual, "localhost:27017")
				So(testCfg.MongoConfig.Database, ShouldEqual, "files")
				So(testCfg.MongoConfig.Collections, ShouldResemble, map[string]string{MetadataCollection: "metadata", CollectionsCollection: "collections", BundlesCollection: "bundles"})
				So(testCfg.MongoConfig.IsStrongReadConcernEnabled, ShouldEqual, false)
				So(testCfg.MongoConfig.IsWriteConcernMajorityEnabled, ShouldEqual, true)
				So(testCfg.MongoConfig.ConnectTimeout, ShouldEqual, 5*time.Second)
				So(testCfg.MongoConfig.QueryTimeout, ShouldEqual, 15*time.Second)
				So(testCfg.MongoConfig.TLSConnectionConfig.IsSSL, ShouldEqual, false)
				So(testCfg.KafkaConfig.Addr, ShouldResemble, []string{"kafka:9092"})
				So(testCfg.KafkaConfig.ProducerMinBrokersHealthy, ShouldEqual, 1)
				So(testCfg.KafkaConfig.Version, ShouldEqual, "2.6.1")
				So(testCfg.KafkaConfig.MaxBytes, ShouldEqual, 2000000)
				So(testCfg.KafkaConfig.SecProtocol, ShouldEqual, "")
				So(testCfg.KafkaConfig.SecCACerts, ShouldEqual, "")
				So(testCfg.KafkaConfig.SecClientKey, ShouldEqual, "")
				So(testCfg.KafkaConfig.SecClientCert, ShouldEqual, "")
				So(testCfg.KafkaConfig.SecSkipVerify, ShouldEqual, false)
				So(testCfg.KafkaConfig.StaticFilePublishedTopic, ShouldEqual, "static-file-published-v2")
				So(testCfg.AuthConfig.Enabled, ShouldEqual, true)
				So(testCfg.AuthConfig.PermissionsAPIURL, ShouldEqual, "http://localhost:25400")
				So(testCfg.AuthConfig.IdentityWebKeySetURL, ShouldEqual, "http://localhost:25600")
				So(testCfg.AuthConfig.PermissionsCacheUpdateInterval, ShouldEqual, time.Minute*5)
				So(testCfg.AuthConfig.PermissionsMaxCacheTime, ShouldEqual, time.Minute*15)
				So(testCfg.AuthConfig.IdentityClientMaxRetries, ShouldEqual, 2)
				So(testCfg.AuthConfig.ZebedeeURL, ShouldEqual, "http://localhost:8082")
			})

			Convey("Then a second call to config should return the same config", func() {
				newCfg, newErr := Get()
				So(newErr, ShouldBeNil)
				So(newCfg, ShouldResemble, testCfg)
			})
		})
	})
}
