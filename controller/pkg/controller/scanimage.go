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
	"strings"
)

type ScanImage struct {
	imageId    string
	taggedName string
	digest     string
	scanned    bool
}

func NewScanImage(ID string, Reference string) *ScanImage {

	tag := strings.Split(Reference, "@")

	return &ScanImage{
		imageId:    ID,
		taggedName: tag[0],
		digest:     Reference,
		scanned:    false,
	}
}

func (image ScanImage) scan() (e error) {

	log.Printf("Scanning: %s (%s)\n", image.taggedName, image.imageId[:10])

	args := []string{}
	args = append(args, "/ose_scanner")

	args = append(args, "-h")
	args = append(args, Hub.Host)

	args = append(args, "-p")
	args = append(args, Hub.Port)

	args = append(args, "-s")
	args = append(args, Hub.Scheme)

	args = append(args, "-u")
	args = append(args, Hub.Username)

	args = append(args, "-w")
	args = append(args, Hub.Password)

	args = append(args, "-id")
	args = append(args, image.imageId)

	args = append(args, "-tag")
	args = append(args, image.taggedName)

	args = append(args, "-digest")
	args = append(args, image.digest)

	docker := NewDocker()

	err := docker.launchContainer(Hub.Scanner, args)

	if err != nil {
		log.Printf("Error creating scanning container: %s\n", err)
		return err
	}

	log.Printf("Done Scanning: %s\n", image.taggedName)

	image.scanned = true

	return nil
}
