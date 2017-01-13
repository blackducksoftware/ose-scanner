package controller

import (
	"log"
	"time"

	osclient "github.com/openshift/origin/pkg/client"
	imageapi "github.com/openshift/origin/pkg/image/api"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/cache"
	"k8s.io/kubernetes/pkg/controller/framework"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/wait"
	"k8s.io/kubernetes/pkg/watch"

)

type Watcher struct {
	openshiftClient *osclient.Client
	Namespace    string
	controller	*Controller
}

// Create a new watcher
func NewWatcher(os *osclient.Client, c *Controller) (*Watcher) {

	namespace := kapi.NamespaceAll

	wc := &Watcher{
		openshiftClient:    os,
		Namespace: namespace,
		controller: c,

	}
	return wc
}

func (w *Watcher) Run() {

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
			UpdateFunc: func(old, obj interface{}) {
				w.ImageUpdated(obj.(*imageapi.ImageStream))
			},
			DeleteFunc: func(obj interface{}) {
				w.ImageDeleted(obj.(*imageapi.ImageStream))
			},
		})


	log.Println ("Watching image streams....")

	go k8sCtl.Run(wait.NeverStop)
	select {}
}

func (w *Watcher) ImageAdded(is *imageapi.ImageStream) {

	tags := is.Status.Tags
	if tags == nil {
		log.Println ("Image added, but no tags")
		return
	}

	digest := is.Spec.DockerImageRepository 

	for _, events := range tags {
		ref := events.Items[0].Image
		image, err := w.openshiftClient.Images().Get(ref)
		if err != nil {
			log.Printf ("Error seeking new image %s@%s: %s\n", digest, ref, err)
			continue
		} 
		w.controller.AddImage (image.DockerImageMetadata.ID, image.DockerImageReference)
	}
}

func (w *Watcher) ImageUpdated(is *imageapi.ImageStream) {

	tags := is.Status.Tags
	if tags == nil {
		log.Println ("Image updated, but no tags")
		return
	}

	digest := is.Spec.DockerImageRepository 

	for _, events := range tags {
		ref := events.Items[0].Image
		image, err := w.openshiftClient.Images().Get(ref)
		if err != nil {
			log.Printf ("Error seeking updated image %s@%s: %s\n", digest, ref, err)
			continue
		} 
		
		w.controller.AddImage (image.DockerImageMetadata.ID, image.DockerImageReference)
	}
}

func (w *Watcher) ImageDeleted(is *imageapi.ImageStream) {

	tags := is.Status.Tags
	if tags == nil {
		log.Println ("Image deleted, but no tags")
		return
	}

	for tag, events := range tags {
		digest := events.Items[0].Image
		log.Printf ("Image %s deleted with digest %s\n", tag, digest)
	}
}