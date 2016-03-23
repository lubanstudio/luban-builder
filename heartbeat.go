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
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/parnurzeal/gorequest"
)

func Heartbeating() {
	for {
		resp, _, errs := gorequest.New().Post(EndPoint+"/builder/heartbeat").
			Set("X-LUBAN-TOKEN", Token).
			Set("X-LUBAN-STATUS", "IDLE").End()
		if len(errs) > 0 {
			log.Errorf("Fail to heart beat: %v", errs[0])
			continue
		}
		if resp.StatusCode/100 != 2 {
			log.Errorf("Unexpected response status '%d' for heart beating.", resp.StatusCode)
			continue
		}

		time.Sleep(10 * time.Second)
	}
}
