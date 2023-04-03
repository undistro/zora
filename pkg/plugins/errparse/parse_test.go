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

package errparse

import (
	"io"
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	type args struct {
		file   string
		plugin string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "invalid plugin",
			args:    args{plugin: "foo"},
			wantErr: true,
		},
		{
			name:    "no plugin",
			args:    args{plugin: ""},
			wantErr: true,
		},
		{
			name:    "no file",
			args:    args{plugin: "popeye", file: ""},
			wantErr: true,
		},
		{
			name:    "empty file",
			args:    args{plugin: "popeye", file: "testdata/dummy_err_1.txt"},
			wantErr: true,
		},
		{
			name: "popeye invalid credentials",
			args: args{plugin: "popeye", file: "testdata/popeye_err_1.txt"},
			want: "the server has asked for the client to provide credentials",
		},
		{
			name: "popeye invalid cluster",
			args: args{plugin: "popeye", file: "testdata/popeye_err_2.txt"},
			want: `Get "http://localhost:8080/version?timeout=30s": dial tcp 127.0.0.1:8080: connect: connection refused`,
		},
		{
			name: "popeye invalid context",
			args: args{plugin: "popeye", file: "testdata/popeye_err_3.txt"},
			want: "invalid configuration: context was not found for specified context: ctx",
		},
		{
			name: "popeye unknown flag",
			args: args{plugin: "popeye", file: "testdata/popeye_err_4.txt"},
			want: "Exec failed unknown flag: --brokenflag",
		},
		{
			name: "marvin invalid cluster",
			args: args{plugin: "marvin", file: "testdata/marvin_err_1.txt"},
			want: `server version error: Get "http://localhost:8080/version?timeout=32s": dial tcp 127.0.0.1:8080: connect: connection refused`,
		},
		{
			name: "marvin invalid kubeconfig file",
			args: args{plugin: "marvin", file: "testdata/marvin_err_2.txt"},
			want: "dynamic client error: stat foo: no such file or directory",
		},
		{
			name: "marvin unknown flag",
			args: args{plugin: "marvin", file: "testdata/marvin_err_3.txt"},
			want: "unknown flag: --foo",
		},
		{
			name: "marvin invalid credentials",
			args: args{plugin: "marvin", file: "testdata/marvin_err_4.txt"},
			want: "server version error: the server has asked for the client to provide credentials",
		},
		{
			name: "marvin invalid flags",
			args: args{plugin: "marvin", file: "testdata/marvin_err_5.txt"},
			want: "please set '--checks/-f' or keep 'disable-builtin' 'false'",
		},
		{
			name: "marvin compile error",
			args: args{plugin: "marvin", file: "testdata/marvin_err_6.txt"},
			want: `failed to compile check M-002: type-check error on validation 0: ERROR: <input>:1:5: Syntax error: mismatched input 'allContainers' expecting <EOF>\n | foo allContainers.all(container,\n | ....^`,
		},
		{
			name: "marvin multiples errors",
			args: args{plugin: "marvin", file: "testdata/marvin_err_7.txt"},
			want: `failed to compile check M-001: cel expression must evaluate to a bool on validation 0`,
		},
		{
			name: "marvin list error",
			args: args{plugin: "marvin", file: "testdata/marvin_err_8.txt"},
			want: `failed to list v1/podx: the server could not find the requested resource`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r io.Reader
			if tt.args.file != "" {
				f, err := os.Open(tt.args.file)
				if err != nil {
					t.Fatal(err)
				}
				r = f
			}
			got, err := Parse(r, tt.args.plugin)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Parse() got = %v, want %v", got, tt.want)
			}
		})
	}
}
