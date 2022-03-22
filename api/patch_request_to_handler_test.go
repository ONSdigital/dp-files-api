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
	States               []patchRequestMetadataStates
	PatchRequestHandlers api.PatchRequestHandlers
}

func (suite *PatchRequestToHandlerSuite) SetupTest() {
	collectionID := "123456"
	stateDecrypted := "DECRYPTED"
	statePublished := "PUBLISHED"
	stateUploaded := "UPLOADED"
	collectionUpdateHandlerBody := "collectionUpdateHandler"
	decryptedHandlerBody := "decryptedHandler"
	publishedHandlerBody := "publishedHandler"
	uploadCompleteHandlerBody := "uploadCompleteHandler"

	suite.States = []patchRequestMetadataStates{
		{Metadata: api.StateMetadata{CollectionID: &collectionID}, ExpectedBody: collectionUpdateHandlerBody},
		{Metadata: api.StateMetadata{State: &stateDecrypted}, ExpectedBody: decryptedHandlerBody},
		{Metadata: api.StateMetadata{State: &statePublished}, ExpectedBody: publishedHandlerBody},
		{Metadata: api.StateMetadata{State: &stateUploaded}, ExpectedBody: uploadCompleteHandlerBody},
	}

	generatePatchRequestHandler := func(body string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(body)) }
	}

	suite.PatchRequestHandlers = api.PatchRequestHandlers{
		UploadComplete:   generatePatchRequestHandler(uploadCompleteHandlerBody),
		Published:        generatePatchRequestHandler(publishedHandlerBody),
		Decrypted:        generatePatchRequestHandler(decryptedHandlerBody),
		CollectionUpdate: generatePatchRequestHandler(collectionUpdateHandlerBody),
	}
}

func (suite *PatchRequestToHandlerSuite) TestPatchRequestToHandlerReturnsCorrectHandler() {
	patchRequestHandler := api.PatchRequestToHandler(suite.PatchRequestHandlers)

	for _, state := range suite.States {
		body, _ := json.Marshal(state.Metadata)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/files/test.txt", bytes.NewBuffer(body))

		patchRequestHandler.ServeHTTP(w, req)

		actualBody, _ := ioutil.ReadAll(w.Body)

		suite.Equal(state.ExpectedBody, string(actualBody))
	}
}

func (suite *PatchRequestToHandlerSuite) TestPatchRequestToHandlerPassesBodyToSubsequentHandler() {
	for _, state := range suite.States {
		var actualRequestBody []byte
		expectedRequestBody, _ := json.Marshal(state.Metadata)

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actualRequestBody, _ = ioutil.ReadAll(r.Body)
		})

		patchRequestHandlers := api.PatchRequestHandlers{
			UploadComplete:   testHandler,
			Published:        testHandler,
			Decrypted:        testHandler,
			CollectionUpdate: testHandler,
		}

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/files/text.txt", bytes.NewBuffer(expectedRequestBody))

		actualHandler := api.PatchRequestToHandler(patchRequestHandlers)
		actualHandler.ServeHTTP(w, req)

		msg := fmt.Sprintf(`Expected "%s" to equal "%s"`, actualRequestBody, expectedRequestBody)

		suite.Equal(expectedRequestBody, actualRequestBody, msg)
	}
}

func TestPatchRequestToHandlerSuite(t *testing.T) {
	suite.Run(t, new(PatchRequestToHandlerSuite))
}
