package controller

import (
	"fmt"
	"strings"
	"sync"

	osclient "github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"

	"github.com/spf13/pflag"
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/meta"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/runtime"
	//"k8s.io/kubernetes/pkg/util/wait"
	//"k8s.io/kubernetes/pkg/watch"
)

const (
	displayNameOldAnnotation = "displayName"
	displayNameAnnotation    = "openshift.io/display-name"
)

type HubParams struct {
	Host     string
	Port     string
	Scheme   string
	Username string
	Password string
	Scanner  string
}

var Hub HubParams

type Controller struct {
	openshiftClient *osclient.Client
	kubeClient      *kclient.Client
	mapper          meta.RESTMapper
	typer           runtime.ObjectTyper
	f               *clientcmd.Factory
	jobQueue	chan Job
	wait		sync.WaitGroup
}

func NewController(os *osclient.Client, kc *kclient.Client, hub HubParams) *Controller {

	f := clientcmd.New(pflag.NewFlagSet("empty", pflag.ContinueOnError))
	mapper, typer := f.Object()

	Hub = hub

	jobQueue := make(chan Job, MaxQueue)

	var wait sync.WaitGroup

	return &Controller{
		openshiftClient: os,
		kubeClient:      kc,
		mapper:          mapper,
		typer:           typer,
		f:               f,
		jobQueue:	 jobQueue,
		wait:		 wait,

	}
}

func (c *Controller) Start() {

	fmt.Println ("Starting controller")
	dispatcher := NewDispatcher(c.jobQueue, MaxWorker)
	dispatcher.Run()

	return
}

func (c *Controller) Stop() {

	fmt.Println ("Waiting for scan queue to drain before stopping...")
	c.wait.Wait()
	
	fmt.Println("Scan queue empty.")
	fmt.Println("Controller stopped.")
	return

}

func (c *Controller) Load(done <-chan struct{}) {

	fmt.Println ("Starting load of existing images ...")
	
	c.getImages( done )

	fmt.Println ("Done load of existing images.")

	return
}

func (c *Controller) getImages (done <-chan struct{}) {

	imageList, err := c.openshiftClient.Images().List(kapi.ListOptions{})

	if err != nil {
		fmt.Println(err)
		return 
	}

	if imageList == nil {
		fmt.Println("No images")
		return
	}

	for _, image := range imageList.Items {
		var imageItem ScanImage
		imageItem.imageId = image.DockerImageMetadata.ID

		tag := strings.Split(image.DockerImageReference, "@")
		imageItem.digest = image.DockerImageReference
		imageItem.taggedName = tag[0]

		//fmt.Printf("ID: %s; tag: %s\n", imageItem.imageId, imageItem.taggedName)
		
		job := Job {
			ScanImage: imageItem,
			controller: c,	
		}

		job.Load()
		c.jobQueue <- job

	}

	return 

}


// DisplayNameAndNameForProject returns a formatted string containing the name
// of the project and includes the display name if it differs.
func DisplayNameAndNameForProject(project kapi.ObjectMeta) string {
	displayName := project.Annotations[displayNameAnnotation]
	
	if len(displayName) == 0 {
		displayName = project.Annotations[displayNameOldAnnotation]
	}

	if len(displayName) > 0 && displayName != project.Name {
		// we want the machine version, not the human readable one
		return project.Name
	}
	return project.Name
}
