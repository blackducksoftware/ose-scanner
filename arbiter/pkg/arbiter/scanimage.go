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

package arbiter

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
	annotate   *bdscommon.Annotator
}

func newScanImage(ID string, Reference string, annotate *bdscommon.Annotator) *ScanImage {

	tag := strings.Split(Reference, "@")
	Ids := strings.Split(ID, "sha256:")

	return &ScanImage{
		imageId:    Ids[len(Ids)-1],
		taggedName: tag[0],
		sha:        tag[1],
		digest:     Reference,
		scanned:    false,
		annotate:   annotate,
	}
}

func (image ScanImage) scanResults(info bdscommon.ImageInfo) (error, bdscommon.ImageInfo) {

	scanId, _ := info.Annotations[bdscommon.ScannerScanId]
	if len(scanId) == 0 {
		return errors.New("No scan ID found"), info
	}

	return bdscommon.ScanResults(info, image.taggedName, image.imageId, scanId, image.sha, image.annotate, Hub.Config)
}

func (image ScanImage) versionResults(info bdscommon.ImageInfo) (error, bdscommon.ImageInfo) {
	scanId, _ := info.Annotations[bdscommon.ScannerScanId]
	projectVersionUrl, _ := info.Annotations[bdscommon.ScannerProjectVersionUrl]

	if len(scanId) == 0 || len(projectVersionUrl) == 0 {
		return errors.New("Missing project information"), info
	}

	hub := bdscommon.NewHubServer(Hub.Config)

	if ok := hub.Login(); !ok {
		log.Printf("Hub credentials not valid\n")
		return errors.New("Invalid Hub credentials"), info
	}

	return bdscommon.ProjectVersionResults(info, image.imageId, image.taggedName, image.sha, scanId, projectVersionUrl, hub, image.annotate)
}
