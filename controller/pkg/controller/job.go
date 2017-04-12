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
	bdscommon "github.com/blackducksoftware/ose-scanner/common"
)

type Job struct {
	ScanImage  *ScanImage
	controller *Controller
}

func (job Job) Done() {
	job.controller.wait.Done()
	return
}

func (job Job) Load() {
	job.controller.wait.Add(1)
	log.Println("Queue image: " + job.ScanImage.taggedName)
	return
}

func (job Job) GetAnnotationInfo() (result bool, info bdscommon.ImageInfo) {
	image, err := job.controller.openshiftClient.Images().Get(job.ScanImage.sha)
	if err != nil {
		log.Printf("Job: Error getting image %s: %s\n", job.ScanImage.sha, err)
		return false, info
	}

	info.Annotations = image.ObjectMeta.Annotations
	if info.Annotations == nil {
		log.Printf("Image %s has no annotations - creating.\n", job.ScanImage.sha)
		info.Annotations = make(map[string]string)
	}

	info.Labels = image.ObjectMeta.Labels
	if info.Labels == nil {
		log.Printf("Image %s has no labels - creating.\n", job.ScanImage.sha)
		info.Labels = make(map[string]string)
	}

	return true, info
}

func (job Job) UpdateAnnotationInfo(info bdscommon.ImageInfo) bool {
	image, err := job.controller.openshiftClient.Images().Get(job.ScanImage.sha)
	if err != nil {
		log.Printf("Job: Error getting image %s: %s\n", job.ScanImage.sha, err)
		return false
	}

	image.ObjectMeta.Annotations = info.Annotations

	image.ObjectMeta.Labels = info.Labels

	image, err = job.controller.openshiftClient.Images().Update(image)
	if err != nil {
		log.Printf("Error updating image: %s. %s\n", job.ScanImage.sha, err)
		return false
	}

	return true
}
