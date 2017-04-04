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
	scanId     string
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

func (image ScanImage) getArgs() []string {

	args := []string{}

	args = append(args, "/ose_scanner")

	args = append(args, "-h")
	args = append(args, Hub.Config.Host)

	args = append(args, "-p")
	args = append(args, Hub.Config.Port)

	args = append(args, "-s")
	args = append(args, Hub.Config.Scheme)

	args = append(args, "-u")
	args = append(args, Hub.Config.User)

	args = append(args, "-w")
	args = append(args, Hub.Config.Password)

	args = append(args, "-id")
	args = append(args, image.imageId)

	args = append(args, "-tag")
	args = append(args, image.taggedName)

	args = append(args, "-digest")
	args = append(args, image.digest)

	return args

}

func (image ScanImage) scan() (e error) {

	log.Printf("Scanning: %s (%s)\n", image.taggedName, image.imageId[:10])

	docker := NewDocker()
	if docker.client == nil {
		log.Printf("No Docker client connection\n")
		return errors.New("Invalid Docker connection")
	}

	if !docker.imageExists(image.digest) {
		log.Printf("Image %s does not exist\n", image.digest)
		return errors.New("Image does not exist")
	}

	args := image.getArgs()

	goodScan, err := docker.launchContainer(Hub.Scanner, args)

	if err != nil {
		log.Printf("Error creating scanning container: %s\n", err)
		return err
	}

	log.Printf("Done Scanning: %s (%s) with result %t using scanId %s\n", image.taggedName, image.imageId[:10], goodScan.completed, goodScan.scanId)

	image.scanned = true
	image.scanId = goodScan.scanId

	if goodScan.completed {
		return image.results()
	}

	return nil
}

func (image ScanImage) results() (e error) {
	log.Printf("Checking for vulnerabilities on: %s\n", image.taggedName)

	hub := HubServer{Config: Hub.Config}
	if ok := hub.login(); !ok {
		log.Printf("Hub credentials not valid\n")
		return errors.New("Invalid Hub credentials")
	}

	scanSummary, ok := hub.getScanSummary(image.scanId)
	if !ok {
		e := fmt.Sprintf("ERROR no scan summary for image: %s", image.imageId)
		log.Printf("%s\n", e)
		return errors.New(e)
	}

	for strings.Compare(scanSummary.Status, "COMPLETE") != 0 {
		time.Sleep(1 * time.Minute)
		scanSummary, ok = hub.getScanSummary(image.scanId)
		if !ok {
			// someone deleted our codelocation underneath us
			e := fmt.Sprintf("ERROR processing scan summary for image %s. No items returned. Was code location modified?", image.imageId)
			log.Printf("%s\n", e)
			return errors.New(e)
		}

		log.Printf("Scan status: %s\n", scanSummary.Status)

		if strings.Compare(scanSummary.Status, "ERROR") == 0 {
			e := fmt.Sprintf("ERROR processing scan summary for image: %s", image.imageId)
			log.Printf("%s\n", e)
			return errors.New(e)
		}
	}

	codeLocationUrl := ""
	for _, Item := range scanSummary.Meta.Links {

		log.Printf("  Processing scan summary link: %s\n", Item.Rel)
		if strings.Compare(Item.Rel, "codelocation") == 0 {
			codeLocationUrl = Item.Href
			break
		}
	}

	if len(codeLocationUrl) == 0 {
		e := fmt.Sprintf("ERROR unable to locate code location URL for image: %s", image.imageId)
		log.Printf("%s\n", e)
		return errors.New(e)
	}

	codeLocation, ok := hub.getCodeLocation(codeLocationUrl)
	if !ok {
		e := fmt.Sprintf("ERROR no code location found for image: %s", image.imageId)
		log.Printf("%s\n", e)
		return errors.New(e)
	}

	projectVersionUrl := codeLocation.MappedProjectVersion
	projectVersion, ok := hub.getProjectVersion(projectVersionUrl)
	if !ok {
		e := fmt.Sprintf("ERROR no project version found for image: %s", image.imageId)
		log.Printf("%s\n", e)
		return errors.New(e)
	}

	vulnerabilities := 0
	violations := 0

	// we have a matching version for our image, need to locate the risk-profile
	for _, Item := range projectVersion.Meta.Links {

		log.Printf("  Processing version link: %s\n", Item.Rel)
		if strings.Compare(Item.Rel, "riskProfile") == 0 {
			riskProfile, ok := hub.getRiskProfile(Item.Href)
			if riskProfile == nil || !ok {
				e := fmt.Sprintf("ERROR unable to load risk profile for image: %s:%s", image.taggedName, image.imageId[:10])
				log.Printf("%s\n", e)
				return errors.New(e)
			}
			vulnerabilities = riskProfile.Categories.VULNERABILITY.HIGH
		}

		if strings.Compare(Item.Rel, "policy-status") == 0 {
			policyStatus, ok := hub.getPolicyStatus(Item.Href)
			if policyStatus == nil || !ok {
				e := fmt.Sprintf("ERROR unable to load policy status for image: %s:%s", image.taggedName, image.imageId[:10])
				log.Printf("%s\n", e)
				return errors.New(e)
			}
			for _, PolicyItem := range policyStatus.ComponentVersionStatusCounts {
				if strings.Compare(PolicyItem.Name, "IN_VIOLATION") == 0 {
					violations = PolicyItem.Value
				}
			}

		}
	}

	log.Printf("Found %d high severity vulnerabilities and %d policy violations for %s:%s\n", vulnerabilities, violations, image.taggedName, image.imageId[:10])

	saved := image.annotate.SaveResults(image.sha, violations, vulnerabilities, projectVersionUrl, image.scanId)

	if !saved {
		log.Printf("Unable to annotate image with results %s\n", image.imageId)
	}

	return nil


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
