package report_stream

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

/*
Eventually:
- Configure client to call RS? Load config?
- Generate JWT to call token endpoint with
- Call token endpoint (POST {{protocol}}://{{rs-host}}:{{rs-port}}/api/token)
- Call 'Send HL7v2 Message' endpoint (POST {{protocol}}://{{rs-host}}:{{rs-port}}/api/waters)

For now:
- Call /api/reports (needs API key in deployed env and no security locally)
*/

type Report struct {
	ReportId string `json:"reportId"`
}

type Sender struct {
	BaseUrl string
}

func (apiHandler Sender) SendMessage(message []byte) (string, error) {

	client := http.Client{}
	req, err := http.NewRequest("POST", apiHandler.BaseUrl+"/api/reports", bytes.NewBuffer(message))

	if err != nil {
		return "", err
	}

	req.Header = http.Header{
		"content-type": {"application/hl7-v2"},
		"client":       {"flexion.simulated-hospital"},
		//"x-functions-key":    {"ABC_REPLACE_ME"},
		//"Authorization: Bearer ": {""},
	}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	responseBodyBytes, err := io.ReadAll(res.Body)
	slog.Info("response body", slog.String("responseBodyBytes", string(responseBodyBytes)))

	if err != nil {
		return "", err
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
Sample response from RS:
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

*/
