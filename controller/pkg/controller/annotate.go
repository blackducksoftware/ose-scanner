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
	//	"encoding/base64"
	"fmt"
	"log"
	"strings"

	kapi "k8s.io/kubernetes/pkg/api"

	osclient "github.com/openshift/origin/pkg/client"
)

const scannerVersionLabel = "blackducksoftware.com/hub-scanner-version"
const scannerHubServerLabel = "blackducksoftware.com/attestation-hub-server"

type Annotator struct {
	openshiftClient *osclient.Client
	Namespace       string
	ScannerVersion  string
	HubServer       string
}

// Create a new annotator
func NewAnnotator(os *osclient.Client, ScannerVersion string, HubServer string) *Annotator {

	namespace := kapi.NamespaceAll

	wc := &Annotator{
		openshiftClient: os,
		Namespace:       namespace,
		ScannerVersion:  ScannerVersion,
		HubServer:       HubServer,
	}
	return wc
}

// Save the results of a scan on the specified image
func (a *Annotator) SaveResults(ref string, violations int, vulnerabilitiies int, projectVersionUrl string, scanId string) bool {

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

	image, err := a.openshiftClient.Images().Get(ref)
	if err != nil {
		log.Printf("Error image reference %s: %s\n", ref, err)
		return false
	}

	labels := image.ObjectMeta.Labels
	if labels == nil {
		log.Printf("Image %s has no labels - creating.\n", ref)
		labels = make(map[string]string)
	}
	labels["com.blackducksoftware.image.policy-violations"] = policy
	labels["com.blackducksoftware.image.has-policy-violations"] = hasPolicyViolations

	labels["com.blackducksoftware.image.vulnerabilities"] = vulns
	labels["com.blackducksoftware.image.has-vulnerabilities"] = hasVulns
	image.ObjectMeta.Labels = labels

	annotations := image.ObjectMeta.Annotations
	if annotations == nil {
		log.Printf("Image %s has no annotations - creating.\n", ref)
		annotations = make(map[string]string)
	}

	annotations[scannerVersionLabel] = a.ScannerVersion
	annotations[scannerHubServerLabel] = a.HubServer
	annotations["blackducksoftware.com/project-endpoint"] = projectVersionUrl
	annotations["blackducksoftware.com/scan-id"] = scanId

	//attestation := fmt.Sprintf("%s~%s", component, project)
	//annotations["blackducksoftware.com/attestation"] = base64.StdEncoding.EncodeToString([]byte(project))
	image.ObjectMeta.Annotations = annotations

	image, err = a.openshiftClient.Images().Update(image)
	if err != nil {
		log.Printf("Error updating image: %s. %s\n", ref, err)
		return false
	}

	log.Printf("Applied annotation for image: %s.\n", ref)

	return true

}

// Determine if a scan of the specified image is required
func (a *Annotator) IsScanNeeded(ref string) bool {

	image, err := a.openshiftClient.Images().Get(ref)
	if err != nil {
		// most likely indicates missing image - resolve that later by queuing for scan
		log.Printf("Error testing if scan needed for %s: %s\n", ref, err)
		return true
	}

	annotations := image.ObjectMeta.Annotations
	if annotations == nil {
		// no annotations means we've never been here before
		log.Printf("Nil annotations on image: %s\n", ref)
		return true
	}

	versionRequired := true
	bdsVer, ok := annotations[scannerVersionLabel]
	if ok && (strings.Compare(bdsVer, a.ScannerVersion) == 0) {
		log.Printf("Image %s has been scanned by our scanner. Skipping new scan.\n", ref)
		versionRequired = false
	}

	hubRequired := true
	hubHost, ok := annotations[scannerHubServerLabel]
	if ok && (strings.Compare(hubHost, a.HubServer) == 0) {
		log.Printf("Image %s has been scanned by our Hub server. Skipping new scan.\n", ref)
		hubRequired = false
	}

	if versionRequired || hubRequired {
		return true
	}

	return false
}
