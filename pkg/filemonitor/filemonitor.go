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
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
)

type FileContentProcessor func([]byte) (any, error)

type FileMonitor struct {
	mutex       sync.RWMutex
	fileContent any
	filePath    string
	process     FileContentProcessor
	log         logr.Logger
}

func NewFileMonitor(filePath string, process FileContentProcessor) *FileMonitor {
	log := ctrl.Log.WithValues("service", "MonitorFile", "monitored_file", filePath)
	return &FileMonitor{
		filePath: filePath,
		process:  process,
		log:      log,
	}
}

func (fm *FileMonitor) logError(err error, msg string) {
	fm.log.Error(err, msg)
}

func (fm *FileMonitor) logInfo(msg string, args ...any) {
	if len(args) > 0 {
		fm.log.Info(fmt.Sprintf(msg, args...))
	} else {
		fm.log.Info(msg)
	}
}

func (fm *FileMonitor) updateContent() error {
	content, err := os.ReadFile(fm.filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		content = nil
	}

	var contentVal any
	if fm.process != nil {
		contentVal, err = fm.process(content)
		if err != nil {
			return err
		}
	} else {
		contentVal = content
	}
	fm.mutex.Lock()
	defer fm.mutex.Unlock()
	fm.fileContent = contentVal
	return nil
}

func (fm *FileMonitor) GetContent() any {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()
	return fm.fileContent
}

func (fm *FileMonitor) MonitorFile(doneCh <-chan struct{}) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fm.logError(err, "Error creating watcher")
		return
	}
	defer watcher.Close()
	errCh := make(chan error)

	go func() {
		trackedFiles, err := fm.handleFileWatch(watcher)
		if err != nil {
			errCh <- err
			return
		}

		err = fm.updateContent()
		if err != nil {
			errCh <- err
			return
		}

		fileWatch := false
		for {
			if fileWatch {
				newTrackedFiles, err := fm.handleFileWatch(watcher)
				if err != nil {
					fm.logError(err, "Error handling file watch, pausing for a few seconds")
					time.Sleep(10 * time.Second)
					continue
				}
				trackedFiles = newTrackedFiles
				err = fm.updateContent()
				if err != nil {
					fm.logError(err, "Error updating content")
				} else {
					fm.logInfo("File content updated successfully")
				}
				fileWatch = false
			}
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					errCh <- errors.New("watcher closed event channel")
					return
				}
				if _, ok := trackedFiles[event.Name]; ok && event.Has(fsnotify.Write|fsnotify.Remove|fsnotify.Create) {
					if event.Has(fsnotify.Write) {
						fm.logInfo("File %s has been modified", event.Name)
					} else if event.Has(fsnotify.Remove) {
						fm.logInfo("File %s has been removed", event.Name)
					} else if event.Has(fsnotify.Create) {
						fm.logInfo("File %s has been created", event.Name)
					}
					fileWatch = true
				}
			case err, ok := <-watcher.Errors:
				fm.logError(err, "Unexpected watch error")
				if !ok {
					errCh <- err
					return
				}
			}
		}
	}()

	fm.logInfo("Starting to monitor %s", fm.filePath)
	select {
	case <-doneCh:
		fm.logInfo("Termination requested, stopping file watch")
	case err := <-errCh:
		fm.logError(err, "Unexpected error, stopping file watch")
	}
}

func (fm *FileMonitor) handleFileWatch(watcher *fsnotify.Watcher) (map[string]struct{}, error) {
	// Starting from the watch file, walk the links and track their locations and directories.
	// We watch modifications for all links and the final file, we do not track any intermediate directories

	trackedFiles, trackedDirs, err := walkLinks(filepath.Split(fm.filePath))
	if err != nil {
		return nil, err
	}
	currentTrackedDirs := watcher.WatchList()
	slices.Sort(currentTrackedDirs)
	if !slices.Equal(currentTrackedDirs, trackedDirs) {
		currentTrackedDirsMap := generateMap(currentTrackedDirs)
		trackedDirsMap := generateMap(trackedDirs)
		unusedTrackedDirs := removeFromMap(currentTrackedDirsMap, trackedDirsMap)
		newTrackedDirs := removeFromMap(trackedDirsMap, currentTrackedDirsMap)
		for _, newTrackedDir := range newTrackedDirs {
			fm.logInfo("Adding watch for %s", newTrackedDir)
			err = watcher.Add(newTrackedDir)
			if err != nil {
				return nil, err
			}
		}
		for _, unusedTrackedDir := range unusedTrackedDirs {
			fm.logInfo("Removing watch for %s", unusedTrackedDir)
			err = watcher.Remove(unusedTrackedDir)
			if err != nil {
				return nil, err
			}
		}
	}
	return trackedFiles, nil
}

func walkLinks(prefix, suffix string) (map[string]struct{}, []string, error) {
	prefix = filepath.Clean(prefix)
	trackedFilesMap := map[string]struct{}{}
	trackedDirsMap := map[string]struct{}{}
	suffices := split(suffix)
	for len(suffices) > 0 {
		filePath := filepath.Join(prefix, suffices[0])
		fileInfo, err := os.Lstat(filePath)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, nil, err
			}
			// file doesn't exist, we watch and stop at this point
			trackedFilesMap[filePath] = struct{}{}
			trackedDirsMap[prefix] = struct{}{}
			break
		}
		fileMode := fileInfo.Mode()
		switch {
		case fileMode&fs.ModeDir != 0:
			// Directory discovered, we move onto the next unless it's the final suffix
			if len(suffices) == 1 {
				trackedFilesMap[filePath] = struct{}{}
				trackedDirsMap[prefix] = struct{}{}
				suffices = nil
			} else {
				prefix = filePath
				suffices = suffices[1:]
			}
		case fileMode&fs.ModeSymlink != 0:
			// Symbolic link discovered, we watch this and also follow the link with the remaining suffices
			if _, ok := trackedFilesMap[filePath]; ok {
				// potential cycle discovered, we stop here
				// Note this could miss some cases
				suffices = nil
				break
			}
			trackedFilesMap[filePath] = struct{}{}
			trackedDirsMap[prefix] = struct{}{}
			link, err := os.Readlink(filePath)
			if err != nil {
				return nil, nil, err
			}
			if filepath.IsAbs(link) {
				prefix = "/"
			}
			suffices = append(split(link), suffices[1:]...)
		case fileMode&fs.ModeType == 0:
			// File discovered, either this is the target file or part of a previous link is invalid
			// Either way, we stop at this point.
			fallthrough
		default:
			// Everything else, we watch this in case it changes
			trackedFilesMap[filePath] = struct{}{}
			trackedDirsMap[prefix] = struct{}{}
			suffices = nil
		}
	}
	trackedDirs := []string{}
	for trackedDir := range trackedDirsMap {
		trackedDirs = append(trackedDirs, trackedDir)
	}
	slices.Sort(trackedDirs)
	return trackedFilesMap, trackedDirs, nil
}

func split(file string) []string {
	dir, base := filepath.Split(file)
	dir = filepath.Clean(dir)
	if dir == "." || dir == "/" {
		return []string{base}
	} else {
		return append(split(dir), base)
	}
}

func removeFromMap(sourceMap, removalsMap map[string]struct{}) []string {
	result := []string{}
	for key := range sourceMap {
		if _, ok := removalsMap[key]; !ok {
			result = append(result, key)
		}
	}
	return result
}

func generateMap(sliceValues []string) map[string]struct{} {
	result := map[string]struct{}{}
	for _, value := range sliceValues {
		result[value] = struct{}{}
	}
	return result
}
