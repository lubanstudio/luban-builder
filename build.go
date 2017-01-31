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
	glog "log"
	"os"
	"path"
	"runtime"
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
	output, err := os.Create("log/" + com.ToStr(buildInfo.Task.ID) + ".output")
	if err != nil {
		log.Errorf("Fail to create log file: %v", err)
		return
	}
	defer output.Close()
	glog.SetOutput(output)

	gopath := com.GetGOPATHs()[0]
	execDir := path.Join(gopath, "src", buildInfo.ImportPath)

	runtime.Gosched()

	// Checkout source code and compile.
	if com.IsExist(execDir) {
		glog.Println("$ git checkout master")
		stdout, stderr, err := com.ExecCmdDirBytes(execDir, "git", "checkout", "master")
		if err != nil {
			output.Write(stderr)
			log.Errorf("Fail to git checkout: %v - %s", err, stderr)
			return
		}
		output.Write(stdout)
		output.Write(stderr)

		runtime.Gosched()
	}

	tags := strings.Replace(buildInfo.Task.Tags, ",", " ", -1)
	glog.Println(fmt.Sprintf("$ go get -u -v -tags %s %s", tags, buildInfo.ImportPath))
	stdout, stderr, err := com.ExecCmdBytes("go", "get", "-u", "-v", "-tags", tags, buildInfo.ImportPath)
	if err != nil {
		output.Write(stderr)
		log.Errorf("Fail to go get: %v - %s", err, stderr)
		return
	}
	output.Write(stdout)
	output.Write(stderr)

	runtime.Gosched()

	glog.Println("$ git fetch origin")
	stdout, stderr, err = com.ExecCmdDirBytes(execDir, "git", "fetch", "origin")
	if err != nil {
		output.Write(stderr)
		log.Errorf("Fail to git fetch: %v - %s", err, stderr)
		return
	}
	output.Write(stdout)
	output.Write(stderr)

	runtime.Gosched()

	glog.Println(fmt.Sprintf("$ git checkout %s", buildInfo.Task.Commit))
	stdout, stderr, err = com.ExecCmdDirBytes(execDir, "git", "checkout", buildInfo.Task.Commit)
	if err != nil {
		output.Write(stderr)
		log.Errorf("Fail to git checkout: %v - %s", err, stderr)
		return
	}
	output.Write(stdout)
	output.Write(stderr)

	runtime.Gosched()

	glog.Println(fmt.Sprintf("$ go build -v tags %s", tags))
	stdout, stderr, err = com.ExecCmdDirBytes(execDir, "go", "build", "-v", "-tags", tags)
	if err != nil {
		output.Write(stderr)
		log.Errorf("Fail to go build: %v - %s", err, stderr)
		return
	}
	output.Write(stdout)
	output.Write(stderr)

	runtime.Gosched()

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
				output.WriteString(err.Error())
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
				output.WriteString(err.Error())
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
