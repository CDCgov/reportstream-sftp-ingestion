package report_stream

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/CDCgov/reportstream-sftp-ingestion/azure"
	"github.com/CDCgov/reportstream-sftp-ingestion/local"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

/*
TODO:
- Set up environment variables for FLEXION_PRIVATE_KEY_NAME and FLEXION_CLIENT_NAME
- TF for those vars - QUESTION: should we be loading private key from TF like as data, or creating it new in TF and then manually updating it?
- Manually update key in TF?
- Get key from ENV
- Generate JWT to call token endpoint with
- Call token endpoint (POST {{protocol}}://{{rs-host}}:{{rs-port}}/api/token) and do something with the response - cache? Use right away?
- Update the send message code to call 'Send HL7v2 Message' endpoint (POST {{protocol}}://{{rs-host}}:{{rs-port}}/api/waters)

For now:
- Call /api/reports (needs API key in deployed env and no security locally)
*/

//TODO - move RS structs to a 'types' file?

type Report struct {
	ReportId string `json:"reportId"`
}

type ReportStreamToken struct {
	Sub              string `json:"sub"`
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int    `json:"expires_in"`
	ExpiresAtSeconds int    `json:"expires_at_seconds"`
	Scope            string `json:"scope"`
}
type ReportStreamError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ErrorUri         string `json:"error_uri"`
}

type Sender struct {
	BaseUrl          string
	PrivateKeyName   string
	ClientName       string
	credentialGetter CredentialGetter
}

func NewSender() Sender {
	// TODO get sender info based on source folder
	var credentialGetter CredentialGetter

	environment := os.Getenv("ENV")
	if environment == "" {
		environment = "local"
	}
	if environment == "local" {
		slog.Info("Using local credentials")
		credentialGetter = local.CredentialGetter{}
	} else {
		slog.Info("Using Azure credentials")
		credentialGetter = azure.CredentialGetter{}
	}

	return Sender{
		BaseUrl:          os.Getenv("REPORT_STREAM_URL_PREFIX"),
		PrivateKeyName:   os.Getenv("FLEXION_PRIVATE_KEY_NAME"),
		ClientName:       os.Getenv("FLEXION_CLIENT_NAME"),
		credentialGetter: credentialGetter,
	}
}

//TODO - cache key and/or JWT and/or token somewhere rather than requesting each time? JWT and token both expire
//TODO - unchain the key/JWT/token/submit sequence?

func (sender Sender) GenerateJwt() (string, error) {

	key, err := sender.credentialGetter.GetPrivateKey(sender.PrivateKeyName)

	if err != nil {
		return "", err
	}
	id, _ := uuid.NewUUID()
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		ID:        id.String(),
		Issuer:    sender.ClientName,
		Subject:   sender.ClientName,
		Audience:  jwt.ClaimStrings{os.Getenv("ENV") + ".prime.cdc.gov"},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	t.Header["kid"] = sender.ClientName

	return t.SignedString(key)
}

func (sender Sender) GetToken() (string, error) {
	senderJwt, err := sender.GenerateJwt()
	if err != nil {
		return "", err
	}

	// TODO
	// - put the scope into config
	// - figure out if org and sender should be split apart (e.g. `flexion.simulated-hospital` vs `flexion` and `simulated-hospital`)
	data := url.Values{
		"scope":                 {"flexion.*.report"},
		"grant_type":            {"client_credentials"},
		"client_assertion_type": {"urn:ietf:params:oauth:client-assertion-type:jwt-bearer"},
		"client_assertion":      {senderJwt},
	}

	req, err := http.NewRequest("POST", sender.BaseUrl+"/api/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	req.Header = http.Header{
		"Content-Type": {"application/x-www-form-urlencoded"},
	}

	client := http.Client{}
	res, err := client.Do(req)

	if err != nil {
		slog.Info("error calling token endpoint", slog.Any("err", err))
		return "", err
	}

	defer res.Body.Close()

	responseBodyBytes, err := io.ReadAll(res.Body)

	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		slog.Info("response body", slog.String("responseBodyBytes", string(responseBodyBytes)))
		// TODO - decide when/whether to parse body - RS error responses are sometimes strings and sometimes JSON
		return "", errors.New(res.Status)
	}
	var token ReportStreamToken
	err = json.Unmarshal(responseBodyBytes, &token)
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}

func (sender Sender) SendMessage(message []byte) (string, error) {
	token, err := sender.GetToken()
	if err != nil {
		return "", err
	}

	client := http.Client{}
	req, err := http.NewRequest("POST", sender.BaseUrl+"/api/waters", bytes.NewBuffer(message))

	if err != nil {
		return "", err
	}

	req.Header = http.Header{
		"content-type":  {"application/hl7-v2"},
		"client":        {sender.ClientName},
		"Authorization": {"Bearer " + token},
	}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	responseBodyBytes, err := io.ReadAll(res.Body)

	if err != nil {
		return "", err
	}

	if res.StatusCode >= 300 {
		slog.Info("status", slog.Any("code", res.StatusCode), slog.String("status", res.Status))
		slog.Info("response body", slog.String("responseBodyBytes", string(responseBodyBytes)))
		// TODO - decide when/whether to parse body - RS error responses are sometimes strings and sometimes JSON
		return "", errors.New(res.Status)
	}

	var report Report
	err = json.Unmarshal(responseBodyBytes, &report)
	if err != nil {
		return "", err
	}

	slog.Info("report", slog.Any("report", report))
	return report.ReportId, nil
}

/**
Sample responses from RS waters
Success:
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

JSON Error:
{
    "id": null,
    "submissionId": 91,
    "overallStatus": "Error",
    "timestamp": "2024-05-23T21:36:46.879Z",
    "plannedCompletionAt": null,
    "actualCompletionAt": null,
    "sender": "",
    "reportItemCount": null,
    "errorCount": 1,
    "warningCount": 0,
    "httpStatus": 400,
    "destinations": [],
    "actionName": "receive",
    "externalName": "",
    "reportId": null,
    "topic": null,
    "bodyFormat": "",
    "errors": [
        {
            "scope": "parameter",
            "message": "Blank message(s) found within file. Blank messages cannot be processed.",
            "errorCode": "UNKNOWN"
        }
    ],
    "warnings": [],
    "destinationCount": 0,
    "fileName": ""
}

String error:
Expected a 'client' query parameter

*/
