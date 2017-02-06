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
	"github.com/Unknwon/com"
	log "gopkg.in/clog.v1"
	"gopkg.in/ini.v1"
)

var (
	EndPoint string
	Token    string

	MatricesData []byte

	ConfFile = ".luban.ini"
	Cfg      *ini.File
)

func init() {
	if err := log.NewLogger(log.CONSOLE, log.ConsoleConfig{log.TRACE, 100}); err != nil {
		panic(err.Error())
	}

	if !com.IsFile(ConfFile) {
		Cfg = ini.Empty()
		return
	}

	var err error
	Cfg, err = ini.Load(ConfFile)
	if err != nil {
		log.Fatal(0, "Fail to load '%s': %v", ConfFile, err)
	}

	sec := Cfg.Section("")
	EndPoint = sec.Key("END_POINT").String()
	Token = sec.Key("TOKEN").String()
}

func SaveSettings() {
	if err := Cfg.SaveTo(ConfFile); err != nil {
		log.Fatal(0, "Fail to save '%s': %v", ConfFile, err)
	}
}

func SaveEndPoint() {
	Cfg.Section("").Key("END_POINT").SetValue(EndPoint)
	SaveSettings()
}

func SaveToken() {
	Cfg.Section("").Key("TOKEN").SetValue(Token)
	SaveSettings()
}
