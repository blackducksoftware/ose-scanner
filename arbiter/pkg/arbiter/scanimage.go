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
	"fmt"
	"log"
	"strings"
	"time"
)

type ScanImage struct {
	imageId    string
	taggedName string
	sha        string
	digest     string
	scanned    bool
	annotate   *Annotator
}

func newScanImage(ID string, Reference string, annotate *Annotator) *ScanImage {

	tag := strings.Split(Reference, "@")

	return &ScanImage{
		imageId:    ID,
		taggedName: tag[0],
		sha:        tag[1],
		digest:     Reference,
		scanned:    false,
		annotate:   annotate,
	}
}

func (image ScanImage) results() (e error) {
	log.Printf("Checking for vulnerabilities on: %s\n", image.taggedName)

	hub := HubServer{Config: Hub.Config}
	if ok := hub.login(); !ok {
		log.Printf("Hub credentials not valid\n")
		return errors.New("Invalid Hub credentials")
	}

	// check if the scan was completed
	codeLocations := hub.findCodeLocations(image.imageId)
	if len(codeLocations.Items) == 0 {
		e := fmt.Sprintf("ERROR no code locations for image: %s", image.imageId)
		log.Printf("%s\n", e)
		return errors.New(e)
	}

	scanSummaryUrl := codeLocations.Items[codeLocations.TotalCount-1].Meta.Links[0].Href

	scanSummaries := hub.findCodeLocationScanSummaries(scanSummaryUrl)
	if len(scanSummaries.Items) != 1 {
		e := fmt.Sprintf("ERROR no scan summary for image: %s", image.imageId)
		log.Printf("%s\n", e)
		return errors.New(e)
	}

	for strings.Compare(scanSummaries.Items[0].Status, "COMPLETE") != 0 {
		time.Sleep(1 * time.Minute)
		scanSummaries = hub.findCodeLocationScanSummaries(scanSummaryUrl)
		log.Printf("Scan status: %s\n", scanSummaries.Items[0].Status)

		if strings.Compare(scanSummaries.Items[0].Status, "ERROR") == 0 {
			e := fmt.Sprintf("ERROR processing scan summaries for image: %s", image.imageId)
			log.Printf("%s\n", e)
			return errors.New(e)
		}
	}

	projects := hub.findProjects(image.taggedName)
	if len(projects.Items) != 1 {
		e := fmt.Sprintf("ERROR no projects summary for image: %s", image.taggedName)
		log.Printf("%s\n", e)
		return errors.New(e)
	}

	index := strings.LastIndex(projects.Items[0].Meta.Href, "/")
	projectId := projects.Items[0].Meta.Href[index+1:]

	projectVersions := hub.findProjectVersions(projectId, image.imageId[:10])
	if len(projectVersions.Items) != 1 {
		e := fmt.Sprintf("ERROR unable to locate project version for image: %s:%s", image.taggedName, image.imageId[:10])
		log.Printf("%s\n", e)
		return errors.New(e)
	}

	vulnerabilities := 0

	// we have a matching version for our image, need to locate the risk-profile
	for _, Item := range projectVersions.Items[0].Meta.Links {

		log.Printf("  Processing version link: %s\n", Item.Rel)
		if strings.Compare(Item.Rel, "riskProfile") == 0 {

			riskProfile := hub.getRiskProfile(Item.Href)
			if riskProfile == nil {
				e := fmt.Sprintf("ERROR unable to load risk profile for image: %s:%s", image.taggedName, image.imageId[:10])
				log.Printf("%s\n", e)
				return errors.New(e)
			}
			vulnerabilities = riskProfile.Categories.VULNERABILITY.HIGH

			break
		}
	}

	log.Printf("Found %d high severity vulnerabilities for %s:%s\n", vulnerabilities, image.taggedName, image.imageId[:10])

	saved := image.annotate.SaveResults(image.sha, vulnerabilities, projectId)

	if !saved {
		log.Printf("Unable to annotate image with results %s\n", image.imageId)
	}

	//TODO - Hook results into AdmissionController
	return nil
}
