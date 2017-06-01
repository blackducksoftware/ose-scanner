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
	"errors"
	"log"
	"strings"

	bdscommon "github.com/blackducksoftware/ose-scanner/common"
)

type ScanImage struct {
	imageId    string
	taggedName string
	sha        string
	digest     string
	scanned    bool
	scanId     string
	annotate   *bdscommon.Annotator
	config *bdscommon.HubConfig
	scanner string
}

func newScanImage(ID string, Reference string, annotate *bdscommon.Annotator, config *bdscommon.HubConfig, scanner string) *ScanImage {

	tag := strings.Split(Reference, "@")

	return &ScanImage{
		imageId:    ID,
		taggedName: tag[0],
		sha:        tag[1],
		digest:     Reference,
		scanned:    false,
		annotate:   annotate,
		config: config,
		scanner: scanner,
	}
}

func (image ScanImage) getArgs() []string {

	args := []string{}

	args = append(args, "/ose_scanner")

	args = append(args, "-h")
	args = append(args, image.config.Host)

	args = append(args, "-p")
	args = append(args, image.config.Port)

	args = append(args, "-s")
	args = append(args, image.config.Scheme)

	args = append(args, "-u")
	args = append(args, image.config.User)

	args = append(args, "-w")
	args = append(args, image.config.Password)

	args = append(args, "-id")
	args = append(args, image.imageId)

	args = append(args, "-tag")
	args = append(args, image.taggedName)

	args = append(args, "-digest")
	args = append(args, image.digest)

	return args

}

func (image ScanImage) scan(info bdscommon.ImageInfo) (error, bdscommon.ImageInfo) {

	log.Printf("Scanning: %s (%s)\n", image.taggedName, image.imageId[:10])

	docker := NewDocker()
	if docker.client == nil {
		log.Printf("No Docker client connection\n")
		return errors.New("Invalid Docker connection"), info
	}

	if !docker.imageExists(image.digest) {
		log.Printf("Image %s does not exist\n", image.digest)
		return errors.New("Image does not exist"), info
	}

	args := image.getArgs()

	goodScan, err := docker.launchContainer(image.scanner, args)

	if err != nil {
		log.Printf("Error creating scanning container: %s\n", err)
		return err, info
	}

	log.Printf("Done Scanning: %s (%s) with result %t using scanId %s\n", image.taggedName, image.imageId[:10], goodScan.completed, goodScan.scanId)

	image.scanned = true
	image.scanId = goodScan.scanId

	if goodScan.completed {
		return image.results(info)
	}

	return nil, info
}

func (image ScanImage) results(info bdscommon.ImageInfo) (error, bdscommon.ImageInfo) {
	return bdscommon.ScanResults(info, image.taggedName, image.imageId, image.scanId, image.sha, image.annotate, image.config)
}

func (image ScanImage) exists() bool {

	log.Printf("Scanning: %s (%s)\n", image.taggedName, image.imageId[:10])

	docker := NewDocker()
	if docker.client == nil {
		log.Printf("No Docker client connection\n")
		return false
	}

	if !docker.imageExists(image.digest) {
		log.Printf("Image %s does not exist\n", image.digest)
		return false
	}

	return true

}
