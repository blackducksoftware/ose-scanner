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

type ScanResult struct {
	completed bool
	scanId    string
}

type Docker struct {
	client  *docker.Client
	shortID string
}

func (d Docker) imageExists(image string) bool {
	imageDetails, err := d.client.InspectImage(image)
	if err != nil {
		log.Printf("Error testing if image %s exists: %s\n", image, err)
		return false
	}

	log.Printf("Image %s exists with id %s\n", image, imageDetails.ID)
	return true

}

func (d Docker) loadScanner (scanner string) (error) {
	if d.imageExists (scanner) {
		log.Println("Scanner found")
		return nil
	}

	loc := strings.LastIndex (scanner, ":") // we know there is always a : to sep the tag because we put it there

	image := scanner[:loc]
	tag := scanner[loc+1:]


	opts := docker.PullImageOptions{
	        Repository: image,
		Tag: tag,
    	}

	auth := docker.AuthConfiguration{}

	log.Printf ("Attempting to pull scanner [%s] [%s]\n", image, tag)

	err := d.client.PullImage(opts, auth)

	if err != nil {
		log.Printf("Error pulling image %s: %s\n", scanner, err)
		return err
	}

	return nil
}

func (d Docker) launchContainer(scanner string, args []string) (ScanResult, error) {

	emptyScanResult := ScanResult{completed: false, scanId: ""}

	err := d.loadScanner(scanner)
	if err != nil {
		log.Printf("Error loading scan container %s: %s\n", scanner, err)
		return emptyScanResult, err
	}

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
		log.Printf("Error creating scan container %s: %s\n", scanner, err)
		return emptyScanResult, err
	}

	d.shortID = container.ID[:10]

	done := make(chan ScanResult)
	abort := make(chan bool, 1)
	d.pipeOutput(container.ID, done, abort)

	err = d.client.StartContainer(container.ID, &docker.HostConfig{Privileged: true})

	if err != nil {
		log.Printf("Error starting scan container ID %s for %s: %s\n", d.shortID, scanner, err)
		return emptyScanResult, err
	}

	log.Printf("Started scan container %s\n", d.shortID)

	exit, err := d.client.WaitContainer(container.ID) // block until done (logs in pipeOutput)

	if err != nil {
		log.Printf("Error waiting scan container ID %s with exit %d: %s\n", d.shortID, exit, err)
		return emptyScanResult, err
	} else {
		log.Printf("Scan container %s exit with status %d\n", d.shortID, exit)
	}

	options := docker.RemoveContainerOptions{
		ID:            container.ID,
		RemoveVolumes: true,
	}

	err = d.client.RemoveContainer(options)

	if err != nil {
		log.Printf("Error removing scan container ID %s: %s\n", d.shortID, err)
		return emptyScanResult, err
	}

	abort <- true
	result := <-done
	log.Printf("Scan complete of %s with result %t\n", d.shortID, result.completed)

	return result, nil
}

func (d Docker) pipeOutput(ID string, done chan ScanResult, abort chan bool) error {
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

	go func(reader io.Reader, shortID string, c chan ScanResult) {
		scanner := bufio.NewScanner(reader)
		scan := ScanResult{completed: false, scanId: ""}

		for scanner.Scan() {
			out := scanner.Text()
			if strings.Contains(out, "Post Scan...") {
				log.Printf("Found completed scan with result for %s\n", shortID)
				scan.completed = true
			}
			log.Printf("%s: %s\n", shortID, out)

			if strings.Contains(out, "ScanContainerView{scanId=") {
				cmd := strings.Split(out, "ScanContainerView{scanId=")
				eos := strings.Index(cmd[1], ",")
				scan.scanId = cmd[1][:eos]
				log.Printf("Found scan ID %s with result for %s\n", scan.scanId, shortID)
			}
		}

		/*
			*** since we exit the loop by closing the reader, an error will occur ***
			*** Need to look into a better way such that legit errors can be captured ***
			if err := scanner.Err(); err != nil {
				log.Printf("Scanner error on %s: %s\n", shortID, err)
			} */

		log.Printf("Placing scan result %t from scanId %s into channel for %s\n", scan.completed, scan.scanId, shortID)
		c <- scan

	}(r, d.shortID, done)

	return nil
}

func NewDocker() Docker {

	endpoint := "unix:///var/run/docker.sock"
	client, err := docker.NewVersionedClient(endpoint, "1.22")
	if err != nil {
		log.Printf("Error connecting to docker engine %s\n", err)
	}

	return Docker{
		client: client,
	}
}
