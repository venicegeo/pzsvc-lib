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
	//	"encoding/json"
	//	"errors"
	//	"fmt"
	"io"
	//  "io/ioutil"
	//	"mime/multipart"
	"net/http"
	//	"net/url"
	//  "strconv"
	//  "testing"
	//	"time"
)

type testRC struct{ io.Reader }

func (testRC) Close() error { return nil }

type stringSliceMockTransport struct {
	statusCode int
	outputs    []string
	iter       *int
}

type fakeRespWriter struct{}

// SetMockClient is a utility function for testing purposes.  When called,
// it sets the current pzsvc client to a mock client.  That client will accept
// any input, and immediately respond.  It's responses are determined by the
// given string slice - it iterates through them one at a time until done,
// with the http status as whatever is provided as such, and returns the
// rudimentary string "{}" for any remaining calls.  Intended to simplify
// testing in the face of a requirement for multiple http calls.
// Explicitly including an empty string in the list will permit a single
// call to pass out through the standard path.
func SetMockClient(outputs []string, status int) {
	client := http.Client{}
	iter := 0
	client.Transport = stringSliceMockTransport{status, outputs, &iter}
	SetHTTPClient(&client)
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
		if t.outputs[*t.iter] == "" {
			*t.iter = *t.iter + 1

			tempTransport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
			return tempTransport.RoundTrip(req)
		}
		response.Body = testRC{bytes.NewBufferString(t.outputs[*t.iter])}
		*t.iter = *t.iter + 1
	}
	return response, nil
}

// GetMockResponseWriter returns a rudimentary fake response writer object
// for testing purposes
func GetMockResponseWriter() http.ResponseWriter {
	return fakeRespWriter{}
}

func (frw fakeRespWriter) Header() http.Header {
	return nil
}

func (frw fakeRespWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (frw fakeRespWriter) WriteHeader(int) {
}
