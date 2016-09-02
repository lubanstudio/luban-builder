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

	log "github.com/Sirupsen/logrus"
	"github.com/parnurzeal/gorequest"
)

type Status int

const (
	STATUS_IDLE Status = iota
	STATUS_BUILDING
	STATUS_UPLOADING
	STATUS_FAILED
	STATUS_SUCCEED
)

var status Status

func Heartbeating() {
	defer time.AfterFunc(15*time.Second, Heartbeating)

	switch status {
	case STATUS_IDLE:
		resp, _, errs := gorequest.New().Post(EndPoint+"/builder/heartbeat").
			Set("X-LUBAN-TOKEN", Token).
			Set("X-LUBAN-STATUS", "IDLE").End()
		if len(errs) > 0 {
			log.Errorf("Fail to heart beat: %v", errs[0])
			return
		}
		if resp.StatusCode/100 != 2 {
			log.Errorf("Unexpected response status '%d' for heart beating.\n%s", resp.StatusCode, resp.Body)
			return
		}

		if resp.Header.Get("X-LUBAN-TASK") == "ASSIGN" {
			buildInfo = new(BuildInfo)
			if err := json.NewDecoder(resp.Body).Decode(buildInfo); err != nil {
				log.Errorf("NewDecoder: %v", err)
				return
			}
			status = STATUS_BUILDING
			Build()
		}

	case STATUS_FAILED:
		resp, _, errs := gorequest.New().Post(EndPoint+"/builder/heartbeat").
			Set("X-LUBAN-TOKEN", Token).
			Set("X-LUBAN-STATUS", "FAILED").End()
		if len(errs) > 0 {
			log.Errorf("Fail to heart beat: %v", errs[0])
			return
		}
		if resp.StatusCode/100 != 2 {
			log.Errorf("Unexpected response status '%d' for heart beating.\n%s", resp.StatusCode, resp.Body)
			return
		}

		status = STATUS_IDLE
	}
}
