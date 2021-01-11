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

var cfgDefaults = config{
	hierarchyFile:        "hierarchy.lst",
	basePath:             "",
	outputFile:           "output.yaml",
	filterExtension:      defaultFileFilter,
	printVersion:         false,
	diffOutput:           false,
	logDebug:             false,
	logTrace:             false,
	failMissingHierarchy: false,
	failMissingPath:      false,
	failMissingEnvVar:    false,
	skipEnvVarContent:    false,
}

// TestGetFilesSuccess verifies that we receive the correct list of files to be merged
// As part of this test, it will also check the proper functioning of the regex for the file filter
// fail.txt and fail.yaml.disabled should never be returned
func TestGetFilesSuccess(t *testing.T) {
	expected := []string{"testdata/default/defaults.json", "testdata/default/defaults.yml"}
	result := getFiles("testdata/default", defaultFileFilter)
	assert.Equal(t, expected, result)

	expected = []string{"testdata/yaml/one.yaml", "testdata/yaml/two.yml"}
	result = getFiles("testdata/yaml", defaultFileFilter)
	assert.Equal(t, expected, result)
}

// TestProcessHierarchySuccess verifies that all directories are correctly added to the list to be processed
// It will also test the correct handling of comments and different ways of specifying a relative path
func TestProcessHierarchySuccess(t *testing.T) {
	cfg := cfgDefaults
	cfg.basePath = "testdata/test1"

	expected := []string{"testdata/default", "testdata/yaml", "testdata/json", "testdata/empty", "testdata/test1"}
	result := processHierarchy(cfg)
	assert.Equal(t, expected, result)
}

// TestFailMissingPath tests the correct behavior of the `--failmissing` command line option
// It spawns a new process to determine the exit code of the application.
// Anything other than a 1 is a problem
// It uses the environment variable TEST_FAIL_EMPTY to signal the actual execution of the functionality
func TestFailMissingPath(t *testing.T) {
	if os.Getenv("TEST_FAIL_EMPTY") == "1" {
		cfg := cfgDefaults
		cfg.basePath = "testdata/test1"
		cfg.failMissingPath = true

		processHierarchy(cfg)

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestFailMissingPath")
	cmd.Env = append(os.Environ(), "TEST_FAIL_EMPTY=1")
	output, err := cmd.CombinedOutput()
	fmt.Printf("%s\n", output)
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		fmt.Printf("Process correctly failed with %v\n", e)
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1.", err)
}

// TestEnd2EndSuccess runs through the full functionality end-to-end
// It compares the generated final file with one stored in git
func TestEnd2EndSuccess(t *testing.T) {

	cfg := cfgDefaults
	cfg.basePath = "testdata/test1"

	// process the hierarchy and get the list of include files
	hierarchy := processHierarchy(cfg)

	// Lets do the deed
	mergeFilesInHierarchy(hierarchy, cfg.filterExtension, cfg.outputFile, false, false)

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

// TestFailHierarchyMissingEnvironmentVariable ensures that the application is correctly failing
// If an environment variable specified in `hierarchy.lst` is not found.
// It spawns a new process to determine the exit code of the application.
// Anything other than a 1 is a problem
// It uses the environment variable TEST_FAIL_EMPTY to signal the actual execution of the functionality
// Hierarchy will always fail if an environment variable as part of the hierarchy is not found.
// This is intentional, to prevent unexpected behaviors.
func TestFailHierarchyMissingEnvironmentVariable(t *testing.T) {
	if os.Getenv("TEST_FAIL_EMPTY") == "1" {
		cfg := cfgDefaults
		cfg.basePath = "testdata/hierarchy-with-env-fail"

		processHierarchy(cfg)

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestFailHierarchyMissingEnvironmentVariable")
	cmd.Env = append(os.Environ(), "TEST_FAIL_EMPTY=1")
	output, err := cmd.CombinedOutput()
	fmt.Printf("%s\n", output)
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		fmt.Printf("Process correctly failed with %v\n", e)
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1.", err)
}

// TestEnd2EndHierarchyEnvironmentVariablesSuccess runs through the full functionality end-to-end,
// specifically testing the correct resolution of environment variables specified in `hierarchy.lst`
// It compares the generated final file with one stored in git
func TestEnd2EndHierarchyEnvironmentVariablesSuccess(t *testing.T) {
	cfg := cfgDefaults
	cfg.basePath = "testdata/hierarchy-with-env"
	cfg.outputFile = "output.yaml"

	// set the test environment variable
	os.Setenv("JSON", "json")

	// process the hierarchy and get the list of include files
	hierarchy := processHierarchy(cfg)

	// Merge files
	mergeFilesInHierarchy(hierarchy, cfg.filterExtension, cfg.outputFile, false, false)

	expected, err := ioutil.ReadFile("testdata/hierarchy-with-env/result/expected.yaml")
	if err != nil {
		t.Fatalf("Error reading file with expected test results: %v", err)
	}
	result, err := ioutil.ReadFile(cfg.outputFile)
	if err != nil {
		t.Fatalf("Error reading output file: %v", err)
	}
	assert.Equal(t, string(expected), string(result))
}

// TestMissingHierarchySuccess runs through the full functionality end-to-end
// It test the case where no hierarchy.lst is provided,
// meaning only the base directory is searched for files to merge
// It compares the generated final file with one stored in git
func TestMissingHierarchySuccess(t *testing.T) {

	cfg := cfgDefaults
	cfg.basePath = "testdata/no-hierarchy"

	// process the hierarchy and get the list of include files
	hierarchy := processHierarchy(cfg)

	// Lets do the deed
	mergeFilesInHierarchy(hierarchy, cfg.filterExtension, cfg.outputFile, false, false)

	expected, err := ioutil.ReadFile("testdata/no-hierarchy/result/expected.yaml")
	if err != nil {
		t.Fatalf("Error reading file with expected test results: %v", err)
	}
	result, err := ioutil.ReadFile(cfg.outputFile)
	if err != nil {
		t.Fatalf("Error reading output file: %v", err)
	}
	assert.Equal(t, string(expected), string(result))
}

// TestFailMissingHierarchy runs through the full functionality end-to-end
// It test the case where no hierarchy.lst is provided,
// meaning only the base directory is searched for files to merge
// It compares the generated final file with one stored in git
func TestFailMissingHierarchy(t *testing.T) {
	if os.Getenv("TEST_FAIL_MISSING_HIERARCHY") == "1" {
		cfg := cfgDefaults
		cfg.basePath = "testdata/no-hierarchy"
		cfg.failMissingHierarchy = true

		processHierarchy(cfg)

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestFailMissingHierarchy")
	cmd.Env = append(os.Environ(), "TEST_FAIL_MISSING_HIERARCHY=1")
	output, err := cmd.CombinedOutput()
	fmt.Printf("%s\n", output)
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		fmt.Printf("Process correctly failed with %v\n", e)
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1.", err)

}

// TestFailContentMissingEnvironmentVariable ensures that the application is correctly failing
// If an environment variable specified in the yaml content is not defined.
// It spawns a new process to determine the exit code of the application.
// Anything other than a 1 is a problem
// It uses the environment variable TEST_FAIL_EMPTY to signal the actual execution of the functionality
func TestFailContentMissingEnvironmentVariable(t *testing.T) {
	if os.Getenv("TEST_FAIL_EMPTY") == "1" {
		cfg := cfgDefaults
		cfg.basePath = "testdata/content-with-env"
		cfg.failMissingEnvVar = true

		// process the hierarchy and get the list of include files
		hierarchy := processHierarchy(cfg)

		// Merge files in hierarchy
		mergeFilesInHierarchy(hierarchy, cfg.filterExtension, cfg.outputFile, cfg.skipEnvVarContent, cfg.failMissingEnvVar)

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestFailContentMissingEnvironmentVariable")
	cmd.Env = append(os.Environ(), "TEST_FAIL_EMPTY=1", "EXISTING_VARIABLE1=one", "EXISTING_VARIABLE2=two")
	output, err := cmd.CombinedOutput()
	fmt.Printf("%s\n", output)
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		fmt.Printf("Process correctly failed with %v\n", e)
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1.", err)
}

// TestContentMissingEnvironmentVariableSuccess runs through the full functionality end-to-end
// It test the case where no hierarchy.lst is provided,
// meaning only the base directory is searched for files to merge
// It compares the generated final file with one stored in git
func TestContentMissingEnvironmentVariableSuccess(t *testing.T) {
	cfg := cfgDefaults
	cfg.basePath = "testdata/content-with-env/"

	// process the hierarchy and get the list of include files
	hierarchy := processHierarchy(cfg)

	// set the test environment variables
	os.Setenv("EXISTING_VARIABLE1", "one")
	os.Setenv("EXISTING_VARIABLE2", "two")

	// merge files in hierarchy
	mergeFilesInHierarchy(hierarchy, cfg.filterExtension, cfg.outputFile, cfg.skipEnvVarContent, cfg.failMissingEnvVar)

	expected, err := ioutil.ReadFile("testdata/content-with-env/result/expected.yaml")
	if err != nil {
		t.Fatalf("Error reading file with expected test results: %v", err)
	}
	result, err := ioutil.ReadFile(cfg.outputFile)
	if err != nil {
		t.Fatalf("Error reading output file: %v", err)
	}
	assert.Equal(t, string(expected), string(result))
}
