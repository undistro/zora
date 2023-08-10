// Copyright 2023 Undistro Authors
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

package worker

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()

	okPath := filepath.Join(tmpDir, "ok")
	if err := os.WriteFile(okPath, []byte("ok"), 0644); err != nil {
		t.Fatal(err)
	}

	noPermPath := filepath.Join(tmpDir, "noperm")
	if err := os.WriteFile(noPermPath, []byte("noperm"), 0000); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name           string
		filename       string
		wantFileExists bool
		wantErr        bool
	}{
		{
			name:           "dir",
			filename:       tmpDir,
			wantFileExists: false,
			wantErr:        true,
		},
		{
			name:           "ok",
			filename:       okPath,
			wantFileExists: true,
			wantErr:        false,
		},
		{
			name:           "exists without permission",
			filename:       noPermPath,
			wantFileExists: true,
			wantErr:        false,
		},
		{
			name:           "not exists",
			filename:       filepath.Join(tmpDir, "results"),
			wantFileExists: false,
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fileExists(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("fileExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantFileExists {
				t.Errorf("fileExists() got = %v, want %v", got, tt.wantFileExists)
			}
		})
	}
}

func TestReadResultsFile(t *testing.T) {
	tmpDir := t.TempDir()

	noPermFile := filepath.Join(tmpDir, "noperm")
	if err := os.WriteFile(noPermFile, []byte("noperm"), 0000); err != nil {
		t.Fatal(err)
	}

	resultsFile := filepath.Join(tmpDir, "results")
	if err := os.WriteFile(resultsFile, []byte("report"), 0644); err != nil {
		t.Fatal(err)
	}

	doneFile := filepath.Join(tmpDir, "done")
	if err := os.WriteFile(doneFile, []byte(resultsFile), 0644); err != nil {
		t.Fatal(err)
	}

	// done file pointing to a file without permission
	doneFile2NoPerm := filepath.Join(tmpDir, "done2noperm")
	if err := os.WriteFile(doneFile2NoPerm, []byte(noPermFile), 0644); err != nil {
		t.Fatal(err)
	}

	// done file pointing to a directory
	doneFile2Dir := filepath.Join(tmpDir, "done2dir")
	if err := os.WriteFile(doneFile2Dir, []byte(tmpDir), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		doneFile string
		want     string
		wantErr  bool
	}{
		{
			name:     "ok",
			doneFile: doneFile,
			want:     "report",
			wantErr:  false,
		},
		{
			name:     "done file is a directory",
			doneFile: tmpDir,
			want:     "",
			wantErr:  true,
		},
		{
			name:     "done file without permission",
			doneFile: noPermFile,
			want:     "",
			wantErr:  true,
		},
		{
			name:     "results file without permissions",
			doneFile: doneFile2NoPerm,
			want:     "",
			wantErr:  true,
		},
		{
			name:     "results file is a directory",
			doneFile: doneFile2Dir,
			want:     "",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := readResultsFile(tt.doneFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("readResultsFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got := readerToString(reader)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readResultsFile() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func readerToString(r io.Reader) string {
	if r == nil {
		return ""
	}
	b, err := io.ReadAll(r)
	if err != nil {
		return err.Error()
	}
	return string(b)
}
