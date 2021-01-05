/*
Copyright 2021 Kohl's Department Stores, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

// example call: go test -v -args -googleAPIjsonkeypath=../../project-credential.json -googleAPIdatasetID=prometheus_test -googleAPItableID=test_stream ./...

func TestGetFiles(t *testing.T) {
	expected := []string{"testdata/default/defaults.json", "testdata/default/defaults.yml"}
	result := getFiles("testdata/default", defaultFileFilter)
	assert.Equal(t, expected, result)

	expected = []string{"testdata/yaml/one.yaml", "testdata/yaml/two.yml"}
	result = getFiles("testdata/yaml", defaultFileFilter)
	assert.Equal(t, expected, result)
}

func TestProcessHierarchy(t *testing.T) {
	var cfg config

	cfg.hierarchyFile = "testdata/test1/hierarchy.lst"
	cfg.basePath = "testdata/test1"
	cfg.outputFile = "output.yaml"
	cfg.filterExtension = defaultFileFilter
	cfg.logDebug = false
	cfg.logTrace = false
	cfg.failMissing = false

	expected := []string{"testdata/default", "testdata/yaml", "testdata/json", "testdata/empty", "testdata/test1"}
	result := processHierarchy(cfg)
	assert.Equal(t, expected, result)
}

func TestFailEmpty(t *testing.T) {
	if os.Getenv("TEST_FAIL_EMPTY") == "1" {
		var cfg config
		cfg.hierarchyFile = "testdata/test1/hierarchy.lst"
		cfg.basePath = "testdata/test1"
		cfg.outputFile = "output.yaml"
		cfg.filterExtension = defaultFileFilter
		cfg.logDebug = false
		cfg.logTrace = false
		cfg.failMissing = true

		processHierarchy(cfg)

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestFailEmpty")
	cmd.Env = append(os.Environ(), "TEST_FAIL_EMPTY=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		fmt.Printf("Process correctly failed with %v\n", e)
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestEnd2End(t *testing.T) {
	var cfg config

	cfg.hierarchyFile = "testdata/test1/hierarchy.lst"
	cfg.basePath = "testdata/test1"
	cfg.outputFile = "output.yaml"
	cfg.filterExtension = defaultFileFilter
	cfg.logDebug = false
	cfg.logTrace = false
	cfg.failMissing = false

	// process the hierarchy and get the list of include files
	hierarchy := processHierarchy(cfg)

	// Lets do the deed
	mergeYamls(hierarchy, cfg.filterExtension, cfg.outputFile)

	expected, err := ioutil.ReadFile("testdata/test1/result/expected.yaml")
	if err != nil {
		t.Fatalf("Error reading file with expected test results: %v", err)
	}
	result, err := ioutil.ReadFile(cfg.outputFile)
	if err != nil {
		t.Fatalf("Error reading output file: %v", err)
	}
	assert.Equal(t, string(expected), string(result))
}

func TestFailMissingEnvironmentVariable(t *testing.T) {
	if os.Getenv("TEST_FAIL_EMPTY") == "1" {
		var cfg config
		cfg.hierarchyFile = "testdata/test2-with-env-fail/hierarchy.lst"
		cfg.basePath = "testdata/test2-with-env-fail"
		cfg.outputFile = "output.yaml"
		cfg.filterExtension = defaultFileFilter
		cfg.logDebug = false
		cfg.logTrace = false
		cfg.failMissing = false

		processHierarchy(cfg)

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestFailMissingEnvironmentVariable")
	cmd.Env = append(os.Environ(), "TEST_FAIL_EMPTY=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		fmt.Printf("Process correctly failed with %v\n", e)
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestEnd2EndEnvironmentVariables(t *testing.T) {
	var cfg config

	cfg.hierarchyFile = "testdata/test2-with-env/hierarchy.lst"
	cfg.basePath = "testdata/test2-with-env"
	cfg.outputFile = "output.yaml"
	cfg.filterExtension = defaultFileFilter
	cfg.logDebug = false
	cfg.logTrace = false
	cfg.failMissing = false

	// set the test environment variable
	os.Setenv("JSON", "json")

	// process the hierarchy and get the list of include files
	hierarchy := processHierarchy(cfg)

	// Lets do the deed
	mergeYamls(hierarchy, cfg.filterExtension, cfg.outputFile)

	expected, err := ioutil.ReadFile("testdata/test2-with-env/result/expected.yaml")
	if err != nil {
		t.Fatalf("Error reading file with expected test results: %v", err)
	}
	result, err := ioutil.ReadFile(cfg.outputFile)
	if err != nil {
		t.Fatalf("Error reading output file: %v", err)
	}
	assert.Equal(t, string(expected), string(result))
}
