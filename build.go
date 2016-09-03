// Copyright 2016 Unknwon
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/Unknwon/cae/tz"
	"github.com/Unknwon/cae/zip"
	"github.com/Unknwon/com"
)

type BuildInfo struct {
	ImportPath  string   `json:"import_path"`
	PackRoot    string   `json:"pack_root"`
	PackEntries []string `json:"pack_entries"`
	PackFormats []string `json:"pack_formats"`
	Task        struct {
		ID     int64  `json:"id"`
		OS     string `json:"os"`
		Arch   string `json:"arch"`
		Tags   string `json:"tags"`
		Commit string `json:"commit"`
	} `json:"task"`
}

var buildInfo *BuildInfo

func Build() {
	log.Infof("Building task: %d - %s", buildInfo.Task.ID, buildInfo.ImportPath)

	defer func() {
		if status == STATUS_BUILDING {
			status = STATUS_FAILED
		}
		log.Debugf("Status changed to: %s", status)
	}()

	os.Mkdir("log", os.ModePerm)
	logger, err := os.Create("log/" + com.ToStr(buildInfo.Task.ID) + ".output")
	if err != nil {
		log.Errorf("Fail to create log file: %v", err)
		return
	}
	defer logger.Close()

	gopath := com.GetGOPATHs()[0]
	execDir := path.Join(gopath, "src", buildInfo.ImportPath)

	// Checkout source code and compile.
	logger.WriteString("$ git checkout master\n")
	stdout, stderr, err := com.ExecCmdDirBytes(execDir, "git", "checkout", "master")
	if err != nil {
		log.Errorf("Fail to git checkout: %v - %s", err, stderr)
		return
	}
	logger.Write(stdout)
	logger.Write(stderr)

	tags := strings.Replace(buildInfo.Task.Tags, ",", " ", -1)
	logger.WriteString(fmt.Sprintf("$ go get -u -v -tags %s %s\n", tags, buildInfo.ImportPath))
	stdout, stderr, err = com.ExecCmdBytes("go", "get", "-v", "-tags", tags, buildInfo.ImportPath)
	if err != nil {
		log.Errorf("Fail to go get: %v - %s", err, stderr)
		return
	}
	logger.Write(stdout)
	logger.Write(stderr)

	logger.WriteString("$ git fetch origin\n")
	stdout, stderr, err = com.ExecCmdDirBytes(execDir, "git", "fetch", "origin")
	if err != nil {
		log.Errorf("Fail to git fetch: %v - %s", err, stderr)
		return
	}
	logger.Write(stdout)
	logger.Write(stderr)

	logger.WriteString(fmt.Sprintf("$ git checkout %s\n", buildInfo.Task.Commit))
	stdout, stderr, err = com.ExecCmdDirBytes(execDir, "git", "checkout", buildInfo.Task.Commit)
	if err != nil {
		log.Errorf("Fail to git checkout: %v - %s", err, stderr)
		return
	}
	logger.Write(stdout)
	logger.Write(stderr)

	logger.WriteString(fmt.Sprintf("$ go build -v tags %s\n", tags))
	stdout, stderr, err = com.ExecCmdDirBytes(execDir, "go", "build", "-v", "-tags", tags)
	if err != nil {
		log.Errorf("Fail to go build: %v - %s", err, stderr)
		return
	}
	logger.Write(stdout)
	logger.Write(stderr)

	// Pack artifacts.
	log.Infof("Packing artifacts: %d - %s", buildInfo.Task.ID, buildInfo.ImportPath)
	os.Mkdir("artifacts", os.ModePerm)
	artifactPath := path.Join("artifacts", com.ToStr(buildInfo.Task.ID))
	for _, ext := range buildInfo.PackFormats {
		tmpPath := artifactPath + "." + ext

		switch ext {
		case "tar.gz":
			artifact, err := tz.Create(tmpPath)
			if err != nil {
				log.Errorf("Fail to create artifact '%s': %v", tmpPath, err)
				return
			}

			for _, entry := range buildInfo.PackEntries {
				entryPath := path.Join(execDir, entry)
				if com.IsDir(entryPath) {
					artifact.AddDir(path.Join(buildInfo.PackRoot, entry), entryPath)
				} else {
					artifact.AddFile(path.Join(buildInfo.PackRoot, entry), entryPath)
				}
			}
			artifact.Close()

		case "zip":
			artifact, err := zip.Create(tmpPath)
			if err != nil {
				log.Errorf("Fail to create artifact '%s': %v", tmpPath, err)
				return
			}

			for _, entry := range buildInfo.PackEntries {
				entryPath := path.Join(execDir, entry)
				if com.IsDir(entryPath) {
					artifact.AddDir(path.Join(buildInfo.PackRoot, entry), entryPath)
				} else {
					artifact.AddFile(path.Join(buildInfo.PackRoot, entry), entryPath)
				}
			}
			artifact.Close()
		}
	}

	status = STATUS_UPLOADING
	Upload()
}
