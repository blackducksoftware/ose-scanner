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
	"log"
	"time"
)

type Worker struct {
	id         int
	jobQueue   chan Job
	workerPool chan chan Job
	quit       chan bool
	arbiter    *Arbiter
}

func NewWorker(index int, workerPool chan chan Job, arbiter *Arbiter) Worker {
	return Worker{
		id:         index,
		workerPool: workerPool,
		jobQueue:   make(chan Job),
		quit:       make(chan bool),
		arbiter:    arbiter,
	}
}

func (w Worker) Start() {
	log.Printf("Starting worker %d\n", w.id)

	go func() {
		for {
			w.workerPool <- w.jobQueue

			select {
			case job := <-w.jobQueue:
				w.RequestScanAuthorization(job)

				job.Done()

			case <-w.quit:
				// we have received a signal to stop
				log.Printf("Aborting worker %d\n", w.id)
				return
			}
		}
	}()
}

func (w Worker) RequestScanAuthorization(job Job) {

	image := job.ScanImage

	log.Printf("Received scan job for image %s:%s\n", image.digest, image.imageId)

	if !image.exists() {
		log.Printf("Image %s not in local Docker engine. Skipping scan.\n", image.digest)
		return
	}

	if job.imageScanned(image.digest) {
		log.Printf("Image %s previously queued and scanned. Skipping scan.\n", image.digest)
		return
	}

	for {
		connected := w.arbiter.heartbeat()
		if connected {
			break
		}

		log.Printf("Arbiter peer offline. Postponing scan of image: %s\n", image.digest)
		// in the real world we shouldn't ever be offline from our arbiter
		time.Sleep(time.Second * 30)
		continue
	}

	_, skip := w.arbiter.alertImage(image.digest, image.imageId)

	if skip {
		log.Printf("Skipping scan of alerted image at arbiter direction: %s\n", image.digest)
		return
	}

	log.Printf("Requesting authorization to scan image %s\n", image.digest)

	for {
		requestHash, skip, startScan := w.arbiter.requestImage(image.digest, image.imageId)

		if skip {
			log.Printf("Skipping scan of image at arbiter direction: %s\n", image.digest)
			break
		}

		if !startScan {
			time.Sleep(time.Second * 30)
			continue
		}

		ok, info := job.GetAnnotationInfo()
		if !ok {
			log.Printf("Error getting annotation info for image: %s", image.digest)
		}

		log.Printf("Starting scan of %s with arbiter hash of %s\n", image.digest, requestHash)

		err, results := image.scan(info)

		if err != nil {
			log.Printf("Scan image error for %s of %s\n", image.digest, err)
			w.arbiter.abortScan(requestHash) // abort will return the item to the arbiter queue, but remove it from ours
			break
		}
		if ok && err == nil {
			ok = job.UpdateAnnotationInfo(results)
			if ok {
				log.Printf("Updated annotation info for image: %s", image.digest)
			}
		}
		w.arbiter.scanDone(requestHash)
		break
	}

	log.Printf("Completed processing for image %s\n", image.digest)
	return

}
