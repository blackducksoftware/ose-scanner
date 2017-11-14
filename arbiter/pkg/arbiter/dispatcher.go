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

const MaxWorkers = 5

type Dispatcher struct {
	// A pool of workers channels that are registered with the dispatcher
	jobQueue   chan *Job
	workerPool chan chan *Job
}

func NewDispatcher(jobQueue chan *Job) *Dispatcher {
	pool := make(chan chan *Job, MaxWorkers)
	return &Dispatcher{
		jobQueue:   jobQueue,
		workerPool: pool,
	}
}

func (d *Dispatcher) Run() {
	// starting n number of workers
	for i := 0; i < MaxWorkers; i++ {
		worker := NewWorker(i+1, d.workerPool)
		worker.Start()
	}

	go d.dispatch()
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case job := <-d.jobQueue:
			// a job request has been received
			go func(job *Job) {

				// try to obtain a worker job channel that is available.
				// this will block until a worker is idle
				workerJobQueue := <-d.workerPool

				// dispatch the job to the worker job channel
				workerJobQueue <- job
			}(job)
		}
	}
}
