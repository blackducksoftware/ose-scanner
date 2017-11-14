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
	"time"

	osclient "github.com/openshift/origin/pkg/client"
	imageapi "github.com/openshift/origin/pkg/image/api"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/cache"
	"k8s.io/kubernetes/pkg/controller/framework"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/wait"
	"k8s.io/kubernetes/pkg/watch"
)

type Watcher struct {
	openshiftClient *osclient.Client
	Namespace       string
	arbiter         *Arbiter
}

// Create a new watcher
func NewWatcher(os *osclient.Client, c *Arbiter) *Watcher {

	namespace := kapi.NamespaceAll

	wc := &Watcher{
		openshiftClient: os,
		Namespace:       namespace,
		arbiter:         c,
	}
	return wc
}

func (w *Watcher) Run() {

	if w.openshiftClient != nil {
		_, k8sCtl := framework.NewInformer(
			&cache.ListWatch{
				ListFunc: func(opts kapi.ListOptions) (runtime.Object, error) {
					return w.openshiftClient.ImageStreams(w.Namespace).List(opts)
				},
				WatchFunc: func(opts kapi.ListOptions) (watch.Interface, error) {
					return w.openshiftClient.ImageStreams(w.Namespace).Watch(opts)
				},
			},
			&imageapi.ImageStream{},
			time.Minute,
			framework.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					w.ImageAdded(obj.(*imageapi.ImageStream))
				},
				DeleteFunc: func(obj interface{}) {
					w.ImageDeleted(obj.(*imageapi.ImageStream))
				},
			})

		log.Println("Watching image streams....")

		go k8sCtl.Run(wait.NeverStop)

	}

	log.Println("Subscribing to pod events ....")

	podWatchList := cache.NewListWatchFromClient(w.arbiter.kubeClient, "pods", kapi.NamespaceAll, fields.Everything())

	_, k8sPodCtl := framework.NewInformer(
		podWatchList,
		&kapi.Pod{},
		time.Minute,
		framework.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				w.PodCreated(obj.(*kapi.Pod))
			},
		})

	go k8sPodCtl.Run(wait.NeverStop)

	log.Println("Watching pods....")

	select {}
}

func (w *Watcher) ImageAdded(is *imageapi.ImageStream) {

	tags := is.Status.Tags
	if tags == nil {
		log.Println("Image added, but no tags")
		return
	}

	digest := is.Spec.DockerImageRepository

	for _, events := range tags {
		tagEvents := events.Items
		if len(tagEvents) == 0 {
			log.Printf("ImageStream %s has no associated image\n", digest)
			return
		}
		ref := tagEvents[0].Image
		image, err := w.openshiftClient.Images().Get(ref)
		if err != nil {
			log.Printf("Error seeking new image %s@%s: %s\n", digest, ref, err)
			continue
		}
		w.arbiter.addImage(image.DockerImageMetadata.ID, image.DockerImageReference)
	}
}

func (w *Watcher) ImageDeleted(is *imageapi.ImageStream) {

	tags := is.Status.Tags
	if tags == nil {
		log.Println("Image deleted, but no tags")
		return
	}

	for tag, events := range tags {
		digest := events.Items[0].Image
		log.Printf("Image %s deleted with digest %s\n", tag, digest)
	}
}

func (w *Watcher) PodCreated(pod *kapi.Pod) {
	log.Printf("Pod creation for %s in namespace %s \n", pod.ObjectMeta.Name, pod.ObjectMeta.Namespace)

	if pod.Status.Phase == kapi.PodPending {
		go w.arbiter.waitPodRunning(pod.ObjectMeta.Name, pod.ObjectMeta.Namespace)
		return
	} else if pod.Status.Phase != kapi.PodRunning {
		log.Printf("Pod %s in phase: %s. Skipping\n", pod.ObjectMeta.Name, pod.Status.Phase)
		return
	}

	w.arbiter.processPod(pod)
}
