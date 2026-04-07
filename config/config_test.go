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
				So(testCfg.ClusterEndpoint, ShouldEqual, "localhost:27017")
				So(testCfg.Database, ShouldEqual, "files")
				So(testCfg.Collections, ShouldResemble, map[string]string{MetadataCollection: "metadata", CollectionsCollection: "collections", BundlesCollection: "bundles", FileEventsCollection: "file_events"})
				So(testCfg.IsStrongReadConcernEnabled, ShouldEqual, false)
				So(testCfg.IsWriteConcernMajorityEnabled, ShouldEqual, true)
				So(testCfg.ConnectTimeout, ShouldEqual, 5*time.Second)
				So(testCfg.QueryTimeout, ShouldEqual, 15*time.Second)
				So(testCfg.IsSSL, ShouldEqual, false)
				So(testCfg.Addr, ShouldResemble, []string{"kafka:9092"})
				So(testCfg.ProducerMinBrokersHealthy, ShouldEqual, 1)
				So(testCfg.Version, ShouldEqual, "2.6.1")
				So(testCfg.MaxBytes, ShouldEqual, 2000000)
				So(testCfg.SecProtocol, ShouldEqual, "")
				So(testCfg.SecCACerts, ShouldEqual, "")
				So(testCfg.SecClientKey, ShouldEqual, "")
				So(testCfg.SecClientCert, ShouldEqual, "")
				So(testCfg.SecSkipVerify, ShouldEqual, false)
				So(testCfg.StaticFilePublishedTopic, ShouldEqual, "static-file-published-v2")
				So(testCfg.Enabled, ShouldEqual, false)
				So(testCfg.PermissionsAPIURL, ShouldEqual, "http://localhost:25400")
				So(testCfg.IdentityWebKeySetURL, ShouldEqual, "http://localhost:25600")
				So(testCfg.PermissionsCacheUpdateInterval, ShouldEqual, time.Minute*1)
				So(testCfg.PermissionsMaxCacheTime, ShouldEqual, time.Minute*5)
				So(testCfg.IdentityClientMaxRetries, ShouldEqual, 2)
				So(testCfg.ZebedeeURL, ShouldEqual, "http://localhost:8082")
			})

			Convey("Then a second call to config should return the same config", func() {
				newCfg, newErr := Get()
				So(newErr, ShouldBeNil)
				So(newCfg, ShouldResemble, testCfg)
			})
		})
	})
}
