package senders

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"github.com/CDCgov/reportstream-sftp-ingestion/mocks"
	"github.com/CDCgov/reportstream-sftp-ingestion/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

const reportStreamPrivateKeyName = "ca-phl-reportstream-private-key-local"

type SenderTestSuite struct {
	suite.Suite
}

func (suite *SenderTestSuite) SetupTest() {
	os.Setenv("ENV", "local")
	os.Setenv("REPORT_STREAM_URL_PREFIX", "rs.com")
	os.Setenv("CA_PHL_CLIENT_NAME", "client")
}

func (suite *SenderTestSuite) TearDownTest() {
	os.Unsetenv("ENV")
	os.Unsetenv("REPORT_STREAM_URL_PREFIX")
	os.Unsetenv("CA_PHL_CLIENT_NAME")
}

func (suite *SenderTestSuite) Test_NewSender_VariablesAreSet_ReturnsSender() {

	sender, err := NewSender()

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), os.Getenv("REPORT_STREAM_URL_PREFIX"), sender.baseUrl)
	assert.Equal(suite.T(), os.Getenv("CA_PHL_CLIENT_NAME"), sender.clientName)
}

func (suite *SenderTestSuite) Test_NewSender_EnvIsEmpty_ReturnsSenderWithLocalCredentials() {
	os.Setenv("ENV", "")
	sender, err := NewSender()

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), os.Getenv("REPORT_STREAM_URL_PREFIX"), sender.baseUrl)
	assert.Equal(suite.T(), os.Getenv("CA_PHL_CLIENT_NAME"), sender.clientName)
}

func (suite *SenderTestSuite) Test_GenerateJWT_ReturnsJWT() {
	sender, err := NewSender()

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	sender.credentialGetter = mockCredentialGetter

	testKey, err := rsa.GenerateKey(rand.Reader, 2048)
	mockCredentialGetter.On("GetPrivateKey", reportStreamPrivateKeyName).Return(testKey, nil)
	jwt, err := sender.generateJwt()

	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), jwt)
}

func (suite *SenderTestSuite) Test_GenerateJWT_UnableToGetPrivateKey_ReturnsError() {

	sender, err := NewSender()

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	sender.credentialGetter = mockCredentialGetter

	mockCredentialGetter.On("GetPrivateKey", reportStreamPrivateKeyName).Return(&rsa.PrivateKey{}, errors.New("failed to retrieve private key"))
	_, err = sender.generateJwt()

	assert.Error(suite.T(), err)
}

func (suite *SenderTestSuite) Test_getToken_ReturnsAccessToken() {
	sender, err := NewSender()

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	sender.credentialGetter = mockCredentialGetter

	testKey, err := rsa.GenerateKey(rand.Reader, 2048)
	mockCredentialGetter.On("GetPrivateKey", reportStreamPrivateKeyName).Return(testKey, nil)

	// Set up a test server for ReportStream
	// Response parts: Body, Status Code, Access Token (part of body), Error (part of body)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
    "sub": "flexion.*.report_e6b68103-dd38-420e-8118-2b2f6c9fa3c4",
    "access_token": "eyJhbGciOiJIUzM4NCJ9.eyJleHAiOjE3MTk1MjcyNzgsInNjb3BlIjoiZmxleGlvbi4qLnJlcG9ydCIsInN1YiI6ImZsZXhpb24uKi5yZXBvcnRfZTZiNjgxMDMtZGQzOC00MjBlLTgxMTgtMmIyZjZjOWZhM2M0In0.liHv9SJYxztgMmCPKGIF2lzcMMMzFAoatLlIC33uz5jbA5wSJa8iIa5yzJ1ZaECI",
    "token_type": "bearer",
    "expires_in": 300,
    "expires_at_seconds": 1719527278,
    "scope": "flexion.*.report"
}
`))
	}))
	defer server.Close()

	sender.baseUrl = server.URL
	token, err := sender.getToken()

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), token)
}

func (suite *SenderTestSuite) Test_getToken_UnableToGenerateJWT_ReturnsError() {
	sender, err := NewSender()

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	sender.credentialGetter = mockCredentialGetter

	mockCredentialGetter.On("GetPrivateKey", reportStreamPrivateKeyName).Return(&rsa.PrivateKey{}, errors.New("failed to retrieve private key"))
	token, err := sender.getToken()

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "", token)
}

func (suite *SenderTestSuite) Test_getToken_UnableToCallTokenEndpoint_ReturnsError() {
	os.Setenv("REPORT_STREAM_URL_PREFIX", "this is not a good URL")
	sender, err := NewSender()

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	sender.credentialGetter = mockCredentialGetter

	testKey, err := rsa.GenerateKey(rand.Reader, 2048)
	mockCredentialGetter.On("GetPrivateKey", reportStreamPrivateKeyName).Return(testKey, nil)

	token, err := sender.getToken()

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "", token)
}

func (suite *SenderTestSuite) Test_getToken_ReportStreamResponseStatusIsInvalid_ReturnsError() {
	sender, err := NewSender()

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	sender.credentialGetter = mockCredentialGetter

	testKey, err := rsa.GenerateKey(rand.Reader, 2048)
	mockCredentialGetter.On("GetPrivateKey", reportStreamPrivateKeyName).Return(testKey, nil)

	// Set up a test server for ReportStream
	// Response parts: Body, Status Code, Access Token (part of body), Error (part of body)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{
    "sub": "flexion.*.report_e6b68103-dd38-420e-8118-2b2f6c9fa3c4",
    "access_token": "eyJhbGciOiJIUzM4NCJ9.eyJleHAiOjE3MTk1MjcyNzgsInNjb3BlIjoiZmxleGlvbi4qLnJlcG9ydCIsInN1YiI6ImZsZXhpb24uKi5yZXBvcnRfZTZiNjgxMDMtZGQzOC00MjBlLTgxMTgtMmIyZjZjOWZhM2M0In0.liHv9SJYxztgMmCPKGIF2lzcMMMzFAoatLlIC33uz5jbA5wSJa8iIa5yzJ1ZaECI",
    "token_type": "bearer",
    "expires_in": 300,
    "expires_at_seconds": 1719527278,
    "scope": "flexion.*.report"
}
`))
	}))
	defer server.Close()

	sender.baseUrl = server.URL
	token, err := sender.getToken()

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "", token)
}

func (suite *SenderTestSuite) Test_getToken_UnableToMarshallResponseBody_ReturnsError() {
	sender, err := NewSender()

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	sender.credentialGetter = mockCredentialGetter

	testKey, err := rsa.GenerateKey(rand.Reader, 2048)
	mockCredentialGetter.On("GetPrivateKey", reportStreamPrivateKeyName).Return(testKey, nil)

	// Set up a test server for ReportStream
	// Response parts: Body, Status Code, Access Token (part of body), Error (part of body)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
    "expires_in": 300,
    "expires_at_seconds": 1719527278,
    "scope": "flexion.*.report"
`))
	}))
	defer server.Close()

	sender.baseUrl = server.URL
	token, err := sender.getToken()

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "", token)
}

func (suite *SenderTestSuite) Test_SendMessage_MessageSentToReportStream_ReturnsReportId() {
	sender, err := NewSender()

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	sender.credentialGetter = mockCredentialGetter

	testKey, err := rsa.GenerateKey(rand.Reader, 2048)
	mockCredentialGetter.On("GetPrivateKey", reportStreamPrivateKeyName).Return(testKey, nil)

	// Set up a test server for ReportStream
	// Response parts: Body, Status Code, Access Token (part of body), Error (part of body)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if r.URL.Path == "/api/token" {
			w.Write([]byte(`{
				"sub": "flexion.*.report_e6b68103-dd38-420e-8118-2b2f6c9fa3c4",
				"access_token": "eyJhbGciOiJIUzM4NCJ9.eyJleHAiOjE3MTk1MjcyNzgsInNjb3BlIjoiZmxleGlvbi4qLnJlcG9ydCIsInN1YiI6ImZsZXhpb24uKi5yZXBvcnRfZTZiNjgxMDMtZGQzOC00MjBlLTgxMTgtMmIyZjZjOWZhM2M0In0.liHv9SJYxztgMmCPKGIF2lzcMMMzFAoatLlIC33uz5jbA5wSJa8iIa5yzJ1ZaECI",
				"token_type": "bearer",
				"expires_in": 300,
				"expires_at_seconds": 1719527278,
				"scope": "flexion.*.report"
			}
			`))
		} else {
			w.Write([]byte(`
			{
			  "id" : "78809588-1193-4861-a6a7-52493f7dd254",
			  "submissionId" : 26,
			  "overallStatus" : "Received",
			  "timestamp" : "2024-05-20T21:11:36.144Z",
			  "plannedCompletionAt" : null,
			  "actualCompletionAt" : null,
			  "sender" : "flexion.simulated-hospital",
			  "reportItemCount" : 1,
			  "errorCount" : 0,
			  "warningCount" : 0,
			  "httpStatus" : 201,
			  "destinations" : [ ],
			  "actionName" : "receive",
			  "externalName" : null,
			  "reportId" : "78809588-1193-4861-a6a7-52493f7dd254",
			  "topic" : "etor-ti",
			  "bodyFormat" : "",
			  "errors" : [ ],
			  "warnings" : [ ],
			  "destinationCount" : 0,
			  "fileName" : ""
			}
			`))
		}

	}))
	defer server.Close()

	sender.baseUrl = server.URL
	message, _ := os.ReadFile(filepath.Join("..", "..", "mock_data", "order_message.hl7"))

	reportId, err := sender.SendMessage(message)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "78809588-1193-4861-a6a7-52493f7dd254", reportId)

}

func (suite *SenderTestSuite) Test_SendMessage_UnableToGetToken_ReturnsError() {
	sender, err := NewSender()

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	sender.credentialGetter = mockCredentialGetter

	testKey, err := rsa.GenerateKey(rand.Reader, 2048)
	mockCredentialGetter.On("GetPrivateKey", reportStreamPrivateKeyName).Return(testKey, errors.New(utils.ErrorKey))

	message, _ := os.ReadFile(filepath.Join("..", "..", "mock_data", "order_message.hl7"))

	reportId, err := sender.SendMessage(message)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "error", err.Error())
	assert.Equal(suite.T(), "", reportId)
}

func (suite *SenderTestSuite) Test_SendMessage_UnableToCallTokenEndpoint_ReturnsError() {
	sender, err := NewSender()

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	sender.credentialGetter = mockCredentialGetter

	testKey, err := rsa.GenerateKey(rand.Reader, 2048)
	mockCredentialGetter.On("GetPrivateKey", reportStreamPrivateKeyName).Return(testKey, nil)

	// Set up a test server for ReportStream
	// Response parts: Body, Status Code, Access Token (part of body), Error (part of body)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"sub": "flexion.*.report_e6b68103-dd38-420e-8118-2b2f6c9fa3c4",
				"access_token": "eyJhbGciOiJIUzM4NCJ9.eyJleHAiOjE3MTk1MjcyNzgsInNjb3BlIjoiZmxleGlvbi4qLnJlcG9ydCIsInN1YiI6ImZsZXhpb24uKi5yZXBvcnRfZTZiNjgxMDMtZGQzOC00MjBlLTgxMTgtMmIyZjZjOWZhM2M0In0.liHv9SJYxztgMmCPKGIF2lzcMMMzFAoatLlIC33uz5jbA5wSJa8iIa5yzJ1ZaECI",
				"token_type": "bearer",
				"expires_in": 300,
				"expires_at_seconds": 1719527278,
				"scope": "flexion.*.report"
			}
			`))
		} else {
			panic("Error calling token endpoint")
		}
	}))
	defer server.Close()

	sender.baseUrl = server.URL
	message, _ := os.ReadFile(filepath.Join("..", "..", "mock_data", "order_message.hl7"))

	reportId, err := sender.SendMessage(message)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "", reportId)
	assert.Contains(suite.T(), err.Error(), "EOF")
}

func (suite *SenderTestSuite) Test_SendMessage_StatusCodeIsAbove300_ReturnsError() {
	sender, err := NewSender()

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	sender.credentialGetter = mockCredentialGetter

	testKey, err := rsa.GenerateKey(rand.Reader, 2048)
	mockCredentialGetter.On("GetPrivateKey", reportStreamPrivateKeyName).Return(testKey, nil)

	// Set up a test server for ReportStream
	// Response parts: Body, Status Code, Access Token (part of body), Error (part of body)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Path == "/api/token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"sub": "flexion.*.report_e6b68103-dd38-420e-8118-2b2f6c9fa3c4",
				"access_token": "eyJhbGciOiJIUzM4NCJ9.eyJleHAiOjE3MTk1MjcyNzgsInNjb3BlIjoiZmxleGlvbi4qLnJlcG9ydCIsInN1YiI6ImZsZXhpb24uKi5yZXBvcnRfZTZiNjgxMDMtZGQzOC00MjBlLTgxMTgtMmIyZjZjOWZhM2M0In0.liHv9SJYxztgMmCPKGIF2lzcMMMzFAoatLlIC33uz5jbA5wSJa8iIa5yzJ1ZaECI",
				"token_type": "bearer",
				"expires_in": 300,
				"expires_at_seconds": 1719527278,
				"scope": "flexion.*.report"
			}
			`))
		} else {
			w.WriteHeader(http.StatusFound)
			w.Write([]byte(`
			{
			  "id" : "78809588-1193-4861-a6a7-52493f7dd254",
			  "submissionId" : 26,
			  "overallStatus" : "Received",
			  "timestamp" : "2024-05-20T21:11:36.144Z",
			  "plannedCompletionAt" : null,
			  "actualCompletionAt" : null,
			  "sender" : "flexion.simulated-hospital",
			  "reportItemCount" : 1,
			  "errorCount" : 0,
			  "warningCount" : 0,
			  "httpStatus" : 201,
			  "destinations" : [ ],
			  "actionName" : "receive",
			  "externalName" : null,
			  "reportId" : "78809588-1193-4861-a6a7-52493f7dd254",
			  "topic" : "etor-ti",
			  "bodyFormat" : "",
			  "errors" : [ ],
			  "warnings" : [ ],
			  "destinationCount" : 0,
			  "fileName" : ""
			}
			`))
		}

	}))
	defer server.Close()

	sender.baseUrl = server.URL
	message, _ := os.ReadFile(filepath.Join("..", "..", "mock_data", "order_message.hl7"))

	reportId, err := sender.SendMessage(message)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "302 Found", err.Error())
	assert.Equal(suite.T(), "", reportId)
}

func (suite *SenderTestSuite) Test_SendMessage_StatusCodeIs400_ReturnsNonTransientError() {
	sender, err := NewSender()

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	sender.credentialGetter = mockCredentialGetter

	testKey, err := rsa.GenerateKey(rand.Reader, 2048)
	mockCredentialGetter.On("GetPrivateKey", reportStreamPrivateKeyName).Return(testKey, nil)

	// Set up a test server for ReportStream
	// Response parts: Body, Status Code, Access Token (part of body), Error (part of body)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Path == "/api/token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"sub": "flexion.*.report_e6b68103-dd38-420e-8118-2b2f6c9fa3c4",
				"access_token": "eyJhbGciOiJIUzM4NCJ9.eyJleHAiOjE3MTk1MjcyNzgsInNjb3BlIjoiZmxleGlvbi4qLnJlcG9ydCIsInN1YiI6ImZsZXhpb24uKi5yZXBvcnRfZTZiNjgxMDMtZGQzOC00MjBlLTgxMTgtMmIyZjZjOWZhM2M0In0.liHv9SJYxztgMmCPKGIF2lzcMMMzFAoatLlIC33uz5jbA5wSJa8iIa5yzJ1ZaECI",
				"token_type": "bearer",
				"expires_in": 300,
				"expires_at_seconds": 1719527278,
				"scope": "flexion.*.report"
			}
			`))
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`
			{
			  "id" : "78809588-1193-4861-a6a7-52493f7dd254",
			  "submissionId" : 26,
			  "overallStatus" : "Received",
			  "timestamp" : "2024-05-20T21:11:36.144Z",
			  "plannedCompletionAt" : null,
			  "actualCompletionAt" : null,
			  "sender" : "flexion.simulated-hospital",
			  "reportItemCount" : 1,
			  "errorCount" : 0,
			  "warningCount" : 0,
			  "httpStatus" : 201,
			  "destinations" : [ ],
			  "actionName" : "receive",
			  "externalName" : null,
			  "reportId" : "78809588-1193-4861-a6a7-52493f7dd254",
			  "topic" : "etor-ti",
			  "bodyFormat" : "",
			  "errors" : [ ],
			  "warnings" : [ ],
			  "destinationCount" : 0,
			  "fileName" : ""
			}
			`))
		}

	}))
	defer server.Close()

	sender.baseUrl = server.URL
	message, _ := os.ReadFile(filepath.Join("..", "..", "mock_data", "order_message.hl7"))

	reportId, err := sender.SendMessage(message)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), utils.ReportStreamNonTransientFailure, err.Error())
	assert.Equal(suite.T(), "", reportId)
}

func (suite *SenderTestSuite) Test_SendMessage_StatusCodeIsAbove499_ReturnsError() {
	sender, err := NewSender()
	assert.NoError(suite.T(), err)

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	sender.credentialGetter = mockCredentialGetter

	testKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(suite.T(), err)

	mockCredentialGetter.On("GetPrivateKey", reportStreamPrivateKeyName).Return(testKey, nil)

	// Set up a test server for ReportStream
	// Response parts: Body, Status Code, Access Token (part of body), Error (part of body)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Path == "/api/token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"sub": "flexion.*.report_e6b68103-dd38-420e-8118-2b2f6c9fa3c4",
				"access_token": "eyJhbGciOiJIUzM4NCJ9.eyJleHAiOjE3MTk1MjcyNzgsInNjb3BlIjoiZmxleGlvbi4qLnJlcG9ydCIsInN1YiI6ImZsZXhpb24uKi5yZXBvcnRfZTZiNjgxMDMtZGQzOC00MjBlLTgxMTgtMmIyZjZjOWZhM2M0In0.liHv9SJYxztgMmCPKGIF2lzcMMMzFAoatLlIC33uz5jbA5wSJa8iIa5yzJ1ZaECI",
				"token_type": "bearer",
				"expires_in": 300,
				"expires_at_seconds": 1719527278,
				"scope": "flexion.*.report"
			}
			`))
		} else {
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte(`
			{
			  "id" : "78809588-1193-4861-a6a7-52493f7dd254",
			  "submissionId" : 26,
			  "overallStatus" : "Received",
			  "timestamp" : "2024-05-20T21:11:36.144Z",
			  "plannedCompletionAt" : null,
			  "actualCompletionAt" : null,
			  "sender" : "flexion.simulated-hospital",
			  "reportItemCount" : 1,
			  "errorCount" : 0,
			  "warningCount" : 0,
			  "httpStatus" : 201,
			  "destinations" : [ ],
			  "actionName" : "receive",
			  "externalName" : null,
			  "reportId" : "78809588-1193-4861-a6a7-52493f7dd254",
			  "topic" : "etor-ti",
			  "bodyFormat" : "",
			  "errors" : [ ],
			  "warnings" : [ ],
			  "destinationCount" : 0,
			  "fileName" : ""
			}
			`))
		}

	}))
	defer server.Close()

	sender.baseUrl = server.URL
	message, _ := os.ReadFile(filepath.Join("..", "..", "mock_data", "order_message.hl7"))

	reportId, err := sender.SendMessage(message)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "502 Bad Gateway", err.Error())
	assert.Equal(suite.T(), "", reportId)
}

func (suite *SenderTestSuite) Test_SendMessage_UnableToParseResponseBody_ReturnsError() {
	sender, err := NewSender()

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	sender.credentialGetter = mockCredentialGetter

	testKey, err := rsa.GenerateKey(rand.Reader, 2048)
	mockCredentialGetter.On("GetPrivateKey", reportStreamPrivateKeyName).Return(testKey, nil)

	// Set up a test server for ReportStream
	// Response parts: Body, Status Code, Access Token (part of body), Error (part of body)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Path == "/api/token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"sub": "flexion.*.report_e6b68103-dd38-420e-8118-2b2f6c9fa3c4",
				"access_token": "eyJhbGciOiJIUzM4NCJ9.eyJleHAiOjE3MTk1MjcyNzgsInNjb3BlIjoiZmxleGlvbi4qLnJlcG9ydCIsInN1YiI6ImZsZXhpb24uKi5yZXBvcnRfZTZiNjgxMDMtZGQzOC00MjBlLTgxMTgtMmIyZjZjOWZhM2M0In0.liHv9SJYxztgMmCPKGIF2lzcMMMzFAoatLlIC33uz5jbA5wSJa8iIa5yzJ1ZaECI",
				"token_type": "bearer",
				"expires_in": 300,
				"expires_at_seconds": 1719527278,
				"scope": "flexion.*.report"
			}
			`))
		} else {
			w.Write([]byte(`Expected a 'client' query parameter`))
		}

	}))
	defer server.Close()

	sender.baseUrl = server.URL
	message, _ := os.ReadFile(filepath.Join("..", "..", "mock_data", "order_message.hl7"))

	reportId, err := sender.SendMessage(message)

	assert.Equal(suite.T(), "invalid character 'E' looking for beginning of value", err.Error())
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "", reportId)
}

func Test_SenderTestSuite(t *testing.T) {
	suite.Run(t, new(SenderTestSuite))
}
