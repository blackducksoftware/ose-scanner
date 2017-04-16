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
	"fmt"
	"log"
	"strings"
	//kapi "k8s.io/kubernetes/pkg/api"
	//osclient "github.com/openshift/origin/pkg/client"
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

// UpdateAnnotations updates the image annotations and labels with the current scan results
func (a *Annotator) UpdateAnnotations(info ImageInfo, ref string, violations int, vulnerabilitiies int, projectVersionUrl string, scanId string) ImageInfo {

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

	labels := info.Labels
	if labels == nil {
		log.Printf("Image %s has no labels - creating.\n", ref)
		labels = make(map[string]string)
	}
	labels["com.blackducksoftware.image.policy-violations"] = policy
	labels["com.blackducksoftware.image.has-policy-violations"] = hasPolicyViolations

	labels["com.blackducksoftware.image.vulnerabilities"] = vulns
	labels["com.blackducksoftware.image.has-vulnerabilities"] = hasVulns
	info.Labels = labels

	annotations := info.Annotations
	if annotations == nil {
		log.Printf("Image %s has no annotations - creating.\n", ref)
		annotations = make(map[string]string)
	}

	annotations[scannerVersionLabel] = a.ScannerVersion
	annotations[scannerHubServerLabel] = a.HubServer
	annotations[ScannerProjectVersionUrl] = projectVersionUrl
	annotations[ScannerScanId] = scanId

	//attestation := fmt.Sprintf("%s~%s", component, project)
	//annotations["blackducksoftware.com/attestation"] = base64.StdEncoding.EncodeToString([]byte(project))
	info.Annotations = annotations

	return info

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
