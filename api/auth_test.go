package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	clientsidentity "github.com/ONSdigital/dp-api-clients-go/v2/identity"
	authMock "github.com/ONSdigital/dp-authorisation/v2/authorisation/mock"
	"github.com/ONSdigital/dp-files-api/config"
	dphttp "github.com/ONSdigital/dp-net/v3/http"
	dprequest "github.com/ONSdigital/dp-net/v3/request"
	permissionsAPISDK "github.com/ONSdigital/dp-permissions-api/sdk"
	"github.com/stretchr/testify/assert"
)

var testServiceIdentityResponse = &dprequest.IdentityResponse{
	Identifier: "service-1",
}

func newMockHTTPClient(retCode int, retBody interface{}) *dphttp.ClienterMock {
	return &dphttp.ClienterMock{
		SetPathsWithNoRetriesFunc: func(paths []string) {},
		GetPathsWithNoRetriesFunc: func() []string {
			return []string{"/healthcheck"}
		},
		DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			body, _ := json.Marshal(retBody)
			return &http.Response{
				StatusCode: retCode,
				Body:       io.NopCloser(bytes.NewReader(body)),
			}, nil
		},
	}
}

func TestAuthInvalidToken(t *testing.T) {
	cfg, _ := config.Get()

	testIdentityClient := clientsidentity.New(cfg.ZebedeeURL)

	authorisationMock := &authMock.MiddlewareMock{
		ParseFunc: func(token string) (*permissionsAPISDK.EntityData, error) {
			return nil, errors.New("parse error")
		},
	}

	authEntityData, err := getAuthEntityData(context.Background(), authorisationMock, testIdentityClient, "invalid.token", nil)
	assert.NotNil(t, err)
	assert.Nil(t, authEntityData)
}

func TestAuthServiceToken(t *testing.T) {
	cfg, _ := config.Get()

	httpClient := newMockHTTPClient(200, testServiceIdentityResponse)
	testIdentityClient := clientsidentity.NewWithHealthClient(healthcheck.NewClientWithClienter("", cfg.ZebedeeURL, httpClient))

	authorisationMock := &authMock.MiddlewareMock{}

	authEntityData, err := getAuthEntityData(context.Background(), authorisationMock, testIdentityClient, "valid-test-service-auth", nil)
	assert.Nil(t, err)
	assert.NotNil(t, authEntityData)
	assert.Equal(t, &AuthEntityData{EntityData: &permissionsAPISDK.EntityData{UserID: "service-1"}, IsServiceAuth: true}, authEntityData)
}
