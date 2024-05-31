package report_stream

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/CDCgov/reportstream-sftp-ingestion/azure"
	"github.com/CDCgov/reportstream-sftp-ingestion/local"
	"github.com/CDCgov/reportstream-sftp-ingestion/utils"
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

type Sender struct {
	baseUrl          string
	privateKeyName   string
	clientName       string
	credentialGetter utils.CredentialGetter
}

func NewSender() (Sender, error) {
	environment := os.Getenv("ENV")
	if environment == "" {
		environment = "local"
	}

	var credentialGetter utils.CredentialGetter

	if environment == "local" {
		slog.Info("Using local credentials")
		credentialGetter = local.CredentialGetter{}
	} else {
		slog.Info("Using Azure credentials")
		var err error
		credentialGetter, err = azure.NewSecretGetter()
		if err != nil {
			return Sender{}, err
		}
	}

	return Sender{
		baseUrl:          os.Getenv("REPORT_STREAM_URL_PREFIX"),
		privateKeyName:   os.Getenv("FLEXION_PRIVATE_KEY_NAME"),
		clientName:       os.Getenv("FLEXION_CLIENT_NAME"),
		credentialGetter: credentialGetter,
	}, nil
}

func (sender Sender) generateJwt() (string, error) {

	key, err := sender.credentialGetter.GetPrivateKey(sender.privateKeyName)

	if err != nil {
		return "", err
	}
	id, _ := uuid.NewUUID()
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		ID:        id.String(),
		Issuer:    sender.clientName,
		Subject:   sender.clientName,
		Audience:  jwt.ClaimStrings{os.Getenv("ENV") + ".prime.cdc.gov"},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	t.Header["kid"] = sender.clientName

	return t.SignedString(key)
}

func (sender Sender) getToken() (string, error) {
	senderJwt, err := sender.generateJwt()
	if err != nil {
		return "", err
	}

	data := url.Values{
		"scope":                 {"flexion.*.report"},
		"grant_type":            {"client_credentials"},
		"client_assertion_type": {"urn:ietf:params:oauth:client-assertion-type:jwt-bearer"},
		"client_assertion":      {senderJwt},
	}

	req, err := http.NewRequest("POST", sender.baseUrl+"/api/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	req.Header = http.Header{
		"Content-Type": {"application/x-www-form-urlencoded"},
	}

	res, err := http.DefaultClient.Do(req)

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
	token, err := sender.getToken()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", sender.baseUrl+"/api/waters", bytes.NewBuffer(message))

	if err != nil {
		return "", err
	}

	req.Header = http.Header{
		"content-type":  {"application/hl7-v2"},
		"client":        {sender.clientName},
		"Authorization": {"Bearer " + token},
	}

	res, err := http.DefaultClient.Do(req)
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
