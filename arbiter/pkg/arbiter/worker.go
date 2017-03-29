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
)

type Worker struct {
	id         int
	jobQueue   chan Job
	workerPool chan chan Job
	quit       chan bool
}

func NewWorker(index int, workerPool chan chan Job) Worker {
	return Worker{
		id:         index,
		workerPool: workerPool,
		jobQueue:   make(chan Job),
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
				/*if err := job.ScanImage.scan(); err != nil {
					log.Printf("Error scanning image: %s", err.Error())
				}*/
				job.Done()

			case <-w.quit:
				// we have received a signal to stop
				log.Printf("Aborting worker %d\n", w.id)
				return
			}
		}
	}()
}
