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

package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/blackducksoftware/ose-scanner/controller/pkg/controller"
	_ "github.com/openshift/origin/pkg/api/install"
	osclient "github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"

	"github.com/spf13/pflag"
	"k8s.io/kubernetes/pkg/client/restclient"

	kclient "k8s.io/kubernetes/pkg/client/unversioned"
)

var hub controller.HubParams

func main() {

	pflag.Parse()

	if !checkExpectedCmdlineParams() {
		os.Exit(1)
	}

	config, err := restclient.InClusterConfig()
	if err != nil {
		log.Printf("Error getting in cluster config. Fallback to native config. Error message: %s", err)

		config, err = clientcmd.DefaultClientConfig(pflag.NewFlagSet("empty", pflag.ContinueOnError)).ClientConfig()
		if err != nil {
			log.Printf("Error creating default client config: %s", err)
			os.Exit(1)
		}
	}

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

	if !c.ValidateDockerConfig() {
		log.Printf("Docker configuation information isn't valid. Please verify connectivity and permissions.")
		os.Exit(1)
	}

	if !c.ValidateConfig() {
		log.Printf("Hub configuation information isn't valid. Please verify connectivity and values.")
		os.Exit(1)
	}

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

	hub.Config = &controller.HubConfig{}

	pflag.StringVar(&hub.Config.Host, "h", "REQUIRED", "The hostname of the Black Duck Hub server.")
	pflag.StringVar(&hub.Config.Port, "p", "REQUIRED", "The port the Hub is communicating on")
	pflag.StringVar(&hub.Config.Scheme, "s", "REQUIRED", "The communication scheme [http,https].")
	pflag.StringVar(&hub.Config.User, "u", "REQUIRED", "The Black Duck Hub user")
	pflag.StringVar(&hub.Config.Password, "w", "REQUIRED", "Password for the user.")
	pflag.StringVar(&hub.Scanner, "scanner", "REQUIRED", "Scanner image")
	pflag.IntVar(&hub.Workers, "workers", 0, "Number of container workers")
}

func checkExpectedCmdlineParams() bool {
	// NOTE: At this point we don't have a logger yet, so don't try and use it.

	if hub.Config.Host == "REQUIRED" {
		val := os.Getenv("BDS_HOST")
		if val == "" {
			log.Println("-h host argument or BDS_HOST environment is required")
			pflag.PrintDefaults()
			return false
		}
		hub.Config.Host = val
	}

	if hub.Config.Port == "REQUIRED" {
		val := os.Getenv("BDS_PORT")
		if val == "" {
			log.Println("-p port argument or BDS_PORT environment is required")
			pflag.PrintDefaults()
			return false
		}
		hub.Config.Port = val
	}

	if hub.Config.Scheme == "REQUIRED" {
		val := os.Getenv("BDS_SCHEME")
		if val == "" {
			log.Println("-s scheme argument or BDS_SCHEME environment is required")
			pflag.PrintDefaults()
			return false
		}
		hub.Config.Scheme = val
	}

	if hub.Config.User == "REQUIRED" {
		val := os.Getenv("BDS_USER")
		if val == "" {
			log.Println("-u username argument or BDS_USER environment is required")
			pflag.PrintDefaults()
			return false
		}
		hub.Config.User = val
	}

	if hub.Config.Password == "REQUIRED" {
		val := os.Getenv("BDS_PASSWORD")
		if val == "" {
			log.Println("-w password argument or BDS_PASSWORD environment is required")
			pflag.PrintDefaults()
			return false
		}
		hub.Config.Password = val
	}

	if hub.Scanner == "REQUIRED" {
		val := os.Getenv("BDS_SCANNER")
		if val == "" {
			log.Println("-scanner argument or BDS_SCANNER environment is required")
			pflag.PrintDefaults()
			return false
		}
		hub.Scanner = val
	}

	if hub.Workers < 1 {
		val := os.Getenv("BDS_WORKERS")
		number, _ := strconv.Atoi(val)
		if number < 1 {
			log.Printf("Setting workers from %d to %d\n", number, controller.MaxWorkers)
			hub.Workers = controller.MaxWorkers
		}
		hub.Workers = number
	}

	hub.Config.Url = fmt.Sprintf("%s://%s", hub.Config.Scheme, hub.Config.Host)

	return true
}
