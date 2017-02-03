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
	"github.com/fsouza/go-dockerclient"
	"io"
	"log"
	"strings"
	"time"
)

type Docker struct {
	client  *docker.Client
	shortID string
}

func (d Docker) launchContainer(scanner string, args []string) (bool, error) {

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
		return false, err
	}

	d.shortID = container.ID[:10]

	done := make(chan bool)
	abort := make(chan bool, 1)
	d.pipeOutput(container.ID, done, abort)

	err = d.client.StartContainer(container.ID, &docker.HostConfig{Privileged: true})

	if err != nil {
		log.Printf("Error starting container ID %s for %s: %s\n", d.shortID, scanner, err)
		return false, err
	}

	log.Printf("Started scan container %s\n", d.shortID)

	exit, err := d.client.WaitContainer(container.ID) // block until done (logs in pipeOutput)

	if err != nil {
		log.Printf("Error waiting container ID %s with exit %d: %s\n", d.shortID, exit, err)
		return false, err
	} else {
		log.Printf("Scan container %s exit with status %d\n", d.shortID, exit)
	}

	options := docker.RemoveContainerOptions{
		ID:            container.ID,
		RemoveVolumes: true,
	}

	err = d.client.RemoveContainer(options)

	if err != nil {
		log.Printf("Error removing container ID %s: %s\n", d.shortID, err)
		return false, err
	}

	abort <- true
	result := <-done
	log.Printf("Scan complete of %s with result %t\n", d.shortID, result)

	return result, nil
}

func (d Docker) pipeOutput(ID string, done chan bool, abort chan bool) error {
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

	go func(reader *io.PipeReader, shortID string, a chan bool) {

		for {
			time.Sleep(time.Second)
			select {
			case _ = <-a:
				log.Printf("Received IO shutdown for scanner %s\n", shortID)
				reader.Close()
				return

			default:
			}

		}

	}(r, d.shortID, abort)

	go func(reader io.Reader, shortID string, c chan bool) {
		scanner := bufio.NewScanner(reader)
		scan := false

		for scanner.Scan() {
			out := scanner.Text()
			if strings.Contains(out, "Post Scan...") {
				log.Printf("Found completed scan with result for %s\n", shortID)
				scan = true
			}
			log.Printf("%s: %s\n", shortID, out)
		}

		/*
			*** since we exit the loop by closing the reader, an error will occur ***
			*** Need to look into a better way such that legit errors can be captured ***
			if err := scanner.Err(); err != nil {
				log.Printf("Scanner error on %s: %s\n", shortID, err)
			} */

		log.Printf("Placing scan result %t into channel for %s\n", scan, shortID)
		c <- scan

	}(r, d.shortID, done)

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
