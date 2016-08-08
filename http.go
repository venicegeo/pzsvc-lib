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
	"io/ioutil"
	"net/http"
	"os"

	"github.com/venicegeo/pz-gocommon/gocommon"
)

var (
	domain = os.Getenv("DOMAIN")
)

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

// GetGateway performs a GET request on the Piazza Gateway
func GetGateway(endpoint, pzAuth string, target interface{}) error {
	var (
		result       []byte
		err          error
		request      *http.Request
		response     *http.Response
		jsonResponse piazza.JsonResponse
	)

	requestURL := "https://pz-gateway." + domain + endpoint
	if request, err = http.NewRequest("GET", requestURL, nil); err != nil {
		return &HTTPError{Status: http.StatusInternalServerError, Message: "Unable to create new GET request: " + err.Error()}
	}
	request.Header.Set("Authorization", pzAuth)

	if response, err = HTTPClient().Do(request); err != nil {
		return &HTTPError{Status: http.StatusInternalServerError, Message: "Unable to do GET request: " + err.Error()}
	}

	defer response.Body.Close()
	if result, err = ioutil.ReadAll(response.Body); err != nil {
		return &HTTPError{Status: http.StatusInternalServerError, Message: "GET request failed and could not retrieve the message body: " + err.Error()}
	}

	// Check for HTTP errors
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return &HTTPError{Status: response.StatusCode, Message: "GET request failed:\n" + string(result)}
	}

	if err = json.Unmarshal(result, &jsonResponse); err != nil {
		return &HTTPError{Status: response.StatusCode, Message: "Failed to unmarhsal GET response:\n" + err.Error()}
	}

	if target != nil {
		return jsonResponse.ExtractData(target)
	}
	return nil
}

// PostGateway performs a GET request on the Piazza Gateway and returns a []byte
func PostGateway(endpoint string, body []byte, pzAuth string) ([]byte, error) {
	var (
		result   []byte
		err      error
		request  *http.Request
		response *http.Response
	)
	requestURL := "https://pz-gateway." + domain + endpoint
	if request, err = http.NewRequest("POST", requestURL, bytes.NewBuffer(body)); err != nil {
		return nil, &HTTPError{Status: http.StatusInternalServerError, Message: "Unable to create new POST request: " + err.Error()}
	}
	request.Header.Set("Authorization", pzAuth)
	request.Header.Set("Content-Type", "application/json")

	if response, err = HTTPClient().Do(request); err != nil {
		return nil, &HTTPError{Status: http.StatusInternalServerError, Message: "Unable to do POST request: " + err.Error()}
	}

	defer response.Body.Close()
	if result, err = ioutil.ReadAll(response.Body); err != nil {
		return nil, &HTTPError{Status: http.StatusInternalServerError, Message: "POST request failed and could not retrieve the message body: " + err.Error()}
	}

	// Check for HTTP errors
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return result, &HTTPError{Status: response.StatusCode, Message: "POST request failed:\n" + string(result)}
	}

	return result, err
}
