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
	"net/http"
	"net/url"
)

// FindMySvc Searches Pz for a service matching the input information.  If it finds
// one, it returns the service ID.  If it does not, returns an empty string.  Currently
// searches on service name and submitting user.
func FindMySvc(svcName, pzAddr, authKey string) (string, error) {
	query := pzAddr + "/service/me?per_page=1000&keyword=" + url.QueryEscape(svcName)
	var respObj SvcList
	_, err := RequestKnownJSON("GET", "", query, authKey, &respObj)
	if err != nil {
		return "", AddRef(err)
	}

	for _, checkServ := range respObj.Data {
		if checkServ.ResMeta.Name == svcName {
			return checkServ.ServiceID, nil
		}
	}

	return "", nil
}

// ManageRegistration Handles Pz registration for a service.  It checks the current
// service list to see if it has been registered already.  If it has not, it performs
// initial registration.  If it has not, it re-registers.  Best practice is to do this
// every time your service starts up.  For those of you code-reading, the filter is
// still somewhat rudimentary.  It will improve as better tools become available.
func ManageRegistration(svcName, svcDesc, svcURL, pzAddr, svcVers, authKey string,
	attributes map[string]string) error {

	fmt.Println("Finding")
	svcID, err := FindMySvc(svcName, pzAddr, authKey)
	if err != nil {
		return AddRef(err)
	}

	svcClass := ClassType{"UNCLASSIFIED"} // TODO: this will have to be updated at some point.
	metaObj := ResMeta{Name: svcName,
		Description: svcDesc,
		ClassType:   svcClass,
		Version:     svcVers,
		Metadata:    make(map[string]string)}
	for key, val := range attributes {
		metaObj.Metadata[key] = val
	}
	svcObj := Service{ServiceID: svcID, URL: svcURL, Method: "POST", ResMeta: metaObj}
	svcJSON, err := json.Marshal(svcObj)

	if svcID == "" {
		fmt.Println("Registering")
		_, err = SubmitSinglePart("POST", string(svcJSON), pzAddr+"/service", authKey)
	} else {
		fmt.Println("Updating")
		_, err = SubmitSinglePart("PUT", string(svcJSON), pzAddr+"/service/"+svcID, authKey)
	}
	if err != nil {
		return AddRef(err)
	}

	return nil
}

// ExecIn is a structure designed to contain all of the information necessary
// to call pzsvc-exec.
type ExecIn struct {
	FuncStr    string
	InFiles    []string
	OutGeoJSON []string
	OutGeoTIFF []string
	OutTxt     []string
	AlgoURL    string
	AuthKey    string
}

// ExecOut represents the output of the standard pzsvc-exec call.
type ExecOut struct {
	InFiles    map[string]string
	OutFiles   map[string]string
	ProgReturn string
	Errors     []string
}

// CallPzsvcExec is a function designed to simplify calls to pzsvc-exec.
// Fill out the inpObj properly, and it'll go through the contact process,
// returning the OutFiles mapping (as that is generally what people are
// interested in, one way or the other)
func CallPzsvcExec(inpObj *ExecIn) (*ExecOut, error) {

	var formVal url.Values

	formVal = make(map[string][]string)
	formVal.Set("cmd", inpObj.FuncStr)
	formVal.Set("inFiles", SliceToCommaSep(inpObj.InFiles))
	formVal.Set("outGeoJson", SliceToCommaSep(inpObj.OutGeoJSON))
	formVal.Set("outTiffs", SliceToCommaSep(inpObj.OutGeoTIFF))
	formVal.Set("outTxt", SliceToCommaSep(inpObj.OutTxt))
	formVal.Set("authKey", inpObj.AuthKey)
	fmt.Println(inpObj.FuncStr)

	resp, err := http.PostForm(inpObj.AlgoURL, formVal)
	if err != nil {
		return nil, fmt.Errorf(`PostForm: %s`, err.Error())
	}

	var respObj ExecOut
	var respBytes []byte
	respBytes, err = ReadBodyJSON(&respObj, resp.Body)
	if err != nil {
		errString := fmt.Sprintf(`Unmarshalling error: %s.  Json: %s`, err.Error(), string(respBytes))
		fmt.Printf(errString)
		return nil, errWithRef(errString)
	}

	if len(respObj.Errors) != 0 {
		return nil, fmt.Errorf(`pzsvc-exec errors: %s`, SliceToCommaSep(respObj.Errors))
	}

	fmt.Println("output:")
	fmt.Println(string(respBytes))

	return &respObj, nil
}
