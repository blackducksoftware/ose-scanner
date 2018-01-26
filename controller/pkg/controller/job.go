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
	"encoding/json"

	bdscommon "github.com/blackducksoftware/ose-scanner/common"

	osimageapi "github.com/openshift/api/image/v1"

	kapi "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"

	"log"
	"strings"
)

type Job struct {
	ScanImage  *ScanImage
	controller *Controller
}

type Spec struct {
	Metadata bdscommon.ImageInfo `json:"metadata"`
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

func (job Job) imageScanned(spec string) bool {
	return job.controller.imageScanned(spec)
}

func (job Job) getImageAnnotationInfo(image *osimageapi.Image) (info bdscommon.ImageInfo) {

	info.Annotations = image.ObjectMeta.Annotations
	if info.Annotations == nil {
		log.Printf("Image %s has no annotations - creating object.\n", job.ScanImage.sha)
		info.Annotations = make(map[string]string)
	}
	info.Labels = image.ObjectMeta.Labels
	if info.Labels == nil {
		log.Printf("Image %s has no labels - creating object.\n", job.ScanImage.sha)
		info.Labels = make(map[string]string)
	}
	return info
}

func (job Job) UpdateImageAnnotationInfo(newInfo bdscommon.ImageInfo) bool {

	if job.controller.openshiftClient == nil {
		// if there's no OpenShift client, there can't be any image annotations
		return false
	}

	image, err := job.controller.openshiftClient.Images().Get(job.ScanImage.sha, metav1.GetOptions{})
	if err != nil {
		log.Printf("Job: Error getting image %s: %s\n", job.ScanImage.sha, err)
		return false
	}

	oldInfo := job.getImageAnnotationInfo(image)

	results := job.mergeAnnotationResults(oldInfo, newInfo)

	image.ObjectMeta.Annotations = results.Annotations

	image.ObjectMeta.Labels = results.Labels

	image, err = job.controller.openshiftClient.Images().Update(image)
	if err != nil {
		log.Printf("Error updating annotations for image: %s. %s\n", job.ScanImage.sha, err)
		return false
	}

	return true
}

func (job Job) getPodAnnotationInfo(pod *kapi.Pod, podName string) (info bdscommon.ImageInfo) {

	info.Annotations = pod.ObjectMeta.Annotations
	if info.Annotations == nil {
		log.Printf("Pod %s has no annotations - creating object.\n", podName)
		info.Annotations = make(map[string]string)
	}
	info.Labels = pod.ObjectMeta.Labels
	if info.Labels == nil {
		log.Printf("Pod %s has no labels - creating object.\n", podName)
		info.Labels = make(map[string]string)
	}
	return info
}

// UpdatePodAnnotationInfo updates an existing pod annotations/lables with our scan results
func (job Job) UpdatePodAnnotationInfo(namespace string, podName string, newInfo bdscommon.ImageInfo) bool {

	if job.controller.kubeClient == nil {
		// k8s client should always be present, but safe trumps sorry
		return false
	}

	spec := &Spec{}

	spec.Metadata.Annotations = newInfo.Annotations
	spec.Metadata.Labels = newInfo.Labels

	patch, err := json.Marshal(spec)
	if err != nil {
		log.Printf("Job: Error marshalling spec for pod %s: %s\n", podName, err)
		return false
	}
	patchBytes := []byte(patch)

	_, err = job.controller.kubeClient.CoreV1().RESTClient().Patch(types.StrategicMergePatchType).
		NamespaceIfScoped(namespace, true).
		Resource("pods").
		Name(podName).
		Body(patchBytes).
		Do().
		Get()

	if err != nil {
		log.Printf("Error updating annotations for pod: %s in namespace %s. %s\n", podName, namespace, err)
		return false
	}

	return true
}

func (job Job) ApplyAnnotationInfoToPods(image string, newInfo bdscommon.ImageInfo) bool {

	for _, podInfo := range job.controller.imageUsage[image] {
		ok := job.UpdatePodAnnotationInfo(podInfo.namespace, podInfo.name, newInfo)
		if !ok {
			log.Printf("Unable to annotate pods for image %s\n", image)
			return false
		}
	}
	return true
}

func (job Job) mergeAnnotationResults(oldInfo bdscommon.ImageInfo, newInfo bdscommon.ImageInfo) bdscommon.ImageInfo {

	for k, v := range newInfo.Labels {
		oldInfo.Labels[k] = v
	}

	for k, v := range newInfo.Annotations {
		oldInfo.Annotations[k] = v
	}

	return oldInfo

}

// IsImageStreamScanNeeded determines if a scan of the specified image in an OpenShift ImageStream is required
func (job Job) IsImageStreamScanNeeded(hubConfig *bdscommon.HubConfig) bool {

	if job.controller.openshiftClient == nil {
		// if there's no OpenShift client, there can't be any image annotations
		return false
	}

	image, err := job.controller.openshiftClient.Images().Get(job.ScanImage.sha, metav1.GetOptions{})
	if err != nil {
		log.Printf("Job: Error getting image %s: %s\n", job.ScanImage.sha, err)
		return false
	}

	ref := job.ScanImage.digest

	info := job.getImageAnnotationInfo(image)

	annotations := info.Annotations
	if annotations == nil {
		// no annotations means we've never been here before
		log.Printf("Nil annotations on image: %s\n", ref)
		return true
	}

	versionRequired := true
	bdsVer, ok := annotations[bdscommon.ScannerVersionLabel]
	if ok && (strings.Compare(bdsVer, job.controller.hubParams.Version) == 0) {
		log.Printf("Image %s has been scanned by our scanner.\n", ref)
		versionRequired = false
	}

	hubRequired := true
	hubHost, ok := annotations[bdscommon.ScannerHubServerLabel]
	if ok && (strings.Compare(hubHost, job.controller.hubParams.Config.Host) == 0) {
		log.Printf("Image %s has been scanned by our Hub server.\n", ref)
		hubRequired = false
	}

	projectVersionRescan := true
	projectVersionUrl, ok := annotations[bdscommon.ScannerProjectVersionUrl]
	if ok && bdscommon.ValidateGetProjectVersion(projectVersionUrl, hubConfig) {
		log.Printf("Image %s is present at url %s.\n", ref, projectVersionUrl)
		projectVersionRescan = false
	}

	if versionRequired || hubRequired || projectVersionRescan {
		log.Printf("Image %s scan required due to missing or invalid configuration\n", ref)
		return true
	}

	return false
}
