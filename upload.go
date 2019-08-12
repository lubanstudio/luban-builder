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
	"github.com/unknwon/com"
	"github.com/parnurzeal/gorequest"
	log "gopkg.in/clog.v1"
)

func Upload() {
	log.Info("Uploading artifacts for task: %d - %s", buildInfo.Task.ID, buildInfo.ImportPath)

	defer func() {
		if status == STATUS_UPLOADING {
			status = STATUS_FAILED
		}
		log.Trace("Status changed to: %s", status)
	}()

	for _, ext := range buildInfo.PackFormats {
		resp, _, errs := gorequest.New().Post(EndPoint+"/builder/upload/artifact").
			Set("X-LUBAN-TOKEN", Token).
			Set("X-LUBAN-FORMAT", ext).
			Type("multipart").
			SendFile("./artifacts/"+com.ToStr(buildInfo.Task.ID)+"."+ext, "", "artifact").End()
		if len(errs) > 0 {
			log.Error(0, "Fail to upload artifact: %v", errs[0])
			return
		}
		if resp.StatusCode/100 != 2 {
			log.Error(0, "Unexpected response status '%d' for updating matrix info:\n%s", resp.StatusCode, resp.Body)
			return
		}
	}

	log.Info("Artifacts for task '%d' uploaded", buildInfo.Task.ID)
	status = STATUS_SUCCEED
}
