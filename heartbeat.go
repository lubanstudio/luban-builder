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
	"encoding/json"
	"time"

	"github.com/parnurzeal/gorequest"
	log "gopkg.in/clog.v1"
)

type Status string

const (
	STATUS_IDLE      Status = "IDLE"
	STATUS_BUILDING  Status = "BUILDING"
	STATUS_UPLOADING Status = "UPLOADING"
	STATUS_FAILED    Status = "FAILED"
	STATUS_SUCCEED   Status = "SUCCEED"
)

var status = STATUS_IDLE

func Heartbeating() {
	defer time.AfterFunc(15*time.Second, Heartbeating)

	resp, _, errs := gorequest.New().Post(EndPoint+"/builder/heartbeat").
		Set("X-LUBAN-TOKEN", Token).
		Set("X-LUBAN-STATUS", string(status)).End()
	if len(errs) > 0 {
		log.Error(0, "Fail to heart beat: %v", errs[0])
		return
	}

	if resp.StatusCode/100 != 2 {
		log.Error(0, "Unexpected response status '%d' for heart beating.\n%s", resp.StatusCode, resp.Body)
		return
	}

	switch status {
	case STATUS_IDLE:
		if resp.Header.Get("X-LUBAN-TASK") == "ASSIGN" {
			buildInfo = new(BuildInfo)
			if err := json.NewDecoder(resp.Body).Decode(buildInfo); err != nil {
				log.Error(0, "NewDecoder: %v", err)
				return
			}
			status = STATUS_BUILDING
			go Build()
		}

	case STATUS_FAILED, STATUS_SUCCEED:
		status = STATUS_IDLE
	}
}
