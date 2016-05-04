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
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Sends a multi-part POST call, including optional uploaded file,
// and returns the response.  Primarily intended to support Ingest calls.
func submitMultipart(bodyStr, runId, address, upload string) (*http.Response, error) {

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	err := writer.WriteField("body", bodyStr)
	if err != nil {
		return nil, err
	}

	if upload != "" {
		file, err := os.Open(fmt.Sprintf(`./%s`, upload))
		if err != nil {
			return nil, err
		}

		defer file.Close()

		part, err := writer.CreateFormFile("file", upload)
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(part, file)
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

	client := &http.Client{}
	resp, err := client.Do(fileReq)
	if err != nil {
		return nil, err
	}

	return resp, err
}

// Downloads a file from Pz using the file access API
func Download(dataId, runId, pzAddr string) (string, error) {

	resp, err := http.Get(pzAddr + "/file/" + dataId)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return "", err
	}

	contDisp := resp.Header.Get("Content-Disposition")
	_, params, err := mime.ParseMediaType(contDisp)
	fmt.Println("%+v", params)
	filename := params["filename"]
	if filename == "" {
		filename = "dummy.txt"
	}
	filepath := fmt.Sprintf(`./%s/%s`, runId, filename)

	out, err := os.Create(filepath)
	if err != nil {
		return "", err
	}

	defer out.Close()
	io.Copy(out, resp.Body)

	return filename, nil
}

// Given the JobId of an ingest call, polls job status
// until the job completes, then acquires and returns
// the resulting DataId.
func getDataId(jobId, pzAddr string) (string, error) {

	type jobResult struct {
		Type   string
		DataId string
	}

	type jobProg struct {
		PercentComplete int
	}

	// the response object for a Check Status call
	type jobResp struct {
		Type     string
		JobId    string
		Result   jobResult
		Status   string
		JobType  string
		Progress jobProg
		Message  string
		Origin   string
	}

	time.Sleep(1000 * time.Millisecond)

	for i := 0; i < 100; i++ {

		resp, err := http.Get(pzAddr + "/job/" + jobId)
		if resp != nil {
			defer resp.Body.Close()
		}
		if err != nil {
			return "", err
		}

		respBuf := &bytes.Buffer{}

		_, err = respBuf.ReadFrom(resp.Body)
		if err != nil {
			return "", err
		}

		fmt.Println(respBuf.String())

		var respObj jobResp
		err = json.Unmarshal(respBuf.Bytes(), &respObj)
		if err != nil {
			return "", err
		}

		if respObj.Status == "Submitted" || respObj.Status == "Running" || respObj.Status == "Pending" || respObj.Message == "Job Not Found" {
			time.Sleep(200 * time.Millisecond)
		} else if respObj.Status == "Success" {
			return respObj.Result.DataId, nil
		} else if respObj.Status == "Error" || respObj.Status == "Fail" {
			return "", errors.New(respObj.Status + ": " + respObj.Message)
		} else {
			return "", errors.New("Unknown status: " + respObj.Status)
		}
	}

	return "", errors.New("Never completed.")
}

// Handles the Pz Ingest process.  Will upload file to Pz and return the
// resulting DataId.
func ingestMultipart(bodyStr, runId, pzAddr, filename string) (string, error) {

	type jobResp struct {
		Type  string
		JobId string
	}

	resp, err := submitMultipart(bodyStr, runId, (pzAddr + "/job"), filename)
	if err != nil {
		return "", err
	}

	respBuf := &bytes.Buffer{}

	_, err = respBuf.ReadFrom(resp.Body)
	if err != nil {
		return "", err
	}

	fmt.Println(respBuf.String())

	var respObj jobResp
	err = json.Unmarshal(respBuf.Bytes(), &respObj)
	if err != nil {
		fmt.Println("error:", err)
	}

	return getDataId(respObj.JobId, pzAddr)
}

// Constructs the ingest call for a GeoTIFF
func IngestTiff(filename, runId, pzAddr, cmdName string) (string, error) {

	jsonStr := fmt.Sprintf(`{ "userName": "my-api-key-38n987", "jobType": { "type": "ingest", "host": "true", "data" : { "dataType": { "type": "raster" }, "metadata": { "name": "%s", "description": "raster uploaded by pzsvc-exec for %s.", "classType": { "classification": "unclassified" } } } } }`, filename, cmdName)

	return ingestMultipart(jsonStr, runId, pzAddr, filename)
}

// Constructs the ingest call for a GeoJson
func IngestGeoJson(filename, runId, pzAddr, cmdName string) (string, error) {

	jsonStr := fmt.Sprintf(`{ "userName": "my-api-key-38n987", "jobType": { "type": "ingest", "host": "true", "data" : { "dataType": { "type": "geojson" }, "metadata": { "name": "%s", "description": "GeoJson uploaded by pzsvc-exec for %s.", "classType": { "classification": "unclassified" } } } } }`, filename, cmdName)

	return ingestMultipart(jsonStr, runId, pzAddr, filename)
}

// Constructs the ingest call for standard text.
func IngestTxt(filename, runId, pzAddr, cmdName string) (string, error) {
	textblock, err := ioutil.ReadFile(fmt.Sprintf(`./%s/%s`, runId, filename))
	if err != nil {
		return "", err
	}

	jsonStr := fmt.Sprintf(`{ "userName": "my-api-key-38n987", "jobType": { "type": "ingest", "host": "true", "data" :{ "dataType": { "type": "text", "mimeType": "application/text", "content": "%s" }, "metadata": { "name": "%s", "description": "text output from pzsvc-exec for %s.", "classType": { "classification": "unclassified" } } } } }`, strconv.QuoteToASCII(string(textblock)), filename, cmdName)

	return ingestMultipart(jsonStr, runId, pzAddr, "")
}