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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"time"
)

// RequestKnownJSON submits an http request where the response is assumed to be JSON
// for which the format is known.  Given an object of the appropriate format for
// said response JSON, an address to call and an authKey to send, it will submit
// the get request, demarshal the result into the given object, and return. It
// returns the response buffer, in case it is needed for debugging purposes.
func RequestKnownJSON(method, bodyStr, address, authKey string, outpObj interface{}, client *http.Client) ([]byte, error) {

	resp, err := SubmitSinglePart(method, bodyStr, address, authKey, client)
	if resp == nil {
		return nil, fmt.Errorf("GetPzObj: no response")
	}
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	return ReadBodyJSON(&outpObj, resp.Body)
}


// SubmitMultipart sends a multi-part POST call, including an optional uploaded file,
// and returns the response.  Primarily intended to support Ingest calls.
func SubmitMultipart(bodyStr, address, filename, authKey string, fileData []byte, client *http.Client) (*http.Response, error) {

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	err := writer.WriteField("data", bodyStr)
	if err != nil {
		return nil, err
	}

	if fileData != nil {
		part, err := writer.CreateFormFile("file", filename)
		if err != nil {
			return nil, err
		}
		if (part == nil) {
			return nil, fmt.Errorf("Failure in Form File Creation.")
		}

		_, err = io.Copy(part, bytes.NewReader(fileData))
		if err != nil {
			return nil, err
		}
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	fileReq, err := http.NewRequest("POST", address, body)
	if err != nil {
		return nil, err
	}

	fileReq.Header.Add("Content-Type", writer.FormDataContentType())
	fileReq.Header.Add("Authorization", authKey)

	resp, err := client.Do(fileReq)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return resp, fmt.Errorf("Failed to POST multipart to " + address + " Status: " + resp.Status)
	}
	return resp, err
}

// SubmitSinglePart sends a single-part GET/POST/PUT/DELETE call to Pz and returns the
// Includes the necessary headers.
func SubmitSinglePart(method, bodyStr, url, authKey string, client *http.Client) (*http.Response, error) {

	var fileReq *http.Request
	var err error

	if bodyStr != "" {
		fileReq, err = http.NewRequest(method, url, bytes.NewBuffer([]byte(bodyStr)))
		if err != nil {
			return nil, err
		}
		fileReq.Header.Add("Content-Type", "application/json")
	} else {
		fileReq, err = http.NewRequest(method, url, nil)
	}
	
	fileReq.Header.Add("Authorization", authKey)

	resp, err := client.Do(fileReq)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return resp, fmt.Errorf("Failed in " + method + " call to " + url + ".  Status : " + resp.Status)
	}

	return resp, err
}

// GetJobResponse will repeatedly poll the job status on the given job Id
// until job completion, then acquires and returns the DataResult.  
func GetJobResponse(jobID, pzAddr, authKey string, client *http.Client) (*DataResult, error) {

	if jobID == "" {
		return nil, fmt.Errorf(`JobID not provided after ingest.  Cannot acquire dataID.`)
	}

	for i := 0; i < 180; i++ { // will wait up to 3 minutes

		var respObj JobStatusResp
		respBuf, err := RequestKnownJSON("GET", "", pzAddr + "/job/" + jobID, authKey, &respObj, client)
		if err != nil {
			return nil, err
		}

		if	respObj.Status == "Submitted" ||
			respObj.Status == "Running" ||
			respObj.Status == "Pending" ||
			( respObj.Status == "Success" && respObj.Result == nil ) ||
			( respObj.Status == "Error" && respObj.Result.Message == "Job Not Found." )  {
			time.Sleep(time.Second)
		} else {
			if respObj.Status == "Success" {
				return respObj.Result, nil
			}
			if respObj.Status == "Fail" {
				return nil, errors.New("Piazza failure when acquiring DataId.  Response json: " + string(respBuf))
			}
			if respObj.Status == "Error" {
				return nil, errors.New("Piazza error when acquiring DataId.  Response json: " + string(respBuf))
			}
			return nil, errors.New("Unknown status when acquiring DataId.  Response json: " + string(respBuf))
		}
	}

	return nil, fmt.Errorf("Never completed.  JobId: %s", jobID)
}

// GetJobID is a simple function to extract the job ID from
// the standard response to job-creating Pz calls
func GetJobID(resp *http.Response) (string, error) {
	var respObj JobInitResp
	_, err := ReadBodyJSON(&respObj, resp.Body)
	if respObj.Data.JobID == "" && err == nil {
		err = errors.New("GetJobID: response did not contain Job ID.")
	}
	return respObj.Data.JobID, err
}

// SliceToCommaSep takes a string slice, and turns it into a comma-separated
// list of strings, suitable for JSON.
func SliceToCommaSep(inSlice []string) string {
	sliLen := len(inSlice)
	if (sliLen == 0){
		return ""
	}
	accum := inSlice[0]
	for i := 1; i < sliLen; i++ {
		accum = accum + "," + inSlice[i]
	}
	return accum
}

// ReadBodyJSON takes the body of either a request object or a response
// object, pulls out the body, and attempts to interpret it as JSON into
// the given interface format.  It's mostly there as a minor simplifying
// function.
func ReadBodyJSON(output interface{}, body io.ReadCloser) ([]byte, error) {
	rBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}	

	err = json.Unmarshal(rBytes, output)	
	return rBytes, err
}