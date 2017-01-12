package controller

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
	images		map[string]*ScanImage
	sync.RWMutex
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
		images:		make(map[string]*ScanImage),

	}
}

func (c *Controller) Start() {

	log.Println ("Starting controller ....")
	dispatcher := NewDispatcher(c.jobQueue, MaxWorker)
	dispatcher.Run()

	return
}

func (c* Controller) Watch () {

	log.Println ("Starting watcher ....")
	watcher := NewWatcher(c.openshiftClient, c)
	watcher.Run()


	return

}

func (c *Controller) Stop() {

	log.Println ("Waiting for scan queue to drain before stopping...")
	c.wait.Wait()
	
	log.Println("Scan queue empty.")
	log.Println("Controller stopped.")
	return

}

func (c *Controller) Load(done <-chan struct{}) {

	log.Println ("Starting load of existing images ...")
	
	c.getImages( done )

	log.Println ("Done load of existing images.")

	return
}

func (c *Controller) AddImage (ID string, Reference string) {

		c.Lock()
		_, ok := c.images[Reference]
		if (!ok) {

			imageItem := NewScanImage (ID, Reference)
			
			c.images[Reference] = imageItem

log.Printf ("Added %s to image map\n", imageItem.digest )
			job := Job {
				ScanImage: imageItem,
				controller: c,	
			}

			job.Load()
			c.jobQueue <- job

		}
		c.Unlock()

}

func (c *Controller) getImages (done <-chan struct{}) {

	imageList, err := c.openshiftClient.Images().List(kapi.ListOptions{})

	if err != nil {
		log.Println(err)
		return 
	}

	if imageList == nil {
		log.Println("No images")
		return
	}

	for _, image := range imageList.Items {
		c.AddImage (image.DockerImageMetadata.ID, image.DockerImageReference)
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

func init () {
	log.SetOutput(os.Stdout)
}
