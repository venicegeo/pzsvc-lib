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
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strconv"

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
		err        error
		eventTypes []workflow.EventType
		result     *workflow.EventType
		ok         bool
	)
	if result, ok = eventTypeMap[root]; !ok {
		if err = GetGateway("/eventType?perPage=10000", auth, &eventTypes); err != nil {
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
		err            error
		eventTypeBytes []byte
		result         *workflow.EventType
	)
	if eventTypeBytes, err = json.Marshal(&eventType); err != nil {
		return result, err
	}

	if eventTypeBytes, err = PostGateway("/eventType", eventTypeBytes, auth); err != nil {
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
		err    error
		result []workflow.Event
	)

	err = GetGateway("/event?eventTypeId="+string(eventTypeID), auth, &result)

	return result, err
}

// AddEvent adds the requested Event and returns a pointer to what was created
func AddEvent(event workflow.Event, auth string) (*workflow.Event, error) {
	var (
		err        error
		eventBytes []byte
		result     *workflow.Event
	)
	if eventBytes, err = json.Marshal(&event); err != nil {
		return result, err
	}

	if eventBytes, err = PostGateway("/event", eventBytes, auth); err != nil {
		return result, err
	}
	result = new(workflow.Event)
	err = json.Unmarshal(eventBytes, result)
	return result, err
}

// GetAlerts will return the group of alerts associated with the given trigger ID,
// under the given pagination. 
func GetAlerts (perPage, pageNo int, trigID, pzAddr, pzAuth string) ([]Alert, error) {

	qParams := "triggerId=" + trigID + "&sortBy=createdOn&order=desc"
	if perPage != 0 {
		qParams += "&perPage=" + strconv.Itoa(perPage)
	}
	if pageNo != 0 {
		qParams += "&page=" + strconv.Itoa(pageNo)
	}

	var outpObj AlertList

	if _, err := RequestKnownJSON("GET", "", pzAddr + "/alert?" + qParams, pzAuth, &outpObj); err != nil {
		return nil, fmt.Errorf("Error: pzsvc.RequestKnownJSON: fail on alert check: " + err.Error())
	}
	return outpObj.Data, nil
}