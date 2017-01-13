package main

import (
	"log"
	"os"

	"github.com/blackducksoftware/ose-scanner/controller/pkg/controller"
	_ "github.com/openshift/origin/pkg/api/install"
	osclient "github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"

	"github.com/spf13/pflag"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
)

var hub controller.HubParams

func main() {

	pflag.Parse() 

	if !checkExpectedCmdlineParams() {
		return
	}

	config, err := clientcmd.DefaultClientConfig(pflag.NewFlagSet("empty", pflag.ContinueOnError)).ClientConfig()
	kubeClient, err := kclient.New(config)
	if err != nil {
		log.Printf("Error creating cluster config: %s", err)
		os.Exit(1)
	}
	openshiftClient, err := osclient.New(config)
	if err != nil {
		log.Printf("Error creating OpenShift client: %s", err)
		os.Exit(2)
	}

	c := controller.NewController(openshiftClient, kubeClient, hub)

	done := make(chan struct{})
	defer close(done)

	c.Start()

	c.Load(done)

	c.Watch()
	
	c.Stop()

}

func init() {
	log.SetFlags(log.LstdFlags)
	log.SetOutput(os.Stdout)

	pflag.StringVar(&hub.Host, "h", "REQUIRED", "The hostname of the Black Duck Hub server.")
	pflag.StringVar(&hub.Port, "p", "REQUIRED", "The port the Hub is communicating on")
	pflag.StringVar(&hub.Scheme, "s", "REQUIRED", "The communication scheme [http,https].")
	pflag.StringVar(&hub.Username, "u", "REQUIRED", "The Black Duck Hub user")
	pflag.StringVar(&hub.Password, "w", "REQUIRED", "Password for the user.")
	pflag.StringVar(&hub.Scanner, "scanner", "REQUIRED", "Scanner image")
	pflag.IntVar(&hub.Workers, "workers", controller.MaxWorkers, "Number of container workers")
}

func checkExpectedCmdlineParams() bool {
	// NOTE: At this point we don't have a logger yet, so don't try and use it.

	if hub.Host == "REQUIRED" {
		log.Println("-h host is required\n")
		pflag.PrintDefaults()
		return false
	} 
	
	if hub.Port == "REQUIRED" {
		log.Println("-p port is required\n")
		pflag.PrintDefaults()
		return false
	} 
	
	if hub.Scheme == "REQUIRED" {
		log.Println("-s scheme is required\n")
		pflag.PrintDefaults()
		return false
	} 
	
	if hub.Username == "REQUIRED" {
		log.Println("-u username is required\n")
		pflag.PrintDefaults()
		return false
	}

	if hub.Password == "REQUIRED" {
		log.Println("-w password is required\n")
		pflag.PrintDefaults()
		return false
	}

	if hub.Scanner == "REQUIRED" {
		log.Println("-scanner Hub scanner image is required\n")
		pflag.PrintDefaults()
		return false
	}

	if hub.Workers < 1 {
		log.Printf("Setting workers from %d to %d\n", hub.Workers)
		hub.Workers = controller.MaxWorkers
	}

	return true
}


