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

// The main package for the executable
package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/imdario/mergo"
	"gopkg.in/yaml.v3"

	"github.com/KohlsTechnology/hierarchy/pkg/version"
	"github.com/kylelemons/godebug/diff"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

type config struct {
	hierarchyFile   string
	basePath        string
	outputFile      string
	filterExtension string
	printVersion    bool
	diffOutput      bool
	logDebug        bool
	logTrace        bool
	failMissing     bool
}

// Default file filter
const defaultFileFilter = "(.yaml|.yml|.json)$"

func parseFlags() config {
	application := kingpin.New(filepath.Base(os.Args[0]), "Hierarchy")
	application.HelpFlag.Short('h')

	cfg := config{}

	application.Flag("file", "Path and name of the hierarchy file.").Short('f').
		Envar("HIERARCHY_FILE").Default("./hierarchy.lst").StringVar(&cfg.hierarchyFile)
	application.Flag("output", "Path and name of the output file.").Short('o').
		Envar("HIERARCHY_OUTPUT").Default("./output.yaml").StringVar(&cfg.outputFile)
	application.Flag("filter", "Regex for allowed file extension(s) of files being merged.").Short('i').
		Envar("HIERARCHY_FILTER").Default(defaultFileFilter).StringVar(&cfg.filterExtension)
	application.Flag("failmissing", "Fail if a directory in the hierarchy is missing.").Short('m').
		Envar("HIERARCHY_FAILMISSING").Default("false").BoolVar(&cfg.failMissing)
	application.Flag("debug", "Print debug output.").Short('d').
		Envar("HIERARCHY_DEBUG").Default("false").BoolVar(&cfg.logDebug)
	application.Flag("trace", "Prints a diff after processing each file. This generates A LOT of output.").
		Envar("HIERARCHY_TRACE").Default("false").BoolVar(&cfg.logTrace)
	application.Flag("version", "Print version and build information, then exit.").Short('V').
		Default("false").BoolVar(&cfg.printVersion)

	_, err := application.Parse(os.Args[1:])

	if cfg.printVersion {
		version.Print()
		os.Exit(0)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "Error parsing commandline arguments"))
		application.Usage(os.Args[1:])
		os.Exit(2)
	}
	return cfg
}

// checkForError fails the program with a fatal error message if e != nil
func checkForError(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

// processHierarchy loads the hierarchy file and generates a file paths list
// of folders to be processed
func processHierarchy(cfg config) []string {
	hierarchy := []string{}
	hierarchyFile, err := os.Open(cfg.hierarchyFile)
	checkForError(err)
	defer hierarchyFile.Close()

	// Start reading from the file with a reader.
	reader := bufio.NewReader(hierarchyFile)
	var line string
	for {
		line, err = reader.ReadString('\n')
		if err != nil && err != io.EOF {
			break
		}

		// Trim spaces and comments
		includePath := strings.Split(line, "#")[0]
		includePath = strings.TrimSpace(includePath)
		includePath = replaceEnvironmentVariables(includePath)
		// Process path
		if len(includePath) > 0 {
			includePath = path.Join(cfg.basePath, includePath)
			// Check if directory exists
			if stat, err := os.Stat(includePath); err == nil && stat.IsDir() {
				hierarchy = append(hierarchy, includePath)
				absPath, _ := filepath.Abs(includePath)
				log.WithFields(log.Fields{
					"path":     includePath,
					"abs_path": absPath,
				}).Debug("Adding path to hierarchy")
			} else {
				if cfg.failMissing {
					log.WithFields(log.Fields{
						"path": includePath,
					}).Fatal("Hierarchy directory not found")
				} else {
					log.WithFields(log.Fields{
						"path": includePath,
					}).Info("Ignoring missing hierarchy directory")
				}
			}
		}

		// Break the for loop on error including EOF
		if err != nil {
			break
		}
	}
	if err != io.EOF {
		checkForError(err)
	}
	return hierarchy
}

// mergeYamls walks through all the folders in the hierarchy
// and merges all files matching the pattern into the structure
// overwriting any existing values
// and finally exports it as a new YAML file
func mergeYamls(hierarchy []string, fileFilter string, outputFile string) {
	// Initialize variables
	var data map[string]interface{}
	counter := 0

	for _, includePath := range hierarchy {
		log.WithFields(log.Fields{
			"path": includePath,
		}).Debug("Inspecting folder")

		// Merge in every file matching the pattern
		for _, file := range getFiles(includePath, fileFilter) {
			// Generate an old version of YAML for comparison
			oldYaml, err := yaml.Marshal(&data)
			checkForError(err)

			// Import the next file
			log.WithFields(log.Fields{
				"path": file,
			}).Info("Importing file")
			mergeFile, err := ioutil.ReadFile(file)
			checkForError(err)
			mergeData := make(map[string]interface{})
			err = yaml.Unmarshal([]byte(mergeFile), &mergeData)
			checkForError(err)

			err = mergo.Merge(&data, mergeData, mergo.WithOverride)
			checkForError(err)

			// Generate the new YAML and print the unified diff to the trace output
			newYaml, err := yaml.Marshal(&data)
			checkForError(err)
			log.Trace(diff.Diff(string(oldYaml), string(newYaml)))

			counter++
		}
	}

	log.WithFields(log.Fields{
		"count": counter,
	}).Info("Completed merging all files")

	// Writing the output file
	log.WithFields(log.Fields{
		"path": outputFile,
	}).Info("Writing output file")
	yamlDoc, err := yaml.Marshal(&data)
	err = ioutil.WriteFile(outputFile, yamlDoc, 0660)

	checkForError(err)
}

// getFiles lists all files in a given path and returns a list of matching the fileFilter
func getFiles(includePath string, fileFilter string) []string {
	var includeFiles []string
	files, err := ioutil.ReadDir(includePath)
	checkForError(err)
	for _, fileInfo := range files {
		if !fileInfo.IsDir() {
			filePath := path.Join(includePath, fileInfo.Name())
			r, err := regexp.MatchString(fileFilter, fileInfo.Name())
			if err == nil && r {
				includeFiles = append(includeFiles, filePath)
				log.WithFields(log.Fields{
					"file": filePath,
				}).Debug("Adding file to list")
			} else {
				log.WithFields(log.Fields{
					"file": filePath,
				}).Debug("Ignoring file")
			}
		}
	}

	return includeFiles
}

// ReplaceEnvironmentVariables replaces all variable names in a string with the content defined on the OS
// If a variable is not defined, it will fail to avoid any unintended results
func replaceEnvironmentVariables(str string) string {
	// variables must be in the format ${NAME}
	// Letters, numbers, and underscores are allowed
	// Variable name must start with a letter
	// Environment variable names will be converted to upper case to avoid ambiguity
	re := regexp.MustCompile(`\$\{[A-Za-z][][A-Za-z_0-9.]*\}`)
	for _, varName := range re.FindAllString(str, -1) {
		envVarName := strings.TrimPrefix(varName, "${")
		envVarName = strings.TrimSuffix(envVarName, "}")
		envVar := os.Getenv(strings.ToUpper(envVarName))
		if len(envVar) == 0 {
			log.WithFields(log.Fields{
				"name": envVarName,
			}).Fatal("Environment variable not defined")
		}
		str = strings.ReplaceAll(str, varName, envVar)
	}
	return str
}

func main() {
	cfg := parseFlags()

	// Configure different logging levels
	log.SetOutput(os.Stdout)
	if cfg.logTrace {
		log.SetLevel(log.TraceLevel)
	} else if cfg.logDebug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	version.Log()

	cfg.basePath = path.Dir(cfg.hierarchyFile)

	log.WithFields(log.Fields{
		"hierarchyFile":     cfg.hierarchyFile,
		"basePath":          cfg.basePath,
		"outputFile":        cfg.outputFile,
		"outputPermissions": cfg.outputFile,
		"filterExtension":   cfg.filterExtension,
		"failMissing":       cfg.failMissing,
	}).Debug("Configuration settings")

	// Make sure we remove the output file if it already exists
	// Just in case the program ends for any reason other than success
	// We don't want to give the impression that we completed the merging
	if _, err := os.Stat(cfg.outputFile); err == nil {
		log.WithFields(log.Fields{
			"path": cfg.outputFile,
		}).Info("Removing existing output file")
		err := os.Remove(cfg.outputFile)
		checkForError(err)
	}

	// Process the hierarchy and get the list of files to be included
	hierarchy := processHierarchy(cfg)

	// Proceed with merging YAML files
	mergeYamls(hierarchy, cfg.filterExtension, cfg.outputFile)
}
