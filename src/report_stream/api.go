package report_stream

import (
	"bytes"
	"fmt"
	"io"
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

type ApiHandler struct {
	BaseUrl string
}

//func (apiHandler *ApiHandler) Login {}

func (apiHandler *ApiHandler) SendReport(hl7message []byte) error {

	client := http.Client{}
	req, err := http.NewRequest("POST", apiHandler.BaseUrl+"/api/reports", bytes.NewBuffer(hl7message))

	if err != nil {
		return err
	}

	req.Header = http.Header{
		"content-type": {"application/hl7-v2"},
		"client":       {"flexion.simulated-hospital"},
		//"x-functions-key":    {"ABC_REPLACE_ME"},
		//"Authorization: Bearer ": {""},
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	responseBodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	fmt.Println(string(responseBodyBytes))

	return nil
}
