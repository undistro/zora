// Copyright 2022 Undistro Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handlers

import (
	"encoding/json"
	"net/http"
)

type responseError struct {
	Error  string `json:"error,omitempty"`
	Detail string `json:"detail,omitempty"`
}

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		response, _ = json.Marshal(responseError{Error: err.Error()})
		code = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(response)
}

func RespondWithError(w http.ResponseWriter, code int, message string) {
	RespondWithJSON(w, code, responseError{Error: message})
}

func RespondWithDetailedError(w http.ResponseWriter, code int, message, detail string) {
	RespondWithJSON(w, code, responseError{Error: message, Detail: detail})
}

func RespondWithCode(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
}
