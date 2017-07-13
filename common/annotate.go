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
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

const scannerVersionLabel = "blackducksoftware.com/hub-scanner-version"
const scannerHubServerLabel = "blackducksoftware.com/attestation-hub-server"
const ScannerScanId = "blackducksoftware.com/scan-id"
const ScannerProjectVersionUrl = "blackducksoftware.com/project-endpoint"

type Annotator struct {
	ScannerVersion string
	HubServer      string
}

type ImageInfo struct {
	Labels      map[string]string
	Annotations map[string]string
}

// Create a new annotator
func NewAnnotator(ScannerVersion string, HubServer string) *Annotator {
	wc := &Annotator{
		ScannerVersion: ScannerVersion,
		HubServer:      HubServer,
	}
	return wc
}

func mapMerge(base map[string]string, new map[string]string) map[string]string {
	newMap := make(map[string]string)
	if base != nil {
		for k, v := range base {
			newMap[k] = v
		}
	}
	for k, v := range new {
		// if we're overwriting w/ a new value, log.  Don't overlog b/c we expect the arbiter
		// to overwrite quite often (every 30 minutes checks in with KB).
		if v != newMap[k] {
			log.Printf("Image annotation update: [ %s ] FROM %s TO %s", k, newMap[k], v)
		}
		newMap[k] = v
	}
	return newMap
}

// UpdateAnnotations creates a NEW image from an old one, and returns a new image info with annotations and labels from scan results.
// TODO Rename this function to express the fact that it isn't actually updating any data structure, but rather creating a new one.
func (a *Annotator) UpdateAnnotations(inputImageInfo ImageInfo, ref string, violations int, vulnerabilitiies int, projectVersionUrl string, scanId string, projectVersionUIUrl string) ImageInfo {
	policy := "None"
	hasPolicyViolations := "false"

	vulns := "None"
	hasVulns := "false"

	if violations != 0 {
		policy = fmt.Sprintf("%d", violations)
		hasPolicyViolations = "true"
	}

	if vulnerabilitiies != 0 {
		vulns = fmt.Sprintf("%d", vulnerabilitiies)
		hasVulns = "true"
	}

	newLabels := make(map[string]string)
	newLabels["com.blackducksoftware.image.policy-violations"] = policy
	newLabels["com.blackducksoftware.image.has-policy-violations"] = hasPolicyViolations
	newLabels["com.blackducksoftware.image.vulnerabilities"] = vulns
	newLabels["com.blackducksoftware.image.has-vulnerabilities"] = hasVulns
	inputImageInfo.Labels = mapMerge(inputImageInfo.Labels, newLabels)

	newAnnotations := make(map[string]string)
	newAnnotations[scannerVersionLabel] = a.ScannerVersion
	newAnnotations[scannerHubServerLabel] = a.HubServer
	newAnnotations[ScannerProjectVersionUrl] = projectVersionUrl
	newAnnotations[ScannerScanId] = scanId
	inputImageInfo.Annotations = mapMerge(inputImageInfo.Annotations, newAnnotations)

	// TODO: What is this commented code for @TMACKEY ?
	//attestation := fmt.Sprintf("%s~%s", component, project)
	//annotations["blackducksoftware.com/attestation"] = base64.StdEncoding.EncodeToString([]byte(project))

	vulnAnnotations := a.CreateBlackduckVulnerabilityAnnotation(hasVulns == "true", projectVersionUIUrl, vulns)
	policyAnnotations := a.CreateBlackduckPolicyAnnotation(hasPolicyViolations == "true", projectVersionUIUrl, policy)

	inputImageInfo.Annotations["quality.images.openshift.io/vulnerability.blackduck"] = vulnAnnotations.AsString()
	inputImageInfo.Annotations["quality.images.openshift.io/policy.blackduck"] = policyAnnotations.AsString()

	return inputImageInfo
}

type BlackduckAnnotation struct {
	name        string              `json:"name"`
	description string              `json:"description"`
	timestamp   time.Time           `json:"timestamp"`
	reference   string              `json:"reference"`
	compliant   bool                `json:"compliant"`
	summary     []map[string]string `json:"summary"`
}

// AsString makes a map corresponding to the Openshift Container Security guide (https://people.redhat.com/aweiteka/docs/preview/20170510/security/container_content.html).
func (o *BlackduckAnnotation) AsString() string {
	m := make(map[string]string)
	m["name"] = o.name
	m["description"] = o.description
	m["timestamp"] = fmt.Sprintf("%v", o.timestamp)
	m["reference"] = o.reference
	m["compliant"] = fmt.Sprintf("%v", o.compliant)
	m["summary"] = fmt.Sprintf("%s", o.summary)
	mp, _ := json.Marshal(m)
	return string(mp)
}

// CreateOpenshiftAnnotations takes the primitive information from UpdateAnnotation and translates it to openshift.
func (a *Annotator) CreateBlackduckVulnerabilityAnnotation(hasVulns bool, humanReadableURL string, vulnCount string) *BlackduckAnnotation {
	return &BlackduckAnnotation{
		"blackducksoftware",
		"Vulnerability Info",
		time.Now(),
		humanReadableURL,
		!hasVulns, // no vunls -> compliant.
		[]map[string]string{
			{
				"label":         "high",
				"score":         fmt.Sprintf("%s", vulnCount),
				"severityIndex": fmt.Sprintf("%v", 1),
			},
		},
	}
}
func (a *Annotator) CreateBlackduckPolicyAnnotation(hasPolicyViolations bool, humanReadableURL string, policyCount string) *BlackduckAnnotation {
	return &BlackduckAnnotation{
		"blackducksoftware",
		"Policy Info",
		time.Now(),
		humanReadableURL,
		!hasPolicyViolations, // no violations -> compliant
		[]map[string]string{
			{
				"label":         "important",
				"score":         fmt.Sprintf("%s", policyCount),
				"severityIndex": fmt.Sprintf("%v", 1),
			},
		},
	}
}

// Determine if a scan of the specified image is required
func (a *Annotator) IsScanNeeded(info ImageInfo, ref string, hubConfig *HubConfig) bool {

	annotations := info.Annotations
	if annotations == nil {
		// no annotations means we've never been here before
		log.Printf("Nil annotations on image: %s\n", ref)
		return true
	}

	versionRequired := true
	bdsVer, ok := annotations[scannerVersionLabel]
	if ok && (strings.Compare(bdsVer, a.ScannerVersion) == 0) {
		log.Printf("Image %s has been scanned by our scanner.\n", ref)
		versionRequired = false
	}

	hubRequired := true
	hubHost, ok := annotations[scannerHubServerLabel]
	if ok && (strings.Compare(hubHost, a.HubServer) == 0) {
		log.Printf("Image %s has been scanned by our Hub server.\n", ref)
		hubRequired = false
	}

	projectVersionRescan := true
	projectVersionUrl, ok := annotations[ScannerProjectVersionUrl]
	if ok && ValidateGetProjectVersion(projectVersionUrl, hubConfig) {
		log.Printf("Image %s is present at url %s.\n", ref, projectVersionUrl)
		projectVersionRescan = false
	}

	if versionRequired || hubRequired || projectVersionRescan {
		log.Printf("Image %s scan required due to missing or invalid configuration\n", ref)
		return true
	}

	return false
}
