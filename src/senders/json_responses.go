package senders

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
  "senders" : "flexion.simulated-hospital",
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
    "senders": "",
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
