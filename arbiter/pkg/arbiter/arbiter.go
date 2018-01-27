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
	"time"

	bdscommon "ose-scanner/common"

	osclient "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"

	kapi "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kclient "k8s.io/client-go/kubernetes"
)

type HubParams struct {
	Config  *bdscommon.HubConfig
	Version string
}

var Hub HubParams

type PodInfo struct {
	namespace string
	name      string
}

type Arbiter struct {
	openshiftClient   *osclient.ImageV1Client
	kubeClient        *kclient.Clientset
	jobQueue          chan *Job
	wait              sync.WaitGroup
	controllerDaemons map[string]*controllerDaemon
	images            map[string]*ScanImage
	requestedImages   map[string]string
	assignedImages    map[string]*assignImage
	imageUsage        map[string][]PodInfo
	annotation        *bdscommon.Annotator
	lastScan          time.Time
	sync.RWMutex
}

func NewArbiter(os *osclient.ImageV1Client, kc *kclient.Clientset, hub HubParams) *Arbiter {

	Hub = hub

	jobQueue := make(chan *Job)

	return &Arbiter{
		openshiftClient:   os,
		kubeClient:        kc,
		jobQueue:          jobQueue,
		images:            make(map[string]*ScanImage),
		requestedImages:   make(map[string]string),
		assignedImages:    make(map[string]*assignImage),
		controllerDaemons: make(map[string]*controllerDaemon),
		annotation:        bdscommon.NewAnnotator(hub.Version, hub.Config.Host),
		imageUsage:        make(map[string][]PodInfo),
	}
}

func (arb *Arbiter) Start() {

	log.Println("Starting arbiter ....")
	dispatcher := NewDispatcher(arb.jobQueue)
	dispatcher.Run()

	ticker := time.NewTicker(time.Minute * 30)
	go func() {
		for t := range ticker.C {
			log.Println("Processing notification status at: ", t)
			arb.queueImagesForNotification()
			arb.queuePodsForNotification()
		}
	}()

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

	log.Println("Starting load of existing pods ...")

	arb.getPods()

	log.Println("Done load of existing configuration. Waiting for initial processing to complete...")

	arb.queueImagesForNotification()
	arb.queuePodsForNotification()

	arb.lastScan = time.Now()
	duration := time.Since(arb.lastScan)

	for duration.Seconds() < 15 {
		time.Sleep(5 * time.Second)
		duration = time.Since(arb.lastScan)
	}

	log.Println("Initial processing complete.")

	return
}

func (arb *Arbiter) setImageStatus(result bool, Reference string) {
	image, ok := arb.images[Reference]
	if ok {
		image.scanned = result
		log.Printf("Set scan status for %s to %t\n", Reference, result)
	} else {
		log.Printf("Unknown image %s found with scan status of %t\n", Reference, result)
	}

	arb.lastScan = time.Now()
}

func (arb *Arbiter) DoneScan(result bool, Reference string) {
	arb.Lock()
	defer arb.Unlock()

	arb.setImageStatus(result, Reference)

	arb.wait.Done()
}

func (arb *Arbiter) DonePod(Reference string) {
	arb.Lock()
	defer arb.Unlock()

	log.Printf("Done processing pod image %s\n", Reference)

	arb.lastScan = time.Now()

	arb.wait.Done()

}

func (arb *Arbiter) Add() {
	arb.wait.Add(1)
}

func (arb *Arbiter) addImage(ID string, Reference string) {

	arb.Lock()
	defer arb.Unlock()

	_, ok := arb.images[Reference]
	if !ok {

		imageItem := newScanImage(ID, Reference, arb.annotation)
		log.Printf("Added %s to image map\n", imageItem.digest)
		arb.images[Reference] = imageItem
	}
}

func (arb *Arbiter) queueImagesForNotification() {
	for _, image := range arb.images {
		log.Printf("Queuing image %s for notification check\n", image.digest)
		job := newJob(image, nil, arb)

		job.Load()
		arb.jobQueue <- job
	}
}

func (arb *Arbiter) queuePodsForNotification() {
	for imageName, podInfo := range arb.imageUsage {
		log.Printf("Queuing pod image %s for notification check\n", imageName)

		pi := newPodImage(imageName, podInfo, arb.annotation)
		job := newJob(nil, pi, arb)

		job.Load()
		arb.jobQueue <- job
	}
}

func (arb *Arbiter) getImages(done <-chan struct{}) {

	if arb.openshiftClient != nil {

		imageList, err := arb.openshiftClient.Images().List(metav1.ListOptions{})

		if err != nil {
			log.Println(err)
			return
		}

		if imageList == nil {
			log.Println("No images")
			return
		}

		for _, image := range imageList.Items {
			arb.addImage(image.GetName(), image.DockerImageReference)
		}

	}

	return

}

func (arb *Arbiter) getPods() {

	podList, err := arb.kubeClient.CoreV1().Pods(kapi.NamespaceAll).List(metav1.ListOptions{})

	if err != nil {
		log.Println(err)
		return
	}

	if podList == nil {
		log.Println("No running pods")
		return
	}

	pods := podList.Items

	log.Printf("Found %d running pods\n", len(pods))

	arb.Lock()
	arb.imageUsage = make(map[string][]PodInfo) // clear for resync
	arb.Unlock()

	for _, pod := range pods {
		time.Sleep(5 * time.Millisecond) // on large systems allow for some time to breathe

		log.Printf("Discovered pod: %s\n", pod.ObjectMeta.Name)

		if pod.Status.Phase == kapi.PodPending {
			go arb.waitPodRunning(pod.ObjectMeta.Name, pod.ObjectMeta.Namespace)
			continue
		} else if pod.Status.Phase == kapi.PodRunning {
			arb.processPod(&pod)
			continue
		} else {
			log.Printf("Pod %s in phase: %s. Skipping\n", pod.ObjectMeta.Name, pod.Status.Phase)
			continue
		}
	}

	return
}

func (arb *Arbiter) waitPodRunning(podName string, namespace string) {

	log.Printf("Waiting for pod %s to enter running state.\n", podName)

	for {
		pod, err := arb.kubeClient.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})

		if err != nil {
			log.Printf("Error getting pod %s. Error: %s\n", podName, err)
			break
		}

		if pod.Status.Phase == kapi.PodPending {
			// defer processing until its running - nominally this delay allows for image download to node
			time.Sleep(time.Millisecond * 50)
			continue
		}

		if pod.Status.Phase != kapi.PodRunning {
			log.Printf("Pod %s in phase: %s. Expected 'running' - skipping\n", pod.ObjectMeta.Name, pod.Status.Phase)
			break
		}

		arb.processPod(pod)
		break
	}

}

func (arb *Arbiter) processPod(pod *kapi.Pod) {

	log.Printf("Processing pod %s\n", pod.ObjectMeta.Name)

	statuses := map[string]kapi.ContainerStatus{}
	for _, status := range pod.Status.ContainerStatuses {
		statuses[status.Name] = status
	}

	for _, container := range pod.Spec.InitContainers {
		status := statuses[container.Name]
		log.Printf("\tInit Container %s with image %s having ID %s on pod %s\n", container.Name, container.Image, status.ImageID, pod.ObjectMeta.Name)
		arb.registerPodUsage(status.ImageID, pod.ObjectMeta.Name, pod.ObjectMeta.Namespace)
	}

	for _, container := range pod.Spec.Containers {
		status := statuses[container.Name]
		log.Printf("\tContainer %s with image %s having ID %s on pod %s\n", container.Name, container.Image, status.ImageID, pod.ObjectMeta.Name)
		arb.registerPodUsage(status.ImageID, pod.ObjectMeta.Name, pod.ObjectMeta.Namespace)
	}
}

func (arb *Arbiter) registerPodUsage(image string, podName string, namespace string) {

	arb.Lock()
	defer arb.Unlock()

	arb.imageUsage[image] = append(arb.imageUsage[image], PodInfo{namespace, podName})
}

// ValidateConfig validates if the Hub server configuration is valid. A login attempt will be performed.
func (arb *Arbiter) ValidateConfig() bool {
	hubServer := bdscommon.NewHubServer(Hub.Config)
	defer hubServer.Logout()
	return hubServer.Login()
}

func init() {
	log.SetFlags(log.LstdFlags)
	log.SetOutput(os.Stdout)
}
