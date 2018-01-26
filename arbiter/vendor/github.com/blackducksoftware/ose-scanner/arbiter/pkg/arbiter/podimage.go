/*
Copyright (C) 2017 Black Duck Software, Inc.
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

type PodImage struct {
	imageName string
	podInfo   []PodInfo
	annotate  *bdscommon.Annotator
}

func newPodImage(name string, info []PodInfo, annotate *bdscommon.Annotator) *PodImage {

	return &PodImage{
		imageName: name,
		podInfo:   info,
		annotate:  annotate,
	}
}

// ScanResults returns the results of a prior scan from our source of truth
func (p PodImage) ScanResults() (error, bdscommon.ImageInfo) {

	dummyInfo := bdscommon.ImageInfo{
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}

	components := strings.Split(p.imageName, "://")
	if len(components) != 2 {
		log.Printf("Invalid component %s\n", p.imageName)
		return errors.New("Invalid image name"), dummyInfo
	}

	path := components[1]
	el := strings.Split(path, "@sha256:")
	if len(el) != 2 {
		log.Printf("Invalid pull spec %s\n", path)
		return errors.New("Invalid pull spec"), dummyInfo
	}

	project := el[0]
	version := el[1][:10]

	return bdscommon.GetScanResultsFromProjectVersion(project, version, p.annotate, Hub.Config)
}
