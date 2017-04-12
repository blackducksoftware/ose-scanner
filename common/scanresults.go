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

package common

import (

	"errors"
	"fmt"
	"log"
	"strings"
	"time"

)

func ScanResults(info ImageInfo, taggedName string, imageId string, scanId string, sha string, annotate *Annotator, hubConfig *HubConfig ) (e error, results ImageInfo) {
	log.Printf("Checking for vulnerabilities on: %s\n", taggedName)

	hub := HubServer{Config: hubConfig}
	if ok := hub.Login(); !ok {
		log.Printf("Hub credentials not valid\n")
		return errors.New("Invalid Hub credentials"), info
	}

	scanSummary, ok := hub.GetScanSummary(scanId)
	if !ok {
		e := fmt.Sprintf("ERROR no scan summary for image: %s", imageId)
		log.Printf("%s\n", e)
		return errors.New(e), info
	}

	for strings.Compare(scanSummary.Status, "COMPLETE") != 0 {
		time.Sleep(1 * time.Minute)
		scanSummary, ok = hub.GetScanSummary(scanId)
		if !ok {
			// someone deleted our codelocation underneath us
			e := fmt.Sprintf("ERROR processing scan summary for image %s. No items returned. Was code location modified?", imageId)
			log.Printf("%s\n", e)
			return errors.New(e), info
		}

		log.Printf("Scan status: %s\n", scanSummary.Status)

		if strings.Compare(scanSummary.Status, "ERROR") == 0 {
			e := fmt.Sprintf("ERROR processing scan summary for image: %s", imageId)
			log.Printf("%s\n", e)
			return errors.New(e), info
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
		e := fmt.Sprintf("ERROR unable to locate code location URL for image: %s", imageId)
		log.Printf("%s\n", e)
		return errors.New(e), info
	}

	codeLocation, ok := hub.GetCodeLocation(codeLocationUrl)
	if !ok {
		e := fmt.Sprintf("ERROR no code location found for image: %s", imageId)
		log.Printf("%s\n", e)
		return errors.New(e), info
	}

	projectVersionUrl := codeLocation.MappedProjectVersion
	projectVersion, ok := hub.GetProjectVersion(projectVersionUrl)
	if !ok {
		e := fmt.Sprintf("ERROR no project version found for image: %s", imageId)
		log.Printf("%s\n", e)
		return errors.New(e), info
	}

	vulnerabilities := 0
	violations := 0

	// we have a matching version for our image, need to locate the risk-profile
	for _, Item := range projectVersion.Meta.Links {

		log.Printf("  Processing version link: %s\n", Item.Rel)
		if strings.Compare(Item.Rel, "riskProfile") == 0 {
			riskProfile, ok := hub.GetRiskProfile(Item.Href)
			if riskProfile == nil || !ok {
				e := fmt.Sprintf("ERROR unable to load risk profile for image: %s:%s", taggedName, imageId[:10])
				log.Printf("%s\n", e)
				return errors.New(e), info
			}
			vulnerabilities = riskProfile.Categories.VULNERABILITY.HIGH
		}

		if strings.Compare(Item.Rel, "policy-status") == 0 {
			policyStatus, ok := hub.GetPolicyStatus(Item.Href)
			if policyStatus == nil || !ok {
				e := fmt.Sprintf("ERROR unable to load policy status for image: %s:%s", taggedName, imageId[:10])
				log.Printf("%s\n", e)
				return errors.New(e), info
			}
			for _, PolicyItem := range policyStatus.ComponentVersionStatusCounts {
				if strings.Compare(PolicyItem.Name, "IN_VIOLATION") == 0 {
					violations = PolicyItem.Value
				}
			}

		}
	}

	log.Printf("Found %d high severity vulnerabilities and %d policy violations for %s:%s\n", vulnerabilities, violations, taggedName, imageId[:10])

	results = annotate.UpdateAnnotations(info, sha, violations, vulnerabilities, projectVersionUrl, scanId)

	return nil, results
}

func ProjectVersionResults (info ImageInfo, imageId string, taggedName string, sha string, scanId string, projectVersionUrl string, hub *HubServer, annotate *Annotator ) (e error, results ImageInfo) {
	log.Printf("Processing vulnerabilities and policy violations for %s:%s\n", taggedName, imageId[:10])

	projectVersion, ok := hub.GetProjectVersion(projectVersionUrl)
	if !ok {
		e := fmt.Sprintf("ERROR no project version found for image: %s", imageId)
		log.Printf("%s\n", e)
		return errors.New(e), info
	}

	vulnerabilities := 0
	violations := 0

	// we have a matching version for our image, need to locate the risk-profile
	for _, Item := range projectVersion.Meta.Links {

		log.Printf("  Processing version link: %s\n", Item.Rel)
		if strings.Compare(Item.Rel, "riskProfile") == 0 {
			riskProfile, ok := hub.GetRiskProfile(Item.Href)
			if riskProfile == nil || !ok {
				e := fmt.Sprintf("ERROR unable to load risk profile for image: %s:%s", taggedName, imageId[:10])
				log.Printf("%s\n", e)
				return errors.New(e), info
			}
			vulnerabilities = riskProfile.Categories.VULNERABILITY.HIGH
		}

		if strings.Compare(Item.Rel, "policy-status") == 0 {
			policyStatus, ok := hub.GetPolicyStatus(Item.Href)
			if policyStatus == nil || !ok {
				e := fmt.Sprintf("ERROR unable to load policy status for image: %s:%s", taggedName, imageId[:10])
				log.Printf("%s\n", e)
				return errors.New(e), info
			}
			for _, PolicyItem := range policyStatus.ComponentVersionStatusCounts {
				if strings.Compare(PolicyItem.Name, "IN_VIOLATION") == 0 {
					violations = PolicyItem.Value
				}
			}

		}
	}

	log.Printf("Found %d high severity vulnerabilities and %d policy violations for %s:%s\n", vulnerabilities, violations, taggedName, imageId[:10])

	results = annotate.UpdateAnnotations(info, sha, violations, vulnerabilities, projectVersionUrl, scanId)

	return nil, results
}