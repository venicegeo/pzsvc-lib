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
	//	"encoding/json"
	//	"errors"
	//	"fmt"
	"io"
	"io/ioutil"
	//	"mime/multipart"
	"net/http"
	//	"net/url"
	"strconv"
	"testing"
	//	"time"
)

type testRC struct{ io.Reader }

func (testRC) Close() error { return nil }

type stringSliceMockTransport struct {
	statusCode int
	outputs    []string
	iter       *int
}

func (t stringSliceMockTransport) RoundTrip(req *http.Request) (*http.Response, error) {

	response := &http.Response{
		Header:     make(http.Header),
		Request:    req,
		StatusCode: t.statusCode,
	}
	response.Header.Set("Content-Type", "application/json")

	if t.outputs == nil || *t.iter >= len(t.outputs) {
		response.Body = testRC{bytes.NewBufferString("{}")}
	} else {
		response.Body = testRC{bytes.NewBufferString(t.outputs[*t.iter])}
		*t.iter = *t.iter + 1
	}
	return response, nil
}

func TestSubmitSinglePart(t *testing.T) {
	testClient := http.Client{}
	iterBase := 0
	testClient.Transport = stringSliceMockTransport{statusCode: 250, outputs: nil, iter: &iterBase}
	SetHTTPClient(&testClient)
	method := "TRACE"
	url := "http://testURL.net"
	bodyStr := "testBody"
	authKey := "testAuthKey"

	resp, err := SubmitSinglePart(method, bodyStr, url, authKey)
	if err != nil {
		t.Error(`received error on basic test of SubmitSinglePart.  Error message: `, err.Error())
	} else {
		req := (*http.Request)(resp.Request)
		if req.Header.Get("Content-Type") != "application/json" {
			t.Error(`SubmitSinglePart: Content-Type not application/json.`)
		}
		if req.Header.Get("Authorization") != authKey {
			t.Error(`SubmitSinglePart: Authorization not sustained properly.`)
		}
		if req.URL.String() != url {
			t.Error(`SubmitSinglePart: URL not sustained properly.`)
		}
		if req.Method != method {
			t.Error(`SubmitSinglePart: method not sustained properly.`)
		}
		bodyBytes, _ := ioutil.ReadAll(req.Body)
		if string(bodyBytes) != bodyStr {
			t.Error(`SubmitSinglePart: body string not sustained properly.`)
		}
	}

	testClient.Transport = stringSliceMockTransport{statusCode: 500, outputs: nil}
	resp, err = SubmitSinglePart(method, bodyStr, url, authKey)
	if err == nil {
		t.Error(`SubmitSinglePart: did not respond to http status error properly.`)
	}

	testClient.Transport = stringSliceMockTransport{statusCode: 100, outputs: nil}
	resp, err = SubmitSinglePart(method, bodyStr, url, authKey)
	if err == nil {
		t.Error(`SubmitSinglePart: did not respond to http status error properly.`)
	}
}

func TestSubmitMultipart(t *testing.T) {
	testClient := http.Client{}
	testClient.Transport = stringSliceMockTransport{statusCode: 250, outputs: nil}
	SetHTTPClient(&testClient)
	bodyStr := "testBody"
	url := "http://testURL.net"
	fileName := "name"
	authKey := "testAuthKey"
	testData := []byte("testtesttest")

	_, err := SubmitMultipart(bodyStr, url, fileName, authKey, testData)
	if err != nil {
		t.Error(`TestSubmitMultipart: failed on what shoudl have been good run.`)
	}
	iterBase := 0
	testClient.Transport = stringSliceMockTransport{statusCode: 550, outputs: nil, iter: &iterBase}
	_, err = SubmitMultipart(bodyStr, url, fileName, authKey, testData)
	if err == nil {
		t.Error(`TestSubmitMultipart: passed on what should have been bad status code.`)
	}

}

func TestRequestKnownJSON(t *testing.T) {
	testClient := http.Client{}
	outStrs := []string{
		`{"PercentComplete":0, "TimeRemaining":"blah", "TimeSpent":"blah"}`,
		`XXXXX`,
	}
	iterBase := 0
	testClient.Transport = stringSliceMockTransport{statusCode: 250, outputs: outStrs, iter: &iterBase}
	SetHTTPClient(&testClient)
	method := "TRACE"
	url := "http://testURL.net"
	bodyStr := "testBody"
	authKey := "testAuthKey"
	var jobProgObj JobProg
	_, err := RequestKnownJSON(method, bodyStr, url, authKey, &jobProgObj)
	if err != nil {
		t.Error(`TestRequestKnownJSON: failed on what shoudl have been clean run.`)
	}
	_, err = RequestKnownJSON(method, bodyStr, url, authKey, &jobProgObj)
	if err == nil {
		t.Error(`TestRequestKnownJSON: passed on what should have been bad JSON.`)
	}
}

func TestGetJobResponse(t *testing.T) {
	testClient := http.Client{}
	outStrs := []string{
		`{"Data":{"Status":"Error", "Result":{"Message":"Job Not Found."}}}`,
		`{"Data":{"Status":"Submitted"}}`,
		`{"Data":{"Status":"Pending"}}`,
		`{"Data":{"Status":"Running"}}`,
		`{"Data":{"Status":"Success"}}`,
		`{"Data":{"Status":"Success", "Result":{"Message":"Job Found."}}}`,
		`{"Data":{"Status":"Fail"}}`,
		`{"Data":{"Status":"Error", "Result":{"Message":"Everything Broken."}}}`,
		`{"Data":{"Status":"Nope"}}`}
	iterBase := 0
	testClient.Transport = stringSliceMockTransport{statusCode: 250, outputs: outStrs, iter: &iterBase}
	SetHTTPClient(&testClient)
	url := "http://testURL.net"
	jobID := "testJobID"
	authKey := "testAuthKey"

	_, err := GetJobResponse(jobID, url, authKey)
	iter := testClient.Transport.(stringSliceMockTransport).iter
	if err != nil {
		t.Error(`TestGetJobResponse: failed incorrectly - attempt #` + strconv.Itoa(*iter) + `.`)
	} else if *iter != 6 {
		t.Error(`TestGetJobResponse: passed on wrong entry - attempt #` + strconv.Itoa(*iter) + `.`)
	} else {
		_, err = GetJobResponse(jobID, url, authKey)
		if err == nil {
			t.Error(`TestGetJobResponse: passed on Fail.`)
		}
		_, err = GetJobResponse(jobID, url, authKey)
		if err == nil {
			t.Error(`TestGetJobResponse: passed on Error.`)
		}
		_, err = GetJobResponse(jobID, url, authKey)
		if err == nil {
			t.Error(`TestGetJobResponse: passed on non-status.`)
		}
	}
}

func TestGetJobID(t *testing.T) {

	testID := "testID"
	bStrings := []string{`b`, `{"PercentComplete":50}`, `{"Data":{"JobID":"` + testID + `"}}`}

	for i, bstr := range bStrings {
		testBody := testRC{bytes.NewBufferString(bstr)}
		testResp := http.Response{Body: testBody}
		testStr, err := GetJobID(&testResp)
		switch i {
		case 0, 1:
			if err == nil {
				t.Error("GetJobID did not throw error on test ", i, ".")
			}
		case 2:
			if err != nil || testStr != testID {
				t.Error("GetJobID did not properly return given JobID on test ", i, ".")
			}
		}
	}
}

func TestReadBodyJSON(t *testing.T) {

	bStrings := []string{``, `b`, `{}`, `{"PercentComplete":50}`}

	for i, bstr := range bStrings {
		var jp JobProg
		body := testRC{bytes.NewBufferString(bstr)}
		_, err := ReadBodyJSON(&jp, body)
		if i < 2 && err == nil {
			t.Error("ReadBodyJson did not throw error on test ", i)
		}
		if i >= 2 && err != nil {
			t.Error("ReadBodyJson threw error on test ", i, ".  Error: ", err.Error())
		}
	}
}
