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
	"runtime"
	"strconv"
)

// Error is a type designed for easy serialization to JSON
type Error struct {
	Message string `json:"error"`
}

func (err Error) Error() string {
	return err.Message
}

// TracedError returns an error that includes information
// about where it was generated
func TracedError(message string) error {
	if _, file, line, ok := runtime.Caller(1); ok {
		return errors.New(`(` + file + `, ` + strconv.Itoa(line) + `): ` + message)
	}
	return nil
}

// AddRef is a simple utility function for adding a local filename and line number
// on to the beginning of an error message before passing it along.
func AddRef(err error) error {
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
	return nil
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
