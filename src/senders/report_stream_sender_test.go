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

type SenderTestSuite struct {
	suite.Suite
}

func (suite *SenderTestSuite) SetupTest() {
	os.Setenv("ENV", "local")
	os.Setenv("REPORT_STREAM_URL_PREFIX", "rs.com")
	os.Setenv("FLEXION_PRIVATE_KEY_NAME", "key")
	os.Setenv("FLEXION_CLIENT_NAME", "client")
}

func (suite *SenderTestSuite) TearDownTest() {
	os.Unsetenv("ENV")
	os.Unsetenv("REPORT_STREAM_URL_PREFIX")
	os.Unsetenv("FLEXION_PRIVATE_KEY_NAME")
	os.Unsetenv("FLEXION_CLIENT_NAME")
}

func (suite *SenderTestSuite) Test_Sender_NewSender_InitsWithValues() {

	sender, err := NewSender()

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), os.Getenv("REPORT_STREAM_URL_PREFIX"), sender.baseUrl)
	assert.Equal(suite.T(), os.Getenv("FLEXION_PRIVATE_KEY_NAME"), sender.privateKeyName)
	assert.Equal(suite.T(), os.Getenv("FLEXION_CLIENT_NAME"), sender.clientName)
}

func (suite *SenderTestSuite) Test_Sender_NewSender_ReturnLocalCredentials() {
	os.Setenv("ENV", "")
	sender, err := NewSender()

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), os.Getenv("REPORT_STREAM_URL_PREFIX"), sender.baseUrl)
	assert.Equal(suite.T(), os.Getenv("FLEXION_PRIVATE_KEY_NAME"), sender.privateKeyName)
	assert.Equal(suite.T(), os.Getenv("FLEXION_CLIENT_NAME"), sender.clientName)
}

func (suite *SenderTestSuite) Test_Sender_GenerateJWT_ReturnsJWT() {

	sender, err := NewSender()
	assert.NoError(suite.T(), err)

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	sender.credentialGetter = mockCredentialGetter

	testKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(suite.T(), err)

	mockCredentialGetter.On("GetPrivateKey", "key").Return(testKey, nil)
	jwt, err := sender.generateJwt()

	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), jwt)
}

func (suite *SenderTestSuite) Test_Sender_GenerateJWT_ReturnsError_WhenGetPrivateKeyReturnsError() {

	sender, err := NewSender()
	assert.NoError(suite.T(), err)

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	sender.credentialGetter = mockCredentialGetter

	mockCredentialGetter.On("GetPrivateKey", "key").Return(&rsa.PrivateKey{}, errors.New("failed to retrieve private key"))
	_, err = sender.generateJwt()

	assert.Error(suite.T(), err)
}

func (suite *SenderTestSuite) Test_Sender_getToken_ReturnsAccessToken() {
	sender, err := NewSender()
	assert.NoError(suite.T(), err)

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	sender.credentialGetter = mockCredentialGetter

	testKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(suite.T(), err)

	mockCredentialGetter.On("GetPrivateKey", "key").Return(testKey, nil)

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

func (suite *SenderTestSuite) Test_Sender_getToken_ReturnErrorWhenUnableToGenerateJWT() {
	sender, err := NewSender()
	assert.NoError(suite.T(), err)

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	sender.credentialGetter = mockCredentialGetter

	mockCredentialGetter.On("GetPrivateKey", "key").Return(&rsa.PrivateKey{}, errors.New("failed to retrieve private key"))
	token, err := sender.getToken()

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "", token)
}

func (suite *SenderTestSuite) Test_Sender_getToken_ReturnErrorWhenUnableToCallTokenEndpoint() {
	os.Setenv("REPORT_STREAM_URL_PREFIX", "this is not a good URL")
	sender, err := NewSender()
	assert.NoError(suite.T(), err)

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	sender.credentialGetter = mockCredentialGetter

	testKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(suite.T(), err)

	mockCredentialGetter.On("GetPrivateKey", "key").Return(testKey, nil)

	token, err := sender.getToken()

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "", token)
}

func (suite *SenderTestSuite) Test_Sender_getToken_ReturnErrorWhenResponseStatusInvalid() {
	sender, err := NewSender()
	assert.NoError(suite.T(), err)

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	sender.credentialGetter = mockCredentialGetter

	testKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(suite.T(), err)

	mockCredentialGetter.On("GetPrivateKey", "key").Return(testKey, nil)

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

func (suite *SenderTestSuite) Test_Sender_getToken_ReturnsErrorWhenUnableToMarshallBody() {
	sender, err := NewSender()
	assert.NoError(suite.T(), err)

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	sender.credentialGetter = mockCredentialGetter

	testKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(suite.T(), err)

	mockCredentialGetter.On("GetPrivateKey", "key").Return(testKey, nil)

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

func (suite *SenderTestSuite) Test_Sender_SendMessage_sendMessage() {
	sender, err := NewSender()
	assert.NoError(suite.T(), err)

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	sender.credentialGetter = mockCredentialGetter

	testKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(suite.T(), err)

	mockCredentialGetter.On("GetPrivateKey", "key").Return(testKey, nil)

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

func (suite *SenderTestSuite) Test_Sender_SendMessage_returnErrorWhenUnableToGetToken() {
	sender, err := NewSender()
	assert.NoError(suite.T(), err)

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	sender.credentialGetter = mockCredentialGetter

	testKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(suite.T(), err)

	mockCredentialGetter.On("GetPrivateKey", "key").Return(testKey, errors.New(utils.ErrorKey))

	message, _ := os.ReadFile(filepath.Join("..", "..", "mock_data", "order_message.hl7"))

	reportId, err := sender.SendMessage(message)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "error", err.Error())
	assert.Equal(suite.T(), "", reportId)
}

func (suite *SenderTestSuite) Test_Sender_SendMessage_returnErrorWhenUnableToCallTokenEndpoint() {
	sender, err := NewSender()
	assert.NoError(suite.T(), err)

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	sender.credentialGetter = mockCredentialGetter

	testKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(suite.T(), err)

	mockCredentialGetter.On("GetPrivateKey", "key").Return(testKey, nil)

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

func (suite *SenderTestSuite) Test_Sender_SendMessage_returnErrorWhenStatusCodeIsAbove300() {
	sender, err := NewSender()
	assert.NoError(suite.T(), err)

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	sender.credentialGetter = mockCredentialGetter

	testKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(suite.T(), err)

	mockCredentialGetter.On("GetPrivateKey", "key").Return(testKey, nil)

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
			w.WriteHeader(http.StatusNotFound)
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

	assert.Equal(suite.T(), "404 Not Found", err.Error())
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "", reportId)
}

func (suite *SenderTestSuite) Test_Sender_SendMessage_returnErrorWhenUnableToParseResponseBody() {
	sender, err := NewSender()
	assert.NoError(suite.T(), err)

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	sender.credentialGetter = mockCredentialGetter

	testKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(suite.T(), err)

	mockCredentialGetter.On("GetPrivateKey", "key").Return(testKey, nil)

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
