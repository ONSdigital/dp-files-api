// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	auth "github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/ONSdigital/dp-files-api/aws"
	"github.com/ONSdigital/dp-files-api/clock"
	"github.com/ONSdigital/dp-files-api/files"
	"github.com/ONSdigital/dp-files-api/health"
	"github.com/ONSdigital/dp-files-api/mongo"
	"github.com/ONSdigital/dp-files-api/service"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"sync"
)

// Ensure, that ServiceContainerMock does implement service.ServiceContainer.
// If this is not the case, regenerate this file with moq.
var _ service.ServiceContainer = &ServiceContainerMock{}

// ServiceContainerMock is a mock implementation of service.ServiceContainer.
//
// 	func TestSomethingThatUsesServiceContainer(t *testing.T) {
//
// 		// make and configure a mocked service.ServiceContainer
// 		mockedServiceContainer := &ServiceContainerMock{
// 			GetAuthMiddlewareFunc: func() auth.Middleware {
// 				panic("mock out the GetAuthMiddleware method")
// 			},
// 			GetClockFunc: func() clock.Clock {
// 				panic("mock out the GetClock method")
// 			},
// 			GetHTTPServerFunc: func() files.HTTPServer {
// 				panic("mock out the GetHTTPServer method")
// 			},
// 			GetHealthCheckFunc: func() health.Checker {
// 				panic("mock out the GetHealthCheck method")
// 			},
// 			GetKafkaProducerFunc: func() kafka.IProducer {
// 				panic("mock out the GetKafkaProducer method")
// 			},
// 			GetMongoDBFunc: func() mongo.Client {
// 				panic("mock out the GetMongoDB method")
// 			},
// 			GetS3ClienterFunc: func() aws.S3Clienter {
// 				panic("mock out the GetS3Clienter method")
// 			},
// 			ShutdownFunc: func(ctx context.Context) error {
// 				panic("mock out the Shutdown method")
// 			},
// 		}
//
// 		// use mockedServiceContainer in code that requires service.ServiceContainer
// 		// and then make assertions.
//
// 	}
type ServiceContainerMock struct {
	// GetAuthMiddlewareFunc mocks the GetAuthMiddleware method.
	GetAuthMiddlewareFunc func() auth.Middleware

	// GetClockFunc mocks the GetClock method.
	GetClockFunc func() clock.Clock

	// GetHTTPServerFunc mocks the GetHTTPServer method.
	GetHTTPServerFunc func() files.HTTPServer

	// GetHealthCheckFunc mocks the GetHealthCheck method.
	GetHealthCheckFunc func() health.Checker

	// GetKafkaProducerFunc mocks the GetKafkaProducer method.
	GetKafkaProducerFunc func() kafka.IProducer

	// GetMongoDBFunc mocks the GetMongoDB method.
	GetMongoDBFunc func() mongo.Client

	// GetS3ClienterFunc mocks the GetS3Clienter method.
	GetS3ClienterFunc func() aws.S3Clienter

	// ShutdownFunc mocks the Shutdown method.
	ShutdownFunc func(ctx context.Context) error

	// calls tracks calls to the methods.
	calls struct {
		// GetAuthMiddleware holds details about calls to the GetAuthMiddleware method.
		GetAuthMiddleware []struct {
		}
		// GetClock holds details about calls to the GetClock method.
		GetClock []struct {
		}
		// GetHTTPServer holds details about calls to the GetHTTPServer method.
		GetHTTPServer []struct {
		}
		// GetHealthCheck holds details about calls to the GetHealthCheck method.
		GetHealthCheck []struct {
		}
		// GetKafkaProducer holds details about calls to the GetKafkaProducer method.
		GetKafkaProducer []struct {
		}
		// GetMongoDB holds details about calls to the GetMongoDB method.
		GetMongoDB []struct {
		}
		// GetS3Clienter holds details about calls to the GetS3Clienter method.
		GetS3Clienter []struct {
		}
		// Shutdown holds details about calls to the Shutdown method.
		Shutdown []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
	}
	lockGetAuthMiddleware sync.RWMutex
	lockGetClock          sync.RWMutex
	lockGetHTTPServer     sync.RWMutex
	lockGetHealthCheck    sync.RWMutex
	lockGetKafkaProducer  sync.RWMutex
	lockGetMongoDB        sync.RWMutex
	lockGetS3Clienter     sync.RWMutex
	lockShutdown          sync.RWMutex
}

// GetAuthMiddleware calls GetAuthMiddlewareFunc.
func (mock *ServiceContainerMock) GetAuthMiddleware() auth.Middleware {
	if mock.GetAuthMiddlewareFunc == nil {
		panic("ServiceContainerMock.GetAuthMiddlewareFunc: method is nil but ServiceContainer.GetAuthMiddleware was just called")
	}
	callInfo := struct {
	}{}
	mock.lockGetAuthMiddleware.Lock()
	mock.calls.GetAuthMiddleware = append(mock.calls.GetAuthMiddleware, callInfo)
	mock.lockGetAuthMiddleware.Unlock()
	return mock.GetAuthMiddlewareFunc()
}

// GetAuthMiddlewareCalls gets all the calls that were made to GetAuthMiddleware.
// Check the length with:
//     len(mockedServiceContainer.GetAuthMiddlewareCalls())
func (mock *ServiceContainerMock) GetAuthMiddlewareCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockGetAuthMiddleware.RLock()
	calls = mock.calls.GetAuthMiddleware
	mock.lockGetAuthMiddleware.RUnlock()
	return calls
}

// GetClock calls GetClockFunc.
func (mock *ServiceContainerMock) GetClock() clock.Clock {
	if mock.GetClockFunc == nil {
		panic("ServiceContainerMock.GetClockFunc: method is nil but ServiceContainer.GetClock was just called")
	}
	callInfo := struct {
	}{}
	mock.lockGetClock.Lock()
	mock.calls.GetClock = append(mock.calls.GetClock, callInfo)
	mock.lockGetClock.Unlock()
	return mock.GetClockFunc()
}

// GetClockCalls gets all the calls that were made to GetClock.
// Check the length with:
//     len(mockedServiceContainer.GetClockCalls())
func (mock *ServiceContainerMock) GetClockCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockGetClock.RLock()
	calls = mock.calls.GetClock
	mock.lockGetClock.RUnlock()
	return calls
}

// GetHTTPServer calls GetHTTPServerFunc.
func (mock *ServiceContainerMock) GetHTTPServer() files.HTTPServer {
	if mock.GetHTTPServerFunc == nil {
		panic("ServiceContainerMock.GetHTTPServerFunc: method is nil but ServiceContainer.GetHTTPServer was just called")
	}
	callInfo := struct {
	}{}
	mock.lockGetHTTPServer.Lock()
	mock.calls.GetHTTPServer = append(mock.calls.GetHTTPServer, callInfo)
	mock.lockGetHTTPServer.Unlock()
	return mock.GetHTTPServerFunc()
}

// GetHTTPServerCalls gets all the calls that were made to GetHTTPServer.
// Check the length with:
//     len(mockedServiceContainer.GetHTTPServerCalls())
func (mock *ServiceContainerMock) GetHTTPServerCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockGetHTTPServer.RLock()
	calls = mock.calls.GetHTTPServer
	mock.lockGetHTTPServer.RUnlock()
	return calls
}

// GetHealthCheck calls GetHealthCheckFunc.
func (mock *ServiceContainerMock) GetHealthCheck() health.Checker {
	if mock.GetHealthCheckFunc == nil {
		panic("ServiceContainerMock.GetHealthCheckFunc: method is nil but ServiceContainer.GetHealthCheck was just called")
	}
	callInfo := struct {
	}{}
	mock.lockGetHealthCheck.Lock()
	mock.calls.GetHealthCheck = append(mock.calls.GetHealthCheck, callInfo)
	mock.lockGetHealthCheck.Unlock()
	return mock.GetHealthCheckFunc()
}

// GetHealthCheckCalls gets all the calls that were made to GetHealthCheck.
// Check the length with:
//     len(mockedServiceContainer.GetHealthCheckCalls())
func (mock *ServiceContainerMock) GetHealthCheckCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockGetHealthCheck.RLock()
	calls = mock.calls.GetHealthCheck
	mock.lockGetHealthCheck.RUnlock()
	return calls
}

// GetKafkaProducer calls GetKafkaProducerFunc.
func (mock *ServiceContainerMock) GetKafkaProducer() kafka.IProducer {
	if mock.GetKafkaProducerFunc == nil {
		panic("ServiceContainerMock.GetKafkaProducerFunc: method is nil but ServiceContainer.GetKafkaProducer was just called")
	}
	callInfo := struct {
	}{}
	mock.lockGetKafkaProducer.Lock()
	mock.calls.GetKafkaProducer = append(mock.calls.GetKafkaProducer, callInfo)
	mock.lockGetKafkaProducer.Unlock()
	return mock.GetKafkaProducerFunc()
}

// GetKafkaProducerCalls gets all the calls that were made to GetKafkaProducer.
// Check the length with:
//     len(mockedServiceContainer.GetKafkaProducerCalls())
func (mock *ServiceContainerMock) GetKafkaProducerCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockGetKafkaProducer.RLock()
	calls = mock.calls.GetKafkaProducer
	mock.lockGetKafkaProducer.RUnlock()
	return calls
}

// GetMongoDB calls GetMongoDBFunc.
func (mock *ServiceContainerMock) GetMongoDB() mongo.Client {
	if mock.GetMongoDBFunc == nil {
		panic("ServiceContainerMock.GetMongoDBFunc: method is nil but ServiceContainer.GetMongoDB was just called")
	}
	callInfo := struct {
	}{}
	mock.lockGetMongoDB.Lock()
	mock.calls.GetMongoDB = append(mock.calls.GetMongoDB, callInfo)
	mock.lockGetMongoDB.Unlock()
	return mock.GetMongoDBFunc()
}

// GetMongoDBCalls gets all the calls that were made to GetMongoDB.
// Check the length with:
//     len(mockedServiceContainer.GetMongoDBCalls())
func (mock *ServiceContainerMock) GetMongoDBCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockGetMongoDB.RLock()
	calls = mock.calls.GetMongoDB
	mock.lockGetMongoDB.RUnlock()
	return calls
}

// GetS3Clienter calls GetS3ClienterFunc.
func (mock *ServiceContainerMock) GetS3Clienter() aws.S3Clienter {
	if mock.GetS3ClienterFunc == nil {
		panic("ServiceContainerMock.GetS3ClienterFunc: method is nil but ServiceContainer.GetS3Clienter was just called")
	}
	callInfo := struct {
	}{}
	mock.lockGetS3Clienter.Lock()
	mock.calls.GetS3Clienter = append(mock.calls.GetS3Clienter, callInfo)
	mock.lockGetS3Clienter.Unlock()
	return mock.GetS3ClienterFunc()
}

// GetS3ClienterCalls gets all the calls that were made to GetS3Clienter.
// Check the length with:
//     len(mockedServiceContainer.GetS3ClienterCalls())
func (mock *ServiceContainerMock) GetS3ClienterCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockGetS3Clienter.RLock()
	calls = mock.calls.GetS3Clienter
	mock.lockGetS3Clienter.RUnlock()
	return calls
}

// Shutdown calls ShutdownFunc.
func (mock *ServiceContainerMock) Shutdown(ctx context.Context) error {
	if mock.ShutdownFunc == nil {
		panic("ServiceContainerMock.ShutdownFunc: method is nil but ServiceContainer.Shutdown was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockShutdown.Lock()
	mock.calls.Shutdown = append(mock.calls.Shutdown, callInfo)
	mock.lockShutdown.Unlock()
	return mock.ShutdownFunc(ctx)
}

// ShutdownCalls gets all the calls that were made to Shutdown.
// Check the length with:
//     len(mockedServiceContainer.ShutdownCalls())
func (mock *ServiceContainerMock) ShutdownCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockShutdown.RLock()
	calls = mock.calls.Shutdown
	mock.lockShutdown.RUnlock()
	return calls
}
