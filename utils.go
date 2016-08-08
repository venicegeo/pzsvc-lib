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
	"errors"
	"fmt"
	"runtime"
	"strconv"
)

func addRef(err error) error {
	if err != nil {
		_, file, line, ok := runtime.Caller(1)
		if ok == true {
			return errors.New(`(` + file + `, ` + strconv.Itoa(line) + `): ` + err.Error())
		}
	}
	return err
}

func errWithRef(errStr string) error {
	if errStr != "" {
		_, file, line, ok := runtime.Caller(1)
		if ok == true {
			return errors.New(`(` + file + `, ` + strconv.Itoa(line) + `): ` + errStr)
		}
	}
	return errors.New(errStr)
}

// SliceToCommaSep takes a string slice, and turns it into a comma-separated
// list of strings, suitable for JSON.
func SliceToCommaSep(inSlice []string) string {
	sliLen := len(inSlice)
	if sliLen == 0 {
		return ""
	}
	accum := inSlice[0]
	for i := 1; i < sliLen; i++ {
		accum = accum + "," + inSlice[i]
	}
	return accum
}

// HTTPError represents any HTTP error
type HTTPError struct {
	Status  int
	Message string
}

func (err HTTPError) Error() string {
	return fmt.Sprintf("%d: %v", err.Status, err.Message)
}
