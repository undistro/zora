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

package filemonitor

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func processContent(content []byte) (any, error) {
	if content == nil {
		return nil, nil
	} else {
		return string(content), nil
	}
}

func TestFileMonitorUpdateContent(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "example")
	if err != nil {
		t.Fatal("Failed to create temp dir:", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpfileName := filepath.Join(tmpDir, "test")

	// Write initial content
	initialContent := "Hello, World!"
	err = os.WriteFile(tmpfileName, []byte(initialContent), 0644)
	if err != nil {
		t.Fatal("Failed to write to temp file:", err)
	}

	fm := NewFileMonitor(tmpfileName, processContent)
	err = fm.updateContent()
	if err != nil {
		t.Fatal("Failed to update content:", err)
	}

	if fm.GetContent() != initialContent {
		t.Errorf("Expected content %q, got %q", initialContent, fm.GetContent())
	}

	// Update file content
	newContent := "Updated content"
	err = os.WriteFile(tmpfileName, []byte(newContent), 0644)
	if err != nil {
		t.Fatal("Failed to write updated content:", err)
	}

	err = fm.updateContent()
	if err != nil {
		t.Fatal("Failed to update content:", err)
	}

	if fm.GetContent() != newContent {
		t.Errorf("Expected updated content %q, got %q", newContent, fm.GetContent())
	}
}

func TestFileMonitorMonitorFile(t *testing.T) {
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

	fm := NewFileMonitor(tmpfileName, processContent)

	// Start monitoring in a goroutine
	done := make(chan struct{})
	defer close(done)

	go fm.MonitorFile(done)

	// Give some time for the initial read
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

	// Remove the file
	err = os.Remove(tmpfileName)
	if err != nil {
		t.Fatal("Failed to remove file:", err)
	}

	// Wait for the file deletion to be detected
	time.Sleep(100 * time.Millisecond)

	if fm.GetContent() != nil {
		t.Errorf("Expected empty content, got %q", fm.GetContent())
	}

	// Write recreated file contents
	recreatedFileContent := "Recreated content"
	err = os.WriteFile(tmpfileName, []byte(recreatedFileContent), 0644)
	if err != nil {
		t.Fatal("Failed to write to temp file:", err)
	}

	// Wait for the file change to be detected
	time.Sleep(100 * time.Millisecond)

	if fm.GetContent() != recreatedFileContent {
		t.Errorf("Expected recreated content %q, got %q", recreatedFileContent, fm.GetContent())
	}

	// Test moved file
	movedFileName := filepath.Join(tmpDir, "moved")

	// Write moved file contents
	movedFileContent := "Moved content"
	err = os.WriteFile(movedFileName, []byte(movedFileContent), 0644)
	if err != nil {
		t.Fatal("Failed to write to temp file:", err)
	}
	err = os.Rename(movedFileName, tmpfileName)
	if err != nil {
		t.Fatal("Failed to rename temp file:", err)
	}

	// Wait for the file change to be detected
	time.Sleep(100 * time.Millisecond)

	if fm.GetContent() != movedFileContent {
		t.Errorf("Expected moved content %q, got %q", movedFileContent, fm.GetContent())
	}
}

func TestFileMonitorConcurrentAccess(t *testing.T) {
	testFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Errorf("Error creating temp file: %v", err)
	}
	defer testFile.Close()
	testpath := testFile.Name()

	fm := NewFileMonitor(filepath.Join(testpath), processContent)
	fm.fileContent = "Test content"

	// Simulate concurrent reads
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			content := fm.GetContent()
			if content != "Test content" {
				t.Errorf("Expected content %q, got %q", "Test content", content)
			}
		}()
	}
	wg.Wait()

	// Simulate concurrent writes
	os.WriteFile(testpath, []byte("This is a test"), 0777)
	defer os.Remove(testpath)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := fm.updateContent()
			if err != nil {
				t.Errorf("Error updating content: %v", err)
			}
		}()
	}

	wg.Wait()
}

func TestFileMonitorInvalidFile(t *testing.T) {
	// Create a directory instead of a file
	tmpdir, err := os.MkdirTemp("", "example")
	if err != nil {
		t.Fatal("Failed to create temp directory:", err)
	}
	defer os.RemoveAll(tmpdir)

	fm := NewFileMonitor(tmpdir, processContent)
	err = fm.updateContent()
	if err == nil {
		t.Error("Expected error for directory, got nil")
	}
}
