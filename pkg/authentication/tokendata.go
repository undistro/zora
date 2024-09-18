// Copyright 2024 Undistro Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package authentication

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

type TokenData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

func ParseTokenData(data []byte) (*TokenData, error) {
	if data == nil {
		return nil, nil
	}
	var tokenData TokenData
	err := json.Unmarshal(data, &tokenData)
	if err != nil {
		return nil, err
	}
	return &tokenData, nil
}

func GetJWTExpiry(tokenData *TokenData) (time.Time, error) {
	if tokenData == nil {
		return time.Time{}, errors.New("Missing token data")
	}

	token := tokenData.AccessToken
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return time.Time{}, errors.New("invalid token format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return time.Time{}, err
	}

	var claims jwt.MapClaims
	err = json.Unmarshal(payload, &claims)
	if err != nil {
		return time.Time{}, err
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return time.Time{}, errors.New("invalid expiry claim")
	}

	return time.Unix(int64(exp), 0), nil
}
