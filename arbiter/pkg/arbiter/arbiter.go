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
	"sync"

	osclient "github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"

	"github.com/spf13/pflag"
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/meta"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/runtime"
)

type HubParams struct {
	Config  *HubConfig
	Version string
}

var Hub HubParams

type Arbiter struct {
	openshiftClient   *osclient.Client
	kubeClient        *kclient.Client
	mapper            meta.RESTMapper
	typer             runtime.ObjectTyper
	f                 *clientcmd.Factory
	jobQueue          chan Job
	wait              sync.WaitGroup
	controllerDaemons map[string]*controllerDaemon
	images            map[string]*ScanImage
	requestedImages   map[string]string
	assignedImages    map[string]*assignImage
	annotation        *Annotator
	sync.RWMutex
}

func NewArbiter(os *osclient.Client, kc *kclient.Client, hub HubParams) *Arbiter {

	f := clientcmd.New(pflag.NewFlagSet("empty", pflag.ContinueOnError))
	mapper, typer := f.Object(false)

	Hub = hub

	jobQueue := make(chan Job)

	return &Arbiter{
		openshiftClient:   os,
		kubeClient:        kc,
		mapper:            mapper,
		typer:             typer,
		f:                 f,
		jobQueue:          jobQueue,
		images:            make(map[string]*ScanImage),
		requestedImages:   make(map[string]string),
		assignedImages:    make(map[string]*assignImage),
		controllerDaemons: make(map[string]*controllerDaemon),
		annotation:        NewAnnotator(os, hub.Version, hub.Config.Host),
	}
}

func (arb *Arbiter) Start() {

	log.Println("Starting arbiter ....")
	dispatcher := NewDispatcher(arb.jobQueue)
	dispatcher.Run()

	return
}

func (arb *Arbiter) Watch() {

	log.Println("Starting watcher ....")
	watcher := NewWatcher(arb.openshiftClient, arb)
	watcher.Run()

	return

}

func (arb *Arbiter) Stop() {

	log.Println("Waiting for notification queue to drain before stopping...")
	arb.wait.Wait()

	log.Println("Notification queue empty.")
	log.Println("Controller stopped.")
	return

}

func (arb *Arbiter) Load(done <-chan struct{}) {

	log.Println("Starting load of existing images ...")

	arb.getImages(done)

	log.Println("Done load of existing images.")

	return
}

func (arb *Arbiter) addImage(ID string, Reference string) {

	arb.Lock()
	defer arb.Unlock()

	_, ok := arb.images[Reference]
	if !ok {

		imageItem := newScanImage(ID, Reference, arb.annotation)

		arb.images[Reference] = imageItem
	}
}

func (arb *Arbiter) queueImagesForNotification() {
	for _, image := range arb.images {
		log.Printf("Added %s to notification map\n", image.digest)
		job := Job{
			ScanImage: image,
			arbiter:   arb,
		}

		job.Load()
		arb.jobQueue <- job

	}
}

func (arb *Arbiter) getImages(done <-chan struct{}) {

	imageList, err := arb.openshiftClient.Images().List(kapi.ListOptions{})

	if err != nil {
		log.Println(err)
		return
	}

	if imageList == nil {
		log.Println("No images")
		return
	}

	for _, image := range imageList.Items {
		arb.addImage(image.DockerImageMetadata.ID, image.DockerImageReference)
	}

	return

}

// ValidateConfig validates if the Hub server configuration is valid. A login attempt will be performed.
func (arb *Arbiter) ValidateConfig() bool {

	hubServer := HubServer{Config: Hub.Config}

	return hubServer.login()

}

func init() {
	log.SetFlags(log.LstdFlags)
	log.SetOutput(os.Stdout)
}
