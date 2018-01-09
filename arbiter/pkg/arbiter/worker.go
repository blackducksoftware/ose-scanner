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
	"os"
)

type Worker struct {
	id         int
	jobQueue   chan *Job
	workerPool chan chan *Job
	quit       chan bool
}

func NewWorker(index int, workerPool chan chan *Job) Worker {
	return Worker{
		id:         index,
		workerPool: workerPool,
		jobQueue:   make(chan *Job),
		quit:       make(chan bool),
	}
}

func (w Worker) Start() {
	log.Printf("Starting worker %d\n", w.id)

	go func() {
		for {
			w.workerPool <- w.jobQueue

			select {
			case job := <-w.jobQueue:
				scanned := false
				log.Printf("Worker OSE_KUBERNETES_CONNECTOR:%s:\n", os.Getenv("OSE_KUBERNETES_CONNECTOR"))
				if (os.Getenv("OSE_KUBERNETES_CONNECTOR") != "Y" && job.ScanImage != nil) {
					scanned = w.ProcessScanImage(job)
				} else if job.PodImage != nil {
					scanned = w.ProcessPodImage(job)
				} else {
					log.Printf("***Error*** Worker %d found job with unknown activity\n", w.id)
				}

				job.Done(scanned)

			case <-w.quit:
				// we have received a signal to stop
				log.Printf("Aborting worker %d\n", w.id)
				return
			}
		}
	}()
}

func (w Worker) ProcessScanImage(job *Job) (scanned bool) {

	log.Printf("Processing scan image %s\n", job.ScanImage.digest)
	scanned = false
	ok, info := job.GetImageAnnotationInfo()
	if !ok {
		log.Printf("Error getting annotation info for image: %s", job.ScanImage.digest)
		return
	}

	err, results := job.ScanImage.versionResults(info)
	if err != nil {
		log.Printf("Error getting notification results for %s: %s", job.ScanImage.digest, err.Error())
		return
	}

	ok = job.UpdateImageAnnotationInfo(results)
	if ok {
		scanned = true
		log.Printf("Updated annotation info for image: %s", job.ScanImage.digest)
	}

	return
}

func (w Worker) ProcessPodImage(job *Job) (scanned bool) {

	log.Printf("Processing pod image %s\n", job.PodImage.imageName)

	scanned = false

	err, results := job.PodImage.ScanResults()

	if err != nil {
		log.Printf("Error getting notification results for pod %s:%s", job.PodImage.imageName, err.Error())
		return
	}

	ok := job.ApplyAnnotationInfoToPods(results)
	if ok {
		log.Printf("Applied annotation info for pod set %s\n", job.PodImage.imageName)
		scanned = true
	}

	return
}
