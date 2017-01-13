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
	"bufio"
	"io"
	"log"

	"github.com/fsouza/go-dockerclient"
)

type Docker struct {
	client  *docker.Client
	shortID string
}

func (d Docker) launchContainer(scanner string, args []string) (e error) {

	binds := []string{}
	binds = append(binds, "/var/run/docker.sock:/var/run/docker.sock")

	container, err := d.client.CreateContainer(
		docker.CreateContainerOptions{
			Config: &docker.Config{
				Image:        scanner,
				AttachStdout: true,
				AttachStderr: true,
				Tty:          true,
				Entrypoint:   args,
			},
			HostConfig: &docker.HostConfig{
				Privileged: true,
				Binds:      binds,
			},
		})

	if err != nil {
		log.Printf("Error creating container %s: %s\n", scanner, err)
		return err
	}

	d.shortID = container.ID[:10]

	//time.Sleep (2 * time.Second)

	d.pipeOutput(container.ID)

	err = d.client.StartContainer(container.ID, &docker.HostConfig{Privileged: true})

	if err != nil {
		log.Printf("Error starting container ID %s for %s: %s\n", d.shortID, scanner, err)
		return err
	}

	log.Printf("Started scan container %s\n", d.shortID)

	exit, err := d.client.WaitContainer(container.ID) // block until done (logs in pipeOutput)

	if err != nil {
		log.Printf("Error waiting container ID %s with exit %d: %s\n", d.shortID, exit, err)
		return err
	} else {
		log.Printf("Scan container %s exit with status %d\n", d.shortID, exit)
	}

	//time.Sleep (5 * time.Second)

	options := docker.RemoveContainerOptions{
		ID:            container.ID,
		RemoveVolumes: true,
	}

	err = d.client.RemoveContainer(options)

	if err != nil {
		log.Printf("Error removing container ID %s: %s\n", d.shortID, err)
		return err
	}

	return nil
}

func (d Docker) pipeOutput(ID string) error {
	r, w := io.Pipe()

	options := docker.AttachToContainerOptions{
		Container:    ID,
		OutputStream: w,
		ErrorStream:  w,
		Stream:       true,
		Stdout:       true,
		Stderr:       true,
		Logs:         true,
		RawTerminal:  true,
	}

	log.Printf("Attaching to IO streams on %s\n", d.shortID)

	go d.client.AttachToContainer(options) // will block so isolate

	go func(reader io.Reader) {
		scanner := bufio.NewScanner(reader)

		for scanner.Scan() {
			log.Printf("%s: %s \n", d.shortID, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			log.Printf("Scanner error on %s: %s\n", d.shortID, err)
		}

	}(r)
	return nil
}

func NewDocker() Docker {

	endpoint := "unix:///var/run/docker.sock"
	client, err := docker.NewClient(endpoint)
	if err != nil {
		log.Printf("Error connecting to docker engine %s\n", err)
	}

	return Docker{
		client: client,
	}
}
