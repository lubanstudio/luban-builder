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
	"io/ioutil"
	"runtime"

	"github.com/unknwon/com"
	"github.com/parnurzeal/gorequest"
	log "gopkg.in/clog.v1"
)

const APP_VER = "0.1.7.0208"

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	log.Info("Luban Builder %s", APP_VER)
	log.Info("GOMAXPROCS: %d", runtime.NumCPU())

	if len(EndPoint) == 0 {
		fmt.Print("Please enter the server end point: ")
		EndPoint = AskLine()
		SaveEndPoint()
	}

	if len(Token) == 0 {
		fmt.Print("Please enter the token given by the admin: ")
		Token = AskLine()
		SaveToken()
	}

	matricesFile := "matrices.json"
	if !com.IsFile(matricesFile) {
		log.Fatal(0, "File '%s' not found, please define it first.", matricesFile)
	}

	var err error
	MatricesData, err = ioutil.ReadFile(matricesFile)
	if err != nil {
		log.Fatal(0, "Fail to load '%s': %v", matricesFile, err)
	}

	resp, _, errs := gorequest.New().Post(EndPoint+"/builder/matrix").
		Set("X-LUBAN-TOKEN", Token).
		SendString(string(MatricesData)).End()
	if len(errs) > 0 {
		log.Fatal(0, "Fail to update matrix info: %v", errs[0])
	}
	if resp.StatusCode/100 != 2 {
		log.Fatal(0, "Unexpected response status '%d' for updating matrix info:\n%s", resp.StatusCode, resp.Body)
	}

	log.Info("All going well, start heart beating...")
	go Heartbeating()

	select {}
}
