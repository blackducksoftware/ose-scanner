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
	"strings"
	"sync/atomic"
	"time"

	"encoding/hex"

	"crypto/md5"

	"encoding/json"

	"net/http"

	"github.com/gorilla/mux"
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

type assignImage struct {
	ControllerID string
	ImageSpec    string
	AssignTime   time.Time
	UpdateTime   time.Time
}

type jsonErr struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

func (arb *Arbiter) QueueHubActivity() uint64 {
	return atomic.AddUint64(&arb.hubActivities, 1)
}

func (arb *Arbiter) DeQueueHubActivity() {
	atomic.AddUint64(&arb.hubActivities, ^uint64(0))
}

func (arb *Arbiter) HubActivityQueue() uint64 {
	return atomic.LoadUint64(&arb.hubActivities)
}

func (arb *Arbiter) CanSendHubJobs() bool {
	// this is hardcoded for the moment, but should be a function of Hub job runners
	return arb.HubActivityQueue() > 7
}

func (arb *Arbiter) ListenForControllers() {

	log.Println("Starting router...")
	router := mux.NewRouter()
	router.HandleFunc("/heartbeat", arb.registerControllerAlive).Methods("POST")
	router.HandleFunc("/image/found", arb.foundImage).Methods("POST")
	router.HandleFunc("/image/request", arb.assignScan).Methods("POST")
	router.HandleFunc("/image/{id}/processing", arb.processingImage).Methods("POST")
	router.HandleFunc("/image/{id}/done", arb.scanDone).Methods("POST")
	router.HandleFunc("/image/{id}/abort", arb.scanAbort).Methods("POST")

	go http.ListenAndServe(":9035", router)

	log.Println("Listening for controller traffic on port 9035")

}

func (arb *Arbiter) scanAbort(w http.ResponseWriter, r *http.Request) {

	log.Println("Request scanAbort")
	params := mux.Vars(r)

	imageHash := params["id"]

	var ci controllerInfo
	_ = json.NewDecoder(r.Body).Decode(&ci)

	cd, ok := arb.controllerDaemons[ci.Id]
	if !ok {
		log.Printf("Unknown controller [%s] claimed abort for image: %s\n", ci.Id, imageHash)
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(jsonErr{Code: http.StatusNotFound, Text: "Not Found"})
		return
	}

	image, ok := arb.assignedImages[imageHash]
	if !ok {
		log.Printf("Controller [%s] claimed abort on unknown image: %s\n", ci.Id, imageHash)
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(jsonErr{Code: http.StatusNotFound, Text: "Not Found"})
		return
	}

	arb.releaseScanForPeer(image, cd, imageHash)

	w.WriteHeader(http.StatusOK)
	arb.DeQueueHubActivity()

	log.Printf("Done abortScan - Hub jobs: %d", arb.HubActivityQueue())

}

func (arb *Arbiter) releaseScanForPeer(image *assignImage, cd *controllerDaemon, imageHash string) {

	arb.Lock()
	defer arb.Unlock()

	cd.AbortScan(image.ImageSpec)
	delete(arb.assignedImages, imageHash)
}

func (arb *Arbiter) scanDone(w http.ResponseWriter, r *http.Request) {

	log.Println("Request scanDone")
	params := mux.Vars(r)

	imageHash := params["id"]

	var ci controllerInfo
	_ = json.NewDecoder(r.Body).Decode(&ci)

	cd, ok := arb.controllerDaemons[ci.Id]
	if !ok {
		log.Printf("Unknown controller [%s] claimed done for image: %s\n", ci.Id, imageHash)
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(jsonErr{Code: http.StatusNotFound, Text: "Not Found"})
		return
	}

	image, ok := arb.assignedImages[imageHash]
	if !ok {
		log.Printf("Controller [%s] claimed done on unknown image: %s\n", ci.Id, imageHash)
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(jsonErr{Code: http.StatusNotFound, Text: "Not Found"})
		return
	}

	arb.finalizeScan(image, cd, imageHash)

	w.WriteHeader(http.StatusOK)
	arb.DeQueueHubActivity()

	log.Printf("Done scanDone - Hub jobs: %d", arb.HubActivityQueue())

}

func (arb *Arbiter) finalizeScan(image *assignImage, cd *controllerDaemon, imageHash string) {

	arb.Lock()
	defer arb.Unlock()

	arb.setStatus(true, image.ImageSpec)

	cd.CompleteScan(image.ImageSpec)

	for _, peer := range arb.controllerDaemons {
		if cd.info.Id == peer.info.Id {
			// don't mess with the actual scanner or we could spin lock
			continue
		}

		peer.SkipScan(image.ImageSpec)
	}

	delete(arb.requestedImages, image.ImageSpec)
	delete(arb.assignedImages, imageHash)
}

func (arb *Arbiter) processingImage(w http.ResponseWriter, r *http.Request) {
	log.Println("Request processingImage")
	params := mux.Vars(r)

	imageHash := params["id"]

	var ci controllerInfo
	_ = json.NewDecoder(r.Body).Decode(&ci)

	_, ok := arb.controllerDaemons[ci.Id]
	if !ok {
		log.Printf("Unknown controller [%s] claimed processing image: %s\n", ci.Id, imageHash)
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(jsonErr{Code: http.StatusNotFound, Text: "Not Found"})
		return
	}

	image, ok := arb.assignedImages[imageHash]
	if !ok {
		log.Printf("Controller [%s] claimed processing unknown image: %s\n", ci.Id, imageHash)
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(jsonErr{Code: http.StatusNotFound, Text: "Not Found"})
		return
	}

	image.UpdateTime = time.Now()

	w.WriteHeader(http.StatusOK)

	log.Println("Done processingImage")

}

func (arb *Arbiter) assignScan(w http.ResponseWriter, r *http.Request) {
	log.Println("Request assignScan")
	var i imageInfo
	var resp imageResult

	if !arb.CanSendHubJobs() {
		resp.RequestId = "" 
		resp.StartScan = false 
		resp.SkipScan = false
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("Error encoding busy image response: %s\n", err)
			w.WriteHeader(500)
			return
		}

		w.WriteHeader(http.StatusOK)

		log.Println("Done assignScan - Hub currently busy with other jobs. Requeue.")
	}

	_ = json.NewDecoder(r.Body).Decode(&i)

	if len(i.ControllerID) == 0 || len(i.ImageSpec) == 0 {
		log.Printf("Got junk on assignScan API: %s\n", r.Body)
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(jsonErr{Code: http.StatusNotFound, Text: "Not Found"})
		return
	}

	cd, ok := arb.controllerDaemons[i.ControllerID]
	if !ok {
		log.Printf("Unknown controller [%s] requested image: %s\n", i.ControllerID, i.ImageSpec)
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(jsonErr{Code: http.StatusNotFound, Text: "Not Found"})
		return
	}

	resp.RequestId, resp.StartScan, resp.SkipScan = arb.findWorker(i.ImageSpec, cd)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error encoding image response: %s\n", err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(http.StatusOK)
	arb.QueueHubActivity()
	log.Printf("Done assignScan - Hub jobs: %d", arb.HubActivityQueue())
}

func (arb *Arbiter) findWorker(spec string, cd *controllerDaemon) (string, bool, bool) {
	arb.Lock()
	defer arb.Unlock()

	image, ok := arb.images[spec]
	if !ok {
		log.Printf("Missing spec. Controller %s is unable to scan %s at this time.\n", cd.info.Id, spec)
		return "", false, false
	}

	if image.scanned {
		// if multiple controllers grab an image, only one will process, and need to signal others to stand down once scan complete
		log.Printf("Requested image %s has completed scan data. Skipping as duplicate request\n", spec)
		return "", false, true
	}

	reqHash, ok := arb.requestedImages[spec]
	if !ok {
		// if multiple controllers grab an image, only one will process, and need to signal others to stand down once scan complete
		log.Printf("Requested image %s isn't in queue\n", spec)
		return "", false, true
	}

	assignedImage, ok := arb.assignedImages[reqHash]
	if ok && strings.Compare(cd.info.Id, assignedImage.ControllerID) != 0 {
		// need to check if previously assigned to another controller -- avoids dup scan as well as worker exhaustion
		log.Printf("Requested image %s from %s is currently assigned to %s\n", spec, cd.info.Id, assignedImage.ControllerID)
		return "", false, false
	}

	if !cd.AssignScan(spec) {
		// we've probably run out of workers, but could be a data error. the latter ges cleaned up once scan is done on legit node
		log.Printf("Controller %s is unable to scan %s at this time.\n", cd.info.Id, spec)
		return "", false, false
	}

	var assigned assignImage
	assigned.ControllerID = cd.info.Id
	assigned.ImageSpec = spec
	assigned.AssignTime = time.Now()
	assigned.UpdateTime = time.Now()

	arb.assignedImages[reqHash] = &assigned

	log.Printf("Assigned image %s identified as %s to controller %s\n", spec, reqHash, assigned.ControllerID)

	return reqHash, true, false
}

func (arb *Arbiter) foundImage(w http.ResponseWriter, r *http.Request) {
	log.Println("Request foundImage")
	var i imageInfo
	_ = json.NewDecoder(r.Body).Decode(&i)

	if len(i.ControllerID) == 0 || len(i.ImageSpec) == 0 {
		log.Printf("Got junk on foundImage API: %s\n", r.Body)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	cd, ok := arb.controllerDaemons[i.ControllerID]
	if !ok {
		log.Printf("Unknown controller [%s] identified image: %s\n", i.ControllerID, i.ImageSpec)
		w.WriteHeader(500)
		return
	}

	var resp imageResult
	resp.RequestId = arb.saveFoundImage(i.ImageSpec, cd)
	resp.StartScan = false
	resp.SkipScan = false

	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error encoding image response: %s\n", err)
		w.WriteHeader(500)
		return
	}

	log.Println("Done foundImage")

}

func (arb *Arbiter) saveFoundImage(spec string, cd *controllerDaemon) string {
	arb.Lock()
	defer arb.Unlock()

	reqHash, ok := arb.requestedImages[spec]
	if !ok {
		reqHashBytes := md5.Sum([]byte(spec))
		reqHash = hex.EncodeToString(reqHashBytes[:])
		arb.requestedImages[spec] = reqHash
		log.Printf("Added spec %s as %s found by controller %s\n", spec, reqHash, cd.info.Id)
	}

	cd.AddScanRequest(spec, reqHash)

	return reqHash
}

// registerControllerAlive is the first communication from a controller to the arbiter.
// It's goal is to first register that a given controller exists, and second to ensure it
// is still alive.

func (arb *Arbiter) registerControllerAlive(w http.ResponseWriter, r *http.Request) {
	var ci controllerInfo
	_ = json.NewDecoder(r.Body).Decode(&ci)

	if len(ci.Id) == 0 {
		log.Printf("Got junk on registerControllerAlive API: %s\n", r.Body)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	cd, ok := arb.controllerDaemons[ci.Id]
	if !ok {
		// add the controller daemon
		log.Printf("Adding new controller for %s with %d workers", ci.Id, ci.WorkerCount)
		cd = newControllerDaemon(ci.Id, ci.WorkerCount)
		arb.controllerDaemons[ci.Id] = cd
	}

	cd.Heartbeat()

	w.WriteHeader(http.StatusCreated)

}
