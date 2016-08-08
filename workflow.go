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
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"

	"github.com/venicegeo/pz-gocommon/elasticsearch"
	"github.com/venicegeo/pz-gocommon/gocommon"
	"github.com/venicegeo/pz-workflow/workflow"
)

var (
	eventTypeMap = make(map[string]*workflow.EventType)
)

// EventType returns the event type ID and fully qualified name
// for the specified EventType and its root
func EventType(root string, mapping map[string]elasticsearch.MappingElementTypeName, auth string) (*workflow.EventType, error) {
	var (
		request      *http.Request
		response     *http.Response
		err          error
		etBytes      []byte
		eventTypes   []workflow.EventType
		jsonResponse piazza.JsonResponse
		result       *workflow.EventType
		ok           bool
	)
	if result, ok = eventTypeMap[root]; !ok {
		requestURL := "https://pz-gateway." + domain + "/eventType?perPage=10000"
		log.Print(requestURL)
		if request, err = http.NewRequest("GET", requestURL, nil); err != nil {
			return result, err
		}
		request.Header.Set("Authorization", auth)
		if response, err = HTTPClient().Do(request); err != nil {
			return result, err
		}

		// Check for HTTP errors
		if response.StatusCode < 200 || response.StatusCode > 299 {
			return result, &HTTPError{Status: response.StatusCode, Message: "Failed to retrieve harvest event ID: " + response.Status}
		}

		defer response.Body.Close()
		if etBytes, err = ioutil.ReadAll(response.Body); err != nil {
			return result, err
		}

		if err = json.Unmarshal(etBytes, &jsonResponse); err != nil {
			return result, err
		}

		if etBytes, err = json.Marshal(jsonResponse.Data); err != nil {
			return result, err
		}

		if err = json.Unmarshal(etBytes, &eventTypes); err != nil {
			return result, err
		}

		// Look for an event type with the same root and same mapping
		// and add the EventType if needed
		for version := 0; ; version++ {
			foundMatch := false
			eventTypeName := fmt.Sprintf("%v:%v", root, version)
			for _, eventType := range eventTypes {
				if eventType.Name == eventTypeName {
					foundMatch = true
					if reflect.DeepEqual(eventType.Mapping, mapping) {
						log.Printf("Found match for %v", eventTypeName)
						result = &eventType
						break
					}
				}
			}
			if result != nil {
				break
			}
			if !foundMatch {
				log.Printf("Found no match for %v; adding.", eventTypeName)
				eventType := workflow.EventType{Name: eventTypeName, Mapping: mapping}
				if result, err = AddEventType(eventType, auth); err == nil {
					break
				} else {
					return result, err
				}
			}
		}

		if result != nil {
			eventTypeMap[root] = result
		}
	}
	return result, nil
}

// AddEventType adds the requested EventType and returns a pointer to what was created
func AddEventType(eventType workflow.EventType, auth string) (*workflow.EventType, error) {
	var (
		request        *http.Request
		response       *http.Response
		err            error
		eventTypeBytes []byte
		result         *workflow.EventType
	)
	if eventTypeBytes, err = json.Marshal(&eventType); err != nil {
		return result, err
	}

	requestURL := "https://pz-gateway." + domain + "/eventType"
	if request, err = http.NewRequest("POST", requestURL, bytes.NewBuffer(eventTypeBytes)); err != nil {
		return result, err
	}

	request.Header.Set("Authorization", auth)
	request.Header.Set("Content-Type", "application/json")

	log.Print(string(eventTypeBytes))
	if response, err = HTTPClient().Do(request); err != nil {
		return result, err
	}

	// Check for HTTP errors
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return result, &HTTPError{Status: response.StatusCode, Message: "Failed to add requested event type: " + response.Status}
	}

	defer response.Body.Close()
	if eventTypeBytes, err = ioutil.ReadAll(response.Body); err != nil {
		return result, err
	}

	log.Print(eventTypeBytes)
	result = new(workflow.EventType)
	err = json.Unmarshal(eventTypeBytes, result)
	return result, err
}

// Events returns the events for the event type ID provided
func Events(eventTypeID piazza.Ident, auth string) ([]workflow.Event, error) {

	var (
		request     *http.Request
		response    *http.Response
		err         error
		eventsBytes []byte
		result      []workflow.Event
	)

	requestURL := "https://pz-gateway." + domain + "/event?eventTypeId=" + string(eventTypeID)
	if request, err = http.NewRequest("GET", requestURL, nil); err != nil {
		return result, err
	}

	request.Header.Set("Authorization", auth)

	log.Print(requestURL)
	if response, err = HTTPClient().Do(request); err != nil {
		return result, err
	}

	// Check for HTTP errors
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return result, &HTTPError{Status: response.StatusCode, Message: "Failed to get available events: " + response.Status}
	}

	defer response.Body.Close()
	if eventsBytes, err = ioutil.ReadAll(response.Body); err != nil {
		return result, err
	}

	log.Print(string(eventsBytes))
	var jsonResponse piazza.JsonResponse
	if err = json.Unmarshal(eventsBytes, &jsonResponse); err != nil {
		return result, err
	}
	result = make([]workflow.Event, 10)
	err = jsonResponse.ExtractData(&result)
	return result, err
}

// AddEvent adds the requested Event and returns a pointer to what was created
// TODO: Fix this (or prove that it works)
func AddEvent(event workflow.Event, auth string) (*workflow.Event, error) {
	var (
		request    *http.Request
		response   *http.Response
		err        error
		eventBytes []byte
		result     *workflow.Event
	)
	if eventBytes, err = json.Marshal(&event); err != nil {
		return result, err
	}

	requestURL := "https://pz-gateway." + domain + "/event"
	if request, err = http.NewRequest("POST", requestURL, bytes.NewBuffer(eventBytes)); err != nil {
		return result, err
	}

	request.Header.Set("Authorization", auth)
	request.Header.Set("Content-Type", "application/json")

	log.Print(string(eventBytes))
	if response, err = HTTPClient().Do(request); err != nil {
		return result, err
	}

	// Check for HTTP errors
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return result, &HTTPError{Status: response.StatusCode, Message: "Failed to add requested event: " + response.Status}
	}

	defer response.Body.Close()
	if eventBytes, err = ioutil.ReadAll(response.Body); err != nil {
		return result, err
	}

	log.Print(string(eventBytes))
	result = new(workflow.Event)
	err = json.Unmarshal(eventBytes, result)
	return result, err
}
