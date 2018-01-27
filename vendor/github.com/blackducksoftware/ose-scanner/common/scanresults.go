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

func ScanResults(taggedName string, imageId string, scanId string, sha string, annotate *Annotator, hubConfig *HubConfig) (e error, results ImageInfo) {
	log.Printf("Checking for vulnerabilities on: %s\n", taggedName)

	info := ImageInfo{
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}

	hub := NewHubServer(hubConfig)
	if ok := hub.Login(); !ok {
		log.Printf("Hub credentials not valid\n")
		return errors.New("Invalid Hub credentials"), info
	}

	defer hub.Logout()

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

		switch scanSummary.Status {
		case "ERROR", "ERROR_BUILDING_BOM", "ERROR_MATCHING", "ERROR_SAVING_SCAN_DATA", "ERROR_SCANNING", "CANCELLED":
			e := fmt.Sprintf("%s processing scan summary for image: %s", scanSummary.Status, imageId)
			log.Printf("%s\n", e)
			return errors.New(e), info
		default:
			log.Printf("Scan status: %s\n", scanSummary.Status)

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
	projectVersionUI := ""

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

		if strings.Compare(Item.Rel, "components") == 0 {
			projectVersionUI = Item.Href
		}

	}

	log.Printf("Found %d high severity vulnerabilities and %d policy violations for %s:%s\n", vulnerabilities, violations, taggedName, imageId[:10])

	results = annotate.UpdateAnnotations(info, violations, vulnerabilities, projectVersionUrl, scanId, projectVersionUI)

	return nil, results
}

func ValidateGetProjectVersion(projectVersionUrl string, hubConfig *HubConfig) bool {

	hub := NewHubServer(hubConfig)
	if ok := hub.Login(); !ok {
		log.Printf("Hub credentials not valid during project version check\n")
		return false
	}

	defer hub.Logout()

	_, ok := hub.GetProjectVersion(projectVersionUrl)
	if !ok {
		log.Printf("Invalid project version url %s on this Hub server\n", projectVersionUrl)
	}

	return ok
}

func ProjectVersionResults(info ImageInfo, imageId string, taggedName string, sha string, scanId string, projectVersionUrl string, hub *hubServer, annotate *Annotator) (e error, results ImageInfo) {
	log.Printf("Processing vulnerabilities and policy violations for %s:%s\n", taggedName, imageId[:10])

	projectVersion, ok := hub.GetProjectVersion(projectVersionUrl)
	if !ok {
		e := fmt.Sprintf("ERROR no project version found for image: %s", imageId)
		log.Printf("%s\n", e)
		return errors.New(e), info
	}

	vulnerabilities := 0
	violations := 0
	projectVersionUI := ""

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

		if strings.Compare(Item.Rel, "components") == 0 {
			projectVersionUI = Item.Href
		}
	}

	log.Printf("Found %d high severity vulnerabilities and %d policy violations for %s:%s\n", vulnerabilities, violations, taggedName, imageId[:10])

	results = annotate.UpdateAnnotations(info, violations, vulnerabilities, projectVersionUrl, scanId, projectVersionUI)

	return nil, results
}

// GetScanResultsFromProjectVersion obtains the vuln data from the Hub based solely on the project and version.
func GetScanResultsFromProjectVersion(projectName string, version string, annotate *Annotator, hubConfig *HubConfig) (e error, results ImageInfo) {
	log.Printf("Checking for vulnerabilities on project: %s:%s\n", projectName, version)

	info := ImageInfo{
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}

	hub := NewHubServer(hubConfig)
	if ok := hub.Login(); !ok {
		log.Printf("Hub credentials not valid\n")
		return errors.New("Invalid Hub credentials"), info
	}

	defer hub.Logout()

	projects := hub.FindProjects(projectName)

	if projects.TotalCount == 0 {
		e := fmt.Sprintf("ERROR no project information found for project: %s", projectName)
		log.Printf("%s\n", e)
		return errors.New(e), info
	}

	if projects.TotalCount != 1 {
		log.Printf("Multiple projects found for project %s. Assuming first is correct\n", projectName)
	}

	href := projects.Items[0].Meta.Href

	el := strings.Split(href, "/")
	projectId := el[len(el)-1]

	projectVersions := hub.FindProjectVersions(projectId, version)

	if projectVersions.TotalCount == 0 {
		e := fmt.Sprintf("ERROR no project version information found for project: %s:%s with ID %s", projectName, version, projectId)
		log.Printf("%s\n", e)
		return errors.New(e), info
	}

	if projectVersions.TotalCount != 1 {
		log.Printf("Multiple project versions found for project %s:%s with ID %s. Assuming first is correct\n", projectName, version, projectId)
	}

	projectVersionUrl := projectVersions.Items[0].Meta.Href

	projectVersion, ok := hub.GetProjectVersion(projectVersionUrl)
	if !ok {
		e := fmt.Sprintf("ERROR no project version found for project %s:%s with ID %s\n", projectName, version, projectId)
		log.Printf("%s\n", e)
		return errors.New(e), info
	}

	vulnerabilities := 0
	violations := 0
	projectVersionUI := ""
	foundPolicy := false
	foundVulns := false

	// we have a matching version for our image, need to locate the risk-profile
	for _, Item := range projectVersion.Meta.Links {

		log.Printf("  Processing project version link: %s with url: %s\n", Item.Rel, Item.Href)
		if strings.Compare(Item.Rel, "riskProfile") == 0 {
			riskProfile, ok := hub.GetRiskProfile(Item.Href)
			if riskProfile == nil || !ok {
				e := fmt.Sprintf("ERROR unable to load risk profile for project: %s:%s", projectName, version)
				log.Printf("%s\n", e)
				return errors.New(e), info
			}
			vulnerabilities = riskProfile.Categories.VULNERABILITY.HIGH
			foundVulns = true
		}

		if strings.Compare(Item.Rel, "policy-status") == 0 {
			policyStatus, ok := hub.GetPolicyStatus(Item.Href)
			if policyStatus == nil || !ok {
				e := fmt.Sprintf("ERROR unable to load policy status for image: %s:%s", projectName, version)
				log.Printf("%s\n", e)
				return errors.New(e), info
			}
			for _, PolicyItem := range policyStatus.ComponentVersionStatusCounts {
				if strings.Compare(PolicyItem.Name, "IN_VIOLATION") == 0 {
					violations = PolicyItem.Value
					foundPolicy = true
				}
			}

		}

		if strings.Compare(Item.Rel, "components") == 0 {
			projectVersionUI = Item.Href
		}

	}

	if foundVulns || foundPolicy {
		log.Printf("Found %d high severity vulnerabilities and %d policy violations for project %s:%s with status %v:%v\n", vulnerabilities, violations, projectName, version, foundVulns, foundPolicy)

		results = annotate.UpdateAnnotations(info, violations, vulnerabilities, projectVersionUrl, "", projectVersionUI)
		return nil, results
	} else {
		e := fmt.Sprintf("ERROR unable to load risk information for project: %s:%s", projectName, version)
		log.Printf("%s\n", e)
		return errors.New(e), info
	}
}
