//  Copyright 2020 Marius Ackerman
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

/*
Package ioutil contains functions for writing directories and files.
*/
package ioutil

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// FilePermission given to all created files and directories
const FilePermission = 0731

// Exist returns true if path exists, otherwise false.
func Exist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// MkdirAll makes all the directories in path.
func MkdirAll(path string) error {
	if path == "" {
		return nil
	}
	return os.MkdirAll(path, FilePermission)
}

// WriteFile creates all the non-existend directories in path before writing
// data to path.
func WriteFile(path string, data []byte) error {
	dir, _ := filepath.Split(path)
	if err := MkdirAll(dir); err != nil {
		return fmt.Errorf("Error creating directory %s: %s", dir, err)
	}
	if err := ioutil.WriteFile(path, data, FilePermission); err != nil {
		return fmt.Errorf("Error writing file %s: %s\n", path, err)
	}
	return nil
}
