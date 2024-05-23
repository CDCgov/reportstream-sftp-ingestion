package report_stream

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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
	BaseUrl        string
	PrivateKeyName string
	ClientName     string
}

func NewSender() Sender {
	// TODO get sender info based on source folder
	return Sender{BaseUrl: os.Getenv("REPORT_STREAM_URL_PREFIX"), PrivateKeyName: os.Getenv("FLEXION_PRIVATE_KEY_NAME"), ClientName: os.Getenv("FLEXION_CLIENT_NAME")}
}

//TODO - cache key and/or JWT and/or token somewhere rather than requesting each time? JWT and token both expire
//TODO - unchain the key/JWT/token/submit sequence?

func (sender Sender) GetPrivateKey() (*rsa.PrivateKey, error) {
	pem, err := os.ReadFile(filepath.Join("mock_credentials", fmt.Sprintf("%s.pem", sender.PrivateKeyName)))
	if err != nil {
		return nil, err
	}
	key, err := jwt.ParseRSAPrivateKeyFromPEM(pem)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (sender Sender) GenerateJwt() (string, error) {
	key, err := sender.GetPrivateKey()
	if err != nil {
		return "", err
	}
	id, _ := uuid.NewUUID()
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		ID:        id.String(),
		// missing -k, --kid <KID> the kid to place in the header - set to flexion.simulated-hospital in command line
		Issuer:  sender.ClientName,
		Subject: sender.ClientName,
		//Audience: "staging.prime.cdc.gov", // not sure why it doesn't like this
		// I think we don't need the --no-iat - looks like it should be excluded automatically
	}

	t := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
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
	reqBody, _ := io.ReadAll(req.Body)
	slog.Info("request", slog.Any("body", reqBody))

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		slog.Info("error calling token endpoint", slog.Any("err", err))
		return "", err
	}

	defer res.Body.Close()

	responseBodyBytes, err := io.ReadAll(res.Body)
	slog.Info("response", slog.Any("res", res))
	slog.Info("responseBodyBytes", slog.Any("responseBodyBytes", responseBodyBytes))

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
		"content-type":           {"application/hl7-v2"},
		"client":                 {sender.ClientName},
		"Authorization: Bearer ": {token},
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

	if res.StatusCode != http.StatusOK {
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
