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
	"encoding/base64"
	"fmt"
	"log"
	//"strings"

	kapi "k8s.io/kubernetes/pkg/api"

	osclient "github.com/openshift/origin/pkg/client"
)

const scannerVersionLabel = "blackducksoftware.com/hub-scanner-version"

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
func (a *Annotator) SaveResults(ref string, violations int, project string) bool {

	policy := "None"
	hasPolicyViolations := "false"

	if violations != 0 {
		policy = fmt.Sprintf("%d", violations)
		hasPolicyViolations = "true"
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
	labels["com.blackducksoftware.policy-violations"] = policy
	labels["com.blackducksoftware.has-policy-violations"] = hasPolicyViolations
	image.ObjectMeta.Labels = labels

	annotations := image.ObjectMeta.Annotations
	if annotations == nil {
		log.Printf("Image %s has no annotations - creating.\n", ref)
		annotations = make(map[string]string)
	}

	annotations[scannerVersionLabel] = a.ScannerVersion
	annotations["blackducksoftware.com/attestation-hub-server"] = a.HubServer

	//attestation := fmt.Sprintf("%s~%s", component, project)
	annotations["blackducksoftware.com/attestation"] = base64.StdEncoding.EncodeToString([]byte(project))
	image.ObjectMeta.Annotations = annotations

	image, err = a.openshiftClient.Images().Update(image)
	if err != nil {
		log.Printf("Error updating image: %s. %s\n", ref, err)
		return false
	}

	log.Printf("Applied annotation for image: %s.\n", ref)

	return true

}
