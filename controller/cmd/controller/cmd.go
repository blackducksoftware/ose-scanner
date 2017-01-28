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

	if !c.ValidateConfig () {
		log.Printf ("Hub configuation information isn't valid. Please verify connectivity and values.")
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

	hub.Config = &controller.HubConfig {}
	
	pflag.StringVar(&hub.Config.Host, "h", "REQUIRED", "The hostname of the Black Duck Hub server.")
	pflag.StringVar(&hub.Config.Port, "p", "REQUIRED", "The port the Hub is communicating on")
	pflag.StringVar(&hub.Config.Scheme, "s", "REQUIRED", "The communication scheme [http,https].")
	pflag.StringVar(&hub.Config.User, "u", "REQUIRED", "The Black Duck Hub user")
	pflag.StringVar(&hub.Config.Password, "w", "REQUIRED", "Password for the user.")
	pflag.StringVar(&hub.Scanner, "scanner", "REQUIRED", "Scanner image")
	pflag.IntVar(&hub.Workers, "workers", controller.MaxWorkers, "Number of container workers")
}

func checkExpectedCmdlineParams() bool {
	// NOTE: At this point we don't have a logger yet, so don't try and use it.

	if hub.Config.Host == "REQUIRED" {
		log.Println("-h host is required")
		pflag.PrintDefaults()
		return false
	}

	if hub.Config.Port == "REQUIRED" {
		log.Println("-p port is required")
		pflag.PrintDefaults()
		return false
	}

	if hub.Config.Scheme == "REQUIRED" {
		log.Println("-s scheme is required")
		pflag.PrintDefaults()
		return false
	}

	if hub.Config.User == "REQUIRED" {
		log.Println("-u username is required")
		pflag.PrintDefaults()
		return false
	}

	if hub.Config.Password == "REQUIRED" {
		log.Println("-w password is required")
		pflag.PrintDefaults()
		return false
	}

	if hub.Scanner == "REQUIRED" {
		log.Println("-scanner Hub scanner image is required")
		pflag.PrintDefaults()
		return false
	}

	if hub.Workers < 1 {
		log.Printf("Setting workers from %d to %d\n", hub.Workers, controller.MaxWorkers)
		hub.Workers = controller.MaxWorkers
	}

	hub.Config.Url = fmt.Sprintf("%s://%s", hub.Config.Scheme, hub.Config.Host)

	return true
}

