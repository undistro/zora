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

//go:build linux
// +build linux

package filemonitor

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileMonitorMonitorLink(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "example")
	if err != nil {
		t.Fatal("Failed to create temp dir:", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpfileName := filepath.Join(tmpDir, "test")

	// Write initial content
	initialContent := "Initial content"
	err = os.WriteFile(tmpfileName, []byte(initialContent), 0644)
	if err != nil {
		t.Fatal("Failed to write to temp file:", err)
	}

	linkFileName := tmpfileName + "_link"
	fm := NewFileMonitor(linkFileName, processContent)

	// Start monitoring in a goroutine
	done := make(chan struct{})
	defer close(done)

	go fm.MonitorFile(done)

	// Give some time for the initial read
	time.Sleep(100 * time.Millisecond)

	if fm.GetContent() != nil {
		t.Errorf("Expected empty content, got %q", fm.GetContent())
	}

	// create the link
	os.Symlink(tmpfileName, linkFileName)

	// Wait for the link to be detected
	time.Sleep(100 * time.Millisecond)

	if fm.GetContent() != initialContent {
		t.Errorf("Expected initial content %q, got %q", initialContent, fm.GetContent())
	}

	// Update file content
	newContent := "Updated content"
	err = os.WriteFile(tmpfileName, []byte(newContent), 0644)
	if err != nil {
		t.Fatal("Failed to write updated content:", err)
	}

	// Wait for the file change to be detected
	time.Sleep(100 * time.Millisecond)

	if fm.GetContent() != newContent {
		t.Errorf("Expected updated content %q, got %q", newContent, fm.GetContent())
	}

	// Remove the link
	err = os.Remove(linkFileName)
	if err != nil {
		t.Fatal("Failed to remove link:", err)
	}

	// Wait for the file deletion to be detected
	time.Sleep(100 * time.Millisecond)

	if fm.GetContent() != nil {
		t.Errorf("Expected empty content, got %q", fm.GetContent())
	}

	// Write new file contents
	recreatedFileContent := "Recreated content"
	newTmpFileName := tmpfileName + "-2"
	err = os.WriteFile(newTmpFileName, []byte(recreatedFileContent), 0644)
	if err != nil {
		t.Fatal("Failed to write to temp file:", err)
	}

	// create the new link
	os.Symlink(newTmpFileName, linkFileName)

	// Wait for the link to be detected
	time.Sleep(100 * time.Millisecond)

	if fm.GetContent() != recreatedFileContent {
		t.Errorf("Expected recreated content %q, got %q", recreatedFileContent, fm.GetContent())
	}
}
