// Copyright 2016, RadiantBlue Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pzsvc

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

// These can go away when we move to Go 1.6
const (
	MethodOptions = "OPTIONS"
	MethodGet     = "GET"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethdoDelete  = "DELETE"
)

var (
	domain = os.Getenv("DOMAIN")
)

// Gateway returns the URL of the Piazza Gateway
func Gateway() string {
	return "https://pz-gateway." + domain
}

// HTTPError represents any HTTP error
type HTTPError struct {
	Status  int
	Message string
}

func (err HTTPError) Error() string {
	return fmt.Sprintf("%d: %v", err.Status, err.Message)
}

var httpClient *http.Client

// HTTPClient is a factory method for a http.Client suitable for common operations
func HTTPClient() *http.Client {
	if httpClient == nil {
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		httpClient = &http.Client{Transport: transport}
	}
	return httpClient
}

// SetHTTPClient is used to set the current http client.  This is mostly useful
// for testing purposes
func SetHTTPClient(newClient *http.Client) {
	httpClient = newClient
}

// RequestKnownJSON submits an http request where the response is assumed to be JSON
// for which the format is known.  Given an object of the appropriate format for
// said response JSON, an address to call and an authKey to send, it will submit
// the get request, unmarshal the result into the given object, and return. It
// returns the response buffer, in case it is needed for debugging purposes.
func RequestKnownJSON(method, bodyStr, address, authKey string, outpObj interface{}) ([]byte, error) {

	resp, err := SubmitSinglePart(method, bodyStr, address, authKey)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		errByt, _ := ioutil.ReadAll(resp.Body)
		return errByt, addRef(err)
	}
	return ReadBodyJSON(&outpObj, resp.Body)
}

// SubmitMultipart sends a multi-part POST call, including an optional uploaded file,
// and returns the response.  Primarily intended to support Ingest calls.
func SubmitMultipart(bodyStr, address, filename, authKey string, fileData []byte) (*http.Response, error) {

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	client := HTTPClient()

	err := writer.WriteField("data", bodyStr)
	if err != nil {
		return nil, addRef(err)
	}

	if fileData != nil {
		var part io.Writer
		part, err = writer.CreateFormFile("file", filename)
		if err != nil {
			return nil, addRef(err)
		}
		if part == nil {
			return nil, errWithRef("Failure in Form File Creation.")
		}

		_, err = io.Copy(part, bytes.NewReader(fileData))
		if err != nil {
			return nil, addRef(err)
		}
	}

	err = writer.Close()
	if err != nil {
		return nil, addRef(err)
	}

	fileReq, err := http.NewRequest("POST", address, body)
	if err != nil {
		return nil, addRef(err)
	}

	fileReq.Header.Add("Content-Type", writer.FormDataContentType())
	fileReq.Header.Add("Authorization", authKey)

	resp, err := client.Do(fileReq)
	if err != nil {
		return nil, addRef(err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return resp, errWithRef("Failed to POST multipart to " + address + " Status: " + resp.Status)
	}
	return resp, addRef(err)
}

// SubmitSinglePart sends a single-part GET/POST/PUT/DELETE call to the target URL
// and returns the result.  Includes the necessary headers.
func SubmitSinglePart(method, bodyStr, url, authKey string) (*http.Response, error) {

	var fileReq *http.Request
	var err error
	client := HTTPClient()

	if bodyStr != "" {
		fileReq, err = http.NewRequest(method, url, bytes.NewBuffer([]byte(bodyStr)))
		if err != nil {
			return nil, addRef(err)
		}
		fileReq.Header.Add("Content-Type", "application/json")
	} else {
		fileReq, err = http.NewRequest(method, url, nil)
		if err != nil {
			return nil, addRef(err)
		}
	}

	fileReq.Header.Add("Authorization", authKey)

	resp, err := client.Do(fileReq)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return resp, errWithRef("Failed in " + method + " call to " + url + ".  Status : " + resp.Status)
	}

	return resp, addRef(err)
}

// GetJobResponse will repeatedly poll the job status on the given job Id
// until job completion, then acquires and returns the DataResult.
func GetJobResponse(jobID, pzAddr, authKey string) (*DataResult, error) {

	if jobID == "" {
		return nil, fmt.Errorf(`JobID not provided.  Cannot acquire DataResult.`)
	}

	for i := 0; i < 180; i++ { // will wait up to 3 minutes

		var outpObj struct {
			Data JobStatusResp `json:"data,omitempty"`
		}
		respBuf, err := RequestKnownJSON("GET", "", pzAddr+"/job/"+jobID, authKey, &outpObj)
		if err != nil {
			return nil, addRef(err)
		}

		respObj := &outpObj.Data
		if respObj.Status == "Submitted" ||
			respObj.Status == "Running" ||
			respObj.Status == "Pending" ||
			(respObj.Status == "Success" && respObj.Result == nil) ||
			(respObj.Status == "Error" && respObj.Result.Message == "Job Not Found.") {
			time.Sleep(time.Second)
		} else {
			if respObj.Status == "Success" {
				return respObj.Result, nil
			}
			if respObj.Status == "Fail" {
				return nil, errWithRef("Piazza failure when acquiring DataId.  Response json: " + string(respBuf))
			}
			if respObj.Status == "Error" {
				return nil, errWithRef("Piazza error when acquiring DataId.  Response json: " + string(respBuf))
			}
			return nil, errWithRef(`Unknown status "` + respObj.Status + `" when acquiring DataId.  Response json: ` + string(respBuf))
		}
	}

	return nil, errWithRef("Never completed.  JobId: " + jobID)
}

// GetJobID is a simple function to extract the job ID from
// the standard response to job-creating Pz calls
func GetJobID(resp *http.Response) (string, error) {
	var respObj JobInitResp
	_, err := ReadBodyJSON(&respObj, resp.Body)
	if respObj.Data.JobID == "" && err == nil {
		err = errWithRef("GetJobID: response did not contain Job ID.")
	}
	return respObj.Data.JobID, addRef(err)
}

// ReadBodyJSON takes the body of either a request object or a response
// object, pulls out the body, and attempts to interpret it as JSON into
// the given interface format.  It's mostly there as a minor simplifying
// function.
func ReadBodyJSON(output interface{}, body io.ReadCloser) ([]byte, error) {
	rBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, addRef(err)
	}

	err = json.Unmarshal(rBytes, output)
	return rBytes, addRef(err)
}

// // GetGateway performs a GET request on the Piazza Gateway
// func GetGateway(endpoint, pzAuth string, target interface{}) error {
// 	var (
// 		result       []byte
// 		err          error
// 		request      *http.Request
// 		response     *http.Response
// 		jsonResponse piazza.JsonResponse
// 	)
//
// 	requestURL := "https://pz-gateway." + domain + endpoint
// 	if request, err = http.NewRequest("GET", requestURL, nil); err != nil {
// 		return &HTTPError{Status: http.StatusInternalServerError, Message: "Unable to create new GET request: " + err.Error()}
// 	}
// 	request.Header.Set("Authorization", pzAuth)
//
// 	if response, err = HTTPClient().Do(request); err != nil {
// 		return &HTTPError{Status: http.StatusInternalServerError, Message: "Unable to do GET request: " + err.Error()}
// 	}
//
// 	defer response.Body.Close()
// 	if result, err = ioutil.ReadAll(response.Body); err != nil {
// 		return &HTTPError{Status: http.StatusInternalServerError, Message: "GET request failed and could not retrieve the message body: " + err.Error()}
// 	}
//
// 	// Check for HTTP errors
// 	if response.StatusCode < 200 || response.StatusCode > 299 {
// 		return &HTTPError{Status: response.StatusCode, Message: "GET request failed:\n" + string(result)}
// 	}
//
// 	if err = json.Unmarshal(result, &jsonResponse); err != nil {
// 		return &HTTPError{Status: http.StatusInternalServerError, Message: "Failed to unmarhsal GET response:\n" + err.Error()}
// 	}
//
// 	if target != nil {
// 		return jsonResponse.ExtractData(target)
// 	}
// 	return nil
// }
//
// // PostGateway performs a GET request on the Piazza Gateway and returns a []byte
// func PostGateway(endpoint string, body []byte, pzAuth string) ([]byte, error) {
// 	var (
// 		result   []byte
// 		err      error
// 		request  *http.Request
// 		response *http.Response
// 	)
// 	requestURL := "https://pz-gateway." + domain + endpoint
// 	if request, err = http.NewRequest("POST", requestURL, bytes.NewBuffer(body)); err != nil {
// 		return nil, &HTTPError{Status: http.StatusInternalServerError, Message: "Unable to create new POST request: " + err.Error()}
// 	}
// 	request.Header.Set("Authorization", pzAuth)
// 	request.Header.Set("Content-Type", "application/json")
//
// 	if response, err = HTTPClient().Do(request); err != nil {
// 		return nil, &HTTPError{Status: http.StatusInternalServerError, Message: "Unable to do POST request: " + err.Error()}
// 	}
//
// 	defer response.Body.Close()
// 	if result, err = ioutil.ReadAll(response.Body); err != nil {
// 		return nil, &HTTPError{Status: http.StatusInternalServerError, Message: "POST request failed and could not retrieve the message body: " + err.Error()}
// 	}
//
// 	// Check for HTTP errors
// 	if response.StatusCode < 200 || response.StatusCode > 299 {
// 		return result, &HTTPError{Status: response.StatusCode, Message: "POST request failed:\n" + string(result)}
// 	}
//
// 	return result, err
// }
