package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-files-api/api"
	"github.com/stretchr/testify/suite"
)

type patchRequestMetadataStates struct {
	Metadata     api.StateMetadata
	ExpectedBody string
}

type PatchRequestToHandlerSuite struct {
	suite.Suite
	TestStates           []patchRequestMetadataStates
	PatchRequestHandlers api.PatchRequestHandlers
}

func (s *PatchRequestToHandlerSuite) SetupTest() {
	collectionID := "123456"
	bundleID := "bundleID"
	stateMoved := "MOVED"
	statePublished := "PUBLISHED"
	stateUploaded := "UPLOADED"
	collectionUpdateHandlerBody := "collectionUpdateHandler"
	bundleUpdateHandlerBody := "bundleUpdateHandler"
	movedHandlerBody := "movedHandler"
	publishedHandlerBody := "publishedHandler"
	uploadCompleteHandlerBody := "uploadCompleteHandler"

	generatePatchRequestHandler := func(body string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(body)) }
	}

	s.TestStates = []patchRequestMetadataStates{
		{Metadata: api.StateMetadata{CollectionID: &collectionID}, ExpectedBody: collectionUpdateHandlerBody},
		{Metadata: api.StateMetadata{BundleID: &bundleID}, ExpectedBody: bundleUpdateHandlerBody},
		{Metadata: api.StateMetadata{State: &stateMoved}, ExpectedBody: movedHandlerBody},
		{Metadata: api.StateMetadata{State: &statePublished}, ExpectedBody: publishedHandlerBody},
		{Metadata: api.StateMetadata{State: &stateUploaded}, ExpectedBody: uploadCompleteHandlerBody},
	}

	s.PatchRequestHandlers = api.PatchRequestHandlers{
		UploadComplete:   generatePatchRequestHandler(uploadCompleteHandlerBody),
		Published:        generatePatchRequestHandler(publishedHandlerBody),
		Moved:            generatePatchRequestHandler(movedHandlerBody),
		CollectionUpdate: generatePatchRequestHandler(collectionUpdateHandlerBody),
		BundleUpdate:     generatePatchRequestHandler(bundleUpdateHandlerBody),
	}
}

func (s *PatchRequestToHandlerSuite) TestPatchRequestToHandlerReturnsCorrectHandler() {
	patchRequestHandler := api.PatchRequestToHandler(s.PatchRequestHandlers)

	for _, testState := range s.TestStates {
		body, _ := json.Marshal(testState.Metadata)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPatch, "/files/test.txt", bytes.NewBuffer(body))

		patchRequestHandler.ServeHTTP(w, req)

		actualBody, _ := io.ReadAll(w.Body)

		s.Equal(testState.ExpectedBody, string(actualBody))
	}
}

func (s *PatchRequestToHandlerSuite) TestPatchRequestToHandlerPassesBodyToSubsequentHandler() {
	var actualRequestBody []byte

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualRequestBody, _ = io.ReadAll(r.Body)
	})

	patchRequestHandlers := api.PatchRequestHandlers{
		UploadComplete:   testHandler,
		Published:        testHandler,
		Moved:            testHandler,
		CollectionUpdate: testHandler,
		BundleUpdate:     testHandler,
	}

	for _, state := range s.TestStates {
		expectedRequestBody, _ := json.Marshal(state.Metadata)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPatch, "/files/text.txt", bytes.NewBuffer(expectedRequestBody))

		actualHandler := api.PatchRequestToHandler(patchRequestHandlers)
		actualHandler.ServeHTTP(w, req)

		msg := fmt.Sprintf(`Expected %q to equal %q`, actualRequestBody, expectedRequestBody)

		s.Equal(expectedRequestBody, actualRequestBody, msg)
	}
}

func TestPatchRequestToHandlerSuite(t *testing.T) {
	suite.Run(t, new(PatchRequestToHandlerSuite))
}
