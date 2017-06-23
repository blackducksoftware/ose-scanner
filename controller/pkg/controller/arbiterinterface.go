/*
Copyright (C) 2016 Black Duck Software, Inc.
http://www.blackducksoftware.com/

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownership. The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied. See the License for the
specific language governing permissions and limitations
under the License.
*/

package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"net/http"
)

type imageInfo struct {
	ControllerID string `json:"id,omitempty"`
	ImageSpec    string `json:"spec,omitempty"`
}

type imageResult struct {
	RequestId string `json:"requestId"`
	StartScan bool   `json:"startScan"`
	SkipScan  bool   `json:"skipScan"`
}

type controllerInfo struct {
	Id          string `json:"id"`
	WorkerCount int    `json:"workers"`
}

type Arbiter struct {
	baseUrl      string
	workerCount  int
	controllerId string
	connected    bool
}

func NewArbiter(baseUrl string, workerCount int, controllerId string) *Arbiter {
	return &Arbiter{
		controllerId: controllerId,
		workerCount:  workerCount,
		baseUrl:      baseUrl,
		connected:    false,
	}
}

func (a *Arbiter) heartbeat() bool {
	var ci controllerInfo

	ci.Id = a.controllerId
	ci.WorkerCount = a.workerCount

	b, err := json.Marshal(ci)
	if err != nil {
		log.Printf("Marshalling heartbeat error %s\n", err)
		return false
	}

	url := a.baseUrl + "/heartbeat"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("HTTP client new image error %s\n", err)
		return false
	}

	resp.Body.Close()

	a.connected = true

	return true
}

func (a *Arbiter) alertImage(spec string) (requestHash string, skipScan bool) {

	log.Printf("Notifying arbiter for image %s\n", spec)

	var i imageInfo

	i.ControllerID = a.controllerId
	i.ImageSpec = spec

	b, err := json.Marshal(i)
	if err != nil {
		log.Printf("Marshalling new image error %s\n", err)
		return "", true
	}

	url := a.baseUrl + "/image/found"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("HTTP client new image error %s\n", err)
		return "", true
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusTooManyRequests:
		log.Printf("Image %s previously requested and in queue\n", spec)
		return "", true

	case http.StatusCreated:
		var result imageResult
		_ = json.NewDecoder(resp.Body).Decode(&result)

		return result.RequestId, false

	default:
		log.Printf("New image http status error %d\n", resp.StatusCode)
		return "", true
	}

}

func (a *Arbiter) requestImage(spec string) (requestHash string, skipScan bool, startScan bool) {

	log.Printf("Requesting arbiter authorization for image %s\n", spec)
	var i imageInfo

	i.ControllerID = a.controllerId
	i.ImageSpec = spec

	b, err := json.Marshal(i)
	if err != nil {
		log.Printf("Marshalling new image error %s\n", err)
		return "", true, false
	}

	url := a.baseUrl + "/image/request"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("HTTP client new image error %s\n", err)
		return "", true, false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("New image http status error %d\n", resp.StatusCode)
		return "", true, false
	}

	var result imageResult
	_ = json.NewDecoder(resp.Body).Decode(&result)

	return result.RequestId, result.SkipScan, result.StartScan

}

func (a *Arbiter) scanDone(requestHash string) {

	log.Printf("Notifying arbiter of completed scan for image hash %s\n", requestHash)
	var ci controllerInfo

	ci.Id = a.controllerId
	ci.WorkerCount = a.workerCount

	b, err := json.Marshal(ci)
	if err != nil {
		log.Printf("Marshalling scan complete error %s\n", err)
		return
	}

	url := fmt.Sprintf("%s/image/%s/done", a.baseUrl, requestHash)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("HTTP client scan complete error %s\n", err)
		return
	}
	defer resp.Body.Close()

	return

}

func (a *Arbiter) abortScan(requestHash string) {

	log.Printf("Notifying arbiter of aborted scan for image %s\n", requestHash)

	var ci controllerInfo

	ci.Id = a.controllerId
	ci.WorkerCount = a.workerCount

	b, err := json.Marshal(ci)
	if err != nil {
		log.Printf("Marshalling abort scan error %s\n", err)
		return
	}

	url := fmt.Sprintf("%s/image/%s/abort", a.baseUrl, requestHash)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("HTTP client abort scan error %s\n", err)
		return
	}
	defer resp.Body.Close()

	return

}
