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
	"os"
	"sync"
	"time"

	bdscommon "github.com/blackducksoftware/ose-scanner/common"

	osclient "github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"

	"github.com/spf13/pflag"
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/meta"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/runtime"
)

type HubParams struct {
	Config  *bdscommon.HubConfig
	Scanner string
	Workers int
	Version string
}

type PodInfo struct {
	namespace string
	name      string
}

type Controller struct {
	openshiftClient *osclient.Client
	kubeClient      *kclient.Client
	mapper          meta.RESTMapper
	typer           runtime.ObjectTyper
	f               *clientcmd.Factory
	jobQueue        chan Job
	wait            sync.WaitGroup
	images          map[string]*ScanImage
	annotation      *bdscommon.Annotator
	imageUsage      map[string][]PodInfo
	sync.RWMutex
	hubParams *HubParams
}

func NewController(os *osclient.Client, kc *kclient.Client, hub *HubParams) *Controller {

	f := clientcmd.New(pflag.NewFlagSet("empty", pflag.ContinueOnError))
	mapper, typer := f.Object(false)

	jobQueue := make(chan Job, hub.Workers)

	return &Controller{
		openshiftClient: os,
		kubeClient:      kc,
		mapper:          mapper,
		typer:           typer,
		f:               f,
		jobQueue:        jobQueue,
		images:          make(map[string]*ScanImage),
		annotation:      bdscommon.NewAnnotator(hub.Version, hub.Config.Host),
		hubParams:       hub,
		imageUsage:      make(map[string][]PodInfo),
	}
}

func (c *Controller) Start(arb *Arbiter) {

	log.Println("Starting controller ....")
	dispatcher := NewDispatcher(c.jobQueue, c.hubParams.Workers)
	arb.heartbeat()
	dispatcher.Run(arb)

	return
}

func (c *Controller) Watch(done <-chan struct{}) {

	log.Println("Starting watcher ....")
	watcher := NewWatcher(c.openshiftClient, c)
	watcher.Run()

	return

}

func (c *Controller) Stop() {

	log.Println("Waiting for scan queue to drain before stopping...")
	c.wait.Wait()

	log.Println("Scan queue empty.")
	log.Println("Controller stopped.")
	return

}

func (c *Controller) Load() {

	log.Println("Starting load of existing images ...")

	c.getImages()

	log.Println("Done load of existing images.")

	log.Println("Starting load of existing pods ...")

	c.getPods()

	log.Println("Done load of existing pods.")

	return
}

func (c *Controller) AddImage(ID string, Reference string) {

	c.Lock()
	defer c.Unlock()

	image, ok := c.images[Reference]
	if !ok {

		imageItem := newScanImage(ID, Reference, c.annotation, c.hubParams.Config, c.hubParams.Scanner)

		if imageItem == nil {
			log.Printf("Image %s:%s not present in node. Not adding to image map.\n", Reference, ID)
			return
		}

		c.queueImage(imageItem, Reference)

	} else if ID != image.imageId {
		log.Printf("Image %s already in image map with imageId %s but have imageId %s\n", Reference, image.imageId, ID)

		validOriginal := image.exists()
		originalScanned := image.scanned

		if !validOriginal {
			log.Printf("Removing original imageId %s as it no longer exists on this node.\n", image.imageId)
			delete(c.images, Reference)
		}

		newImage := newScanImage(ID, Reference, c.annotation, c.hubParams.Config, c.hubParams.Scanner)

		if newImage == nil {
			log.Printf("Requested imageId %s does not exist on this node. Skipping.\n", Reference)
			return
		} else if validOriginal {
			// we have both an original and an update which is valid
			if originalScanned {
				// since we see the original, and its been scanned, and it has the same pullspec, then we replace it and refresh
				log.Printf("Replacing queued and scanned imageId %s with %s for image %s\n", image.imageId, newImage.imageId, Reference)
				c.queueImage(newImage, Reference)
			} else if image.imageId != newImage.imageId {
				/*
				* in this case we have both a unscanned original and an unscanned update
				* not quite certain what to do here just yet TODO
				 */
				log.Printf("***ImageId %s hasn't been scanned, but we've found imageId %s for image %s as a replacement ****\n", image.imageId, newImage.imageId, Reference)
				c.queueImage(newImage, Reference)
			}
		} else {
			// add the image to queue
			c.queueImage(newImage, Reference)
		}

	}

}

func (c *Controller) queueImage(imageItem *ScanImage, Reference string) {

	job := Job{
		ScanImage:  imageItem,
		controller: c,
	}

	if !job.IsImageStreamScanNeeded(c.hubParams.Config) {
		log.Printf("Image %s previously scanned. Skipping scan.\n", imageItem.digest)
		imageItem.scanned = true
		return
	}

	c.images[Reference] = imageItem
	log.Printf("Added %s:%s to image map\n", imageItem.digest, imageItem.imageId)

	job.Load()
	c.jobQueue <- job

}

func (c *Controller) imageScanned(Reference string) bool {

	c.Lock()
	defer c.Unlock()

	imageItem, ok := c.images[Reference]
	if !ok {
		log.Printf("Requested scan status for unknown image %s\n", Reference)
		return true
	}

	return imageItem.scanned
}

func (c *Controller) RemoveImage(ID string, Reference string) {

	c.Lock()
	defer c.Unlock()

	_, ok := c.images[Reference]
	if ok {
		delete(c.images, Reference)
		log.Printf("Removed %s from map\n", Reference)
	}

}

func (c *Controller) getImages() {

	if c.openshiftClient == nil {
		// if there's no OpenShift client, there can't be any image annotations
		log.Println("Not running in OpenShift mode")
		return
	}

	imageList, err := c.openshiftClient.Images().List(kapi.ListOptions{})

	if err != nil {
		log.Println(err)
		return
	}

	if imageList == nil {
		log.Println("No images")
		return
	}

	images := imageList.Items

	log.Printf("Discovered %d images\n", len(images))

	for _, image := range images {
		c.AddImage(image.DockerImageMetadata.ID, image.DockerImageReference)
		time.Sleep(10 * time.Millisecond)
	}

	log.Printf("Queued all images\n")

	return

}

func (c *Controller) getPods() {

	podList, err := c.kubeClient.Pods(kapi.NamespaceAll).List(kapi.ListOptions{})

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

	for _, pod := range pods {
		time.Sleep(10 * time.Millisecond)

		log.Printf("Discovered pod: %s\n", pod.ObjectMeta.Name)

		if pod.Status.Phase == kapi.PodPending {
			// defer processing until its running
			go c.waitPodRunning(pod.ObjectMeta.Name, pod.ObjectMeta.Namespace)
			continue
		}

		if pod.Status.Phase != kapi.PodRunning {
			log.Printf("Pod %s in phase: %s. Skipping\n", pod.ObjectMeta.Name, pod.Status.Phase)
			continue
		}

		c.processPod(&pod)

	}

	return

}

func (c *Controller) waitPodRunning(podName string, namespace string) {

	log.Printf("Waiting for pod %s to enter running state.\n", podName)

	for {
		pod, err := c.kubeClient.Pods(namespace).Get(podName)

		if err != nil {
			log.Printf("Error getting pod %s. Error: %s\n", podName, err)
			break
		}

		if pod.Status.Phase == kapi.PodPending {
			// defer processing until its running - nominally this delay allows for image download to node
			time.Sleep(time.Second * 5)
			continue
		}

		if pod.Status.Phase != kapi.PodRunning {
			log.Printf("Pod %s in phase: %s. Expected 'running' - skipping\n", pod.ObjectMeta.Name, pod.Status.Phase)
			break
		}

		c.processPod(pod)
		break
	}

}

func (c *Controller) processPod(pod *kapi.Pod) {

	log.Printf("Processing pod %s\n", pod.ObjectMeta.Name)

	d := NewDocker()

	for _, container := range pod.Spec.Containers {
		log.Printf("\tContainer %s with image %s on pod %s\n", container.Name, container.Image, pod.ObjectMeta.Name)

		digests, imageId, found := d.digestFromImage(container.Image)

		if !found {
			log.Printf("\tImage %s not found\n", container.Image)
			continue
		}

		for _, digest := range digests {
			log.Printf("\tFound pod image %s: %s\n", imageId, digest)
			c.AddImage(imageId, digest)
			c.registerPodUsage(digest, pod.ObjectMeta.Name, pod.ObjectMeta.Namespace)
		}
	}
}

func (c *Controller) registerPodUsage(image string, podName string, namespace string) {
	c.imageUsage[image] = append(c.imageUsage[image], PodInfo{namespace, podName})
}

func (c *Controller) ValidateConfig() bool {
	hubServer := bdscommon.NewHubServer(c.hubParams.Config)
	defer hubServer.Logout()
	return hubServer.Login()
}

func (c *Controller) ValidateDockerConfig() bool {
	docker := NewDocker()
	if docker.client == nil {
		log.Printf("Unable to connect to Docker runtime\n")
		return false
	}

	_, err := docker.client.Info()
	if err != nil {
		log.Printf("Unable to connect to Docker runtime. %s\n", err)
		return false
	}

	log.Printf("Validated Docker runtime connection\n")
	return true

}

func init() {
	log.SetFlags(log.LstdFlags)
	log.SetOutput(os.Stdout)
}
