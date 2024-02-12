package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ONSdigital/dp-files-api/api"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
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

func (suite *PatchRequestToHandlerSuite) SetupTest() {
	collectionID := "123456"
	statePublished := "PUBLISHED"
	stateUploaded := "UPLOADED"
	collectionUpdateHandlerBody := "collectionUpdateHandler"
	publishedHandlerBody := "publishedHandler"
	uploadCompleteHandlerBody := "uploadCompleteHandler"

	generatePatchRequestHandler := func(body string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(body)) }
	}

	suite.TestStates = []patchRequestMetadataStates{
		{Metadata: api.StateMetadata{CollectionID: &collectionID}, ExpectedBody: collectionUpdateHandlerBody},
		{Metadata: api.StateMetadata{State: &statePublished}, ExpectedBody: publishedHandlerBody},
		{Metadata: api.StateMetadata{State: &stateUploaded}, ExpectedBody: uploadCompleteHandlerBody},
	}

	suite.PatchRequestHandlers = api.PatchRequestHandlers{
		UploadComplete:   generatePatchRequestHandler(uploadCompleteHandlerBody),
		Published:        generatePatchRequestHandler(publishedHandlerBody),
		CollectionUpdate: generatePatchRequestHandler(collectionUpdateHandlerBody),
	}
}

func (suite *PatchRequestToHandlerSuite) TestPatchRequestToHandlerReturnsCorrectHandler() {
	patchRequestHandler := api.PatchRequestToHandler(suite.PatchRequestHandlers)

	for _, testState := range suite.TestStates {
		body, _ := json.Marshal(testState.Metadata)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPatch, "/files/test.txt", bytes.NewBuffer(body))

		patchRequestHandler.ServeHTTP(w, req)

		actualBody, _ := ioutil.ReadAll(w.Body)

		suite.Equal(testState.ExpectedBody, string(actualBody))
	}
}

func (suite *PatchRequestToHandlerSuite) TestPatchRequestToHandlerPassesBodyToSubsequentHandler() {
	var actualRequestBody []byte

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualRequestBody, _ = ioutil.ReadAll(r.Body)
	})

	patchRequestHandlers := api.PatchRequestHandlers{
		UploadComplete:   testHandler,
		Published:        testHandler,
		CollectionUpdate: testHandler,
	}

	for _, state := range suite.TestStates {
		expectedRequestBody, _ := json.Marshal(state.Metadata)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPatch, "/files/text.txt", bytes.NewBuffer(expectedRequestBody))

		actualHandler := api.PatchRequestToHandler(patchRequestHandlers)
		actualHandler.ServeHTTP(w, req)

		msg := fmt.Sprintf(`Expected "%s" to equal "%s"`, actualRequestBody, expectedRequestBody)

		suite.Equal(expectedRequestBody, actualRequestBody, msg)
	}
}

func TestPatchRequestToHandlerSuite(t *testing.T) {
	suite.Run(t, new(PatchRequestToHandlerSuite))
}
