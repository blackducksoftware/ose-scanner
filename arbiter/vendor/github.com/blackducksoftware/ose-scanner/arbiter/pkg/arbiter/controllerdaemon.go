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

package arbiter

import (
	"log"
	"sync"
	"time"
)

type controllerInfo struct {
	Id          string `json:"id"`
	WorkerCount int    `json:"workers"`
}

type controllerDaemon struct {
	info           controllerInfo
	lastMessage    time.Time
	requestedScans map[string]string
	assignedScans  map[string]time.Time
	sync.RWMutex
}

func newControllerDaemon(ID string, workerCount int) *controllerDaemon {

	return &controllerDaemon{
		info: controllerInfo{
			Id:          ID,
			WorkerCount: workerCount,
		},
		lastMessage:    time.Now(),
		requestedScans: make(map[string]string),
		assignedScans:  make(map[string]time.Time),
	}
}

func (cd *controllerDaemon) Heartbeat() {
	cd.lastMessage = time.Now()
}

func (cd *controllerDaemon) AssignScan(imageSpec string) bool {
	cd.Lock()
	defer cd.Unlock()

	cd.Heartbeat()

	_, ok := cd.requestedScans[imageSpec]
	if !ok {
		log.Printf("Unable to assign scan to controller %s. %s doesn't exist for this controller\n", cd.info.Id, imageSpec)
		return false
	}

	_, ok = cd.assignedScans[imageSpec]
	if ok {
		log.Printf("Previously assigned scan %s to controller %s. Claiming again.\n", imageSpec, cd.info.Id)
		return true
	}

	if cd.info.WorkerCount > 0 {
		cd.assignedScans[imageSpec] = time.Now()
		cd.info.WorkerCount--
		log.Printf("Assigning scan for %s to controller %s. Free workers now at %d \n", imageSpec, cd.info.Id, cd.info.WorkerCount)
		return true
	} else {
		log.Printf("Unable to assign scan for %s to controller %s at this time. No free workers\n", imageSpec, cd.info.Id)
		return false
	}
}

func (cd *controllerDaemon) CompleteScan(imageSpec string) {
	cd.Lock()
	defer cd.Unlock()

	delete(cd.requestedScans, imageSpec)

	_, ok := cd.assignedScans[imageSpec]

	if ok {
		delete(cd.assignedScans, imageSpec)
		cd.info.WorkerCount++
		log.Printf("Completed scan of %s on controller %s. Free workers now at %d \n", imageSpec, cd.info.Id, cd.info.WorkerCount)
	} else {
		log.Printf("Request to complete scan of %s, but scan not assigned to controller %s\n", imageSpec, cd.info.Id)
	}

}

func (cd *controllerDaemon) AbortScan(imageSpec string) {
	cd.Lock()
	defer cd.Unlock()

	delete(cd.requestedScans, imageSpec)

	_, ok := cd.assignedScans[imageSpec]

	if ok {
		delete(cd.assignedScans, imageSpec)
		cd.info.WorkerCount++
		log.Printf("Aborted scan of %s on controller %s. Free workers now at %d \n", imageSpec, cd.info.Id, cd.info.WorkerCount)
	} else {
		log.Printf("Request to abort scan of %s, but scan not assigned to controller %s\n", imageSpec, cd.info.Id)
	}

}

func (cd *controllerDaemon) SkipScan(imageSpec string) {
	cd.Lock()
	defer cd.Unlock()
	log.Printf("Skipping scan of %s in controller %s\n", imageSpec, cd.info.Id)

	delete(cd.requestedScans, imageSpec)
}

func (cd *controllerDaemon) HasFreeWorkers() bool {
	return (cd.info.WorkerCount > 0)
}

func (cd *controllerDaemon) AddScanRequest(imageSpec string, md5hash string) {
	cd.Lock()
	defer cd.Unlock()

	cd.Heartbeat()
	cd.requestedScans[imageSpec] = md5hash
}
