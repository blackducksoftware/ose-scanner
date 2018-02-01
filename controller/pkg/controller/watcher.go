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
	"time"

	imageapi "github.com/openshift/api/image/v1"
	osclient "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	kapi "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

type Watcher struct {
	openshiftClient *osclient.ImageV1Client
	Namespace       string
	controller      *Controller
}

// Create a new watcher
func NewWatcher(os *osclient.ImageV1Client, c *Controller) *Watcher {

	namespace := kapi.NamespaceAll

	wc := &Watcher{
		openshiftClient: os,
		Namespace:       namespace,
		controller:      c,
	}
	return wc
}

func (w *Watcher) Run() {

	if os.Getenv("OSE_KUBERNETES_CONNECTOR") != "Y" && w.openshiftClient != nil {
		log.Println("Subscribing to image stream events ....")

		_, k8sCtl := cache.NewInformer(
			&cache.ListWatch{
				ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
					return w.openshiftClient.ImageStreams(w.Namespace).List(opts)
				},
				WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
					return w.openshiftClient.ImageStreams(w.Namespace).Watch(opts)
				},
			},
			&imageapi.ImageStream{},
			2 * time.Minute,
			cache.ResourceEventHandlerFuncs{
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

	podWatchList := cache.NewListWatchFromClient(w.controller.kubeClient.CoreV1().RESTClient(), "pods", kapi.NamespaceAll, fields.Everything())

	_, k8sPodCtl := cache.NewInformer(
		podWatchList,
		&kapi.Pod{},
		time.Minute,
		cache.ResourceEventHandlerFuncs{
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

	log.Printf("ImageStream created: %s\n", digest)

	for _, events := range tags {
		tagEvents := events.Items
		if len(tagEvents) == 0 {
			log.Printf("ImageStream %s has no associated image\n", digest)
			return
		}
		ref := tagEvents[0].Image
		image, err := w.openshiftClient.Images().Get(ref, metav1.GetOptions{})
		if err != nil {
			log.Printf("Error seeking new image %s@%s: %s\n", digest, ref, err)
			continue
		}
		w.controller.AddImage(image.GetName(), image.DockerImageReference)
	}
}

// care should be used with updates. Per kube docs:
//    OnUpdate is also called when a re-list happens, and it will
//      get called even if nothing changed. This is useful for periodically
//      evaluating or syncing something.
func (w *Watcher) ImageUpdated(is *imageapi.ImageStream) {

	tags := is.Status.Tags
	if tags == nil {
		log.Println("Image updated, but no tags")
		return
	}

	digest := is.Spec.DockerImageRepository

	log.Printf("ImageStream updated: %s\n", digest)

	for _, events := range tags {
		tagEvents := events.Items
		if len(tagEvents) == 0 {
			log.Printf("ImageStream %s has no associated image\n", digest)
			return
		}
		ref := tagEvents[0].Image
		image, err := w.openshiftClient.Images().Get(ref, metav1.GetOptions{})
		if err != nil {
			log.Printf("Error seeking updated image %s@%s: %s\n", digest, ref, err)
			continue
		}

		w.controller.AddImage(image.GetName(), image.DockerImageReference)
	}
}

func (w *Watcher) ImageDeleted(is *imageapi.ImageStream) {

	tags := is.Status.Tags
	if tags == nil {
		log.Println("Image deleted, but no tags")
		return
	}

	digest := is.Spec.DockerImageRepository

	for _, events := range tags {
		ref := events.Items[0].Image
		image, err := w.openshiftClient.Images().Get(ref, metav1.GetOptions{})
		if err != nil {
			log.Printf("Error seeking deleted image %s@%s: %s\n", digest, ref, err)
			continue
		}

		w.controller.RemoveImage(image.GetName(), image.DockerImageReference)
	}
}

func (w *Watcher) PodCreated(pod *kapi.Pod) {
	log.Printf("Pod creation for %s in namespace %s \n", pod.ObjectMeta.Name, pod.ObjectMeta.Namespace)

	if pod.Status.Phase == kapi.PodPending {
		// defer processing until its running
		go w.controller.waitPodRunning(pod.ObjectMeta.Name, pod.ObjectMeta.Namespace)
		return
	}

	if pod.Status.Phase != kapi.PodRunning {
		log.Printf("Pod %s in phase: %s. Skipping\n", pod.ObjectMeta.Name, pod.Status.Phase)
		return
	}

	w.controller.processPod(pod)

}
