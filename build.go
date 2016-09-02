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

	log "github.com/Sirupsen/logrus"
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
	defer func() {
		if status == STATUS_BUILDING {
			log.Debug("Status changed to: STATUS_FAILED")
			status = STATUS_FAILED
		}
	}()

	os.Mkdir("log", os.ModePerm)
	logger, err := os.Create("log/" + com.ToStr(buildInfo.Task.ID) + ".output")
	if err != nil {
		log.Errorf("Fail to create log file: %v", err)
		return
	}
	defer logger.Close()

	stdout, stderr, err := com.ExecCmd("go", "get", "-u", "-v", "-tags", buildInfo.Task.Tags, buildInfo.ImportPath)
	if err != nil {
		log.Errorf("Fail to go get: %v - %s", err, stderr)
		return
	}
	fmt.Println(stdout)

	stdout, stderr, err = com.ExecCmdDir(os.Getenv("GOPATH")+"/src/"+buildInfo.ImportPath, "git", "checkout", buildInfo.Task.Commit)
	if err != nil {
		log.Errorf("Fail to git checkout: %v - %s", err, stderr)
		return
	}
	fmt.Println(stdout)

	stdout, stderr, err = com.ExecCmdDir(os.Getenv("GOPATH")+"/src/"+buildInfo.ImportPath, "go", "build", "-v", "-tags", buildInfo.Task.Tags)
	if err != nil {
		log.Errorf("Fail to go build: %v - %s", err, stderr)
		return
	}
	fmt.Println(stdout)
}
