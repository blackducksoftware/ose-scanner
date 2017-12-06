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
	"encoding/json"

	bdscommon "github.com/blackducksoftware/ose-scanner/common"
	kapi "k8s.io/kubernetes/pkg/api"
	"log"
	"time"
)

// Job represents a scan activity
type Job struct {
	ScanImage *ScanImage
	PodImage  *PodImage
	arbiter   *Arbiter
}

// Spec is a metadata primative for k8s annotations
type Spec struct {
	Metadata bdscommon.ImageInfo `json:"metadata"`
}

func newJob(scanImage *ScanImage, podImage *PodImage, arb *Arbiter) *Job {
	return &Job{
		ScanImage: scanImage,
		arbiter:   arb,
		PodImage:  podImage,
	}
}

// Done completes processing on the job
func (job Job) Done(result bool) {
	if job.ScanImage != nil {
		job.arbiter.DoneScan(result, job.ScanImage.digest)
	} else if job.PodImage != nil {
		job.arbiter.DonePod(job.PodImage.imageName)
	}

	time.Sleep(100 * time.Millisecond) // allow API server some time to breathe
	return
}

// Load adds a job to the queue
func (job Job) Load() {
	job.arbiter.Add()
	return
}

// GetImageAnnotationInfo gets the current annotation set for the job's image
func (job Job) GetImageAnnotationInfo() (result bool, info bdscommon.ImageInfo) {

	if job.arbiter.openshiftClient == nil {
		// if there's no OpenShift client, there can't be any image annotations
		return false, info
	}

	image, err := job.arbiter.openshiftClient.Images().Get(job.ScanImage.sha)
	if err != nil {
		log.Printf("Job: Error getting image %s: %s\n", job.ScanImage.sha, err)
		return false, info
	}

	info.Annotations = image.ObjectMeta.Annotations
	if info.Annotations == nil {
		log.Printf("Image %s has no annotations - creating object.\n", job.ScanImage.digest)
		info.Annotations = make(map[string]string)
	}

	info.Labels = image.ObjectMeta.Labels
	if info.Labels == nil {
		log.Printf("Image %s has no labels - creating object.\n", job.ScanImage.digest)
		info.Labels = make(map[string]string)
	}
	return true, info
}

// UpdateImageAnnotationInfo applies a merged set of annoations to the specified image
func (job Job) UpdateImageAnnotationInfo(newInfo bdscommon.ImageInfo) bool {

	if job.arbiter.openshiftClient == nil {
		// if there's no OpenShift client, there can't be any image annotations
		return false
	}

	image, err := job.arbiter.openshiftClient.Images().Get(job.ScanImage.sha)
	if err != nil {
		log.Printf("Job: Error getting image %s: %s\n", job.ScanImage.sha, err)
		return false
	}

	_, oldInfo := job.GetImageAnnotationInfo()

	results := job.mergeAnnotationResults(oldInfo, newInfo)

	image.ObjectMeta.Annotations = results.Annotations

	image.ObjectMeta.Labels = results.Labels

	image, err = job.arbiter.openshiftClient.Images().Update(image)
	if err != nil {
		log.Printf("Error updating annotations for image: %s. %s\n", job.ScanImage.digest, err)
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

// UpdatePodAnnotationInfo updates an existing pod annotations/lables with our scna results
func (job Job) UpdatePodAnnotationInfo(namespace string, podName string, newInfo bdscommon.ImageInfo) bool {

	if job.arbiter.kubeClient == nil {
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

	_, err = job.arbiter.kubeClient.RESTClient.Patch(kapi.StrategicMergePatchType).
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

// ApplyAnnotationInfoToPods iterates the list of pods where the image is used and applies optimistic annotations
func (job Job) ApplyAnnotationInfoToPods(newInfo bdscommon.ImageInfo) bool {

	image := job.PodImage.imageName
	for _, podInfo := range job.arbiter.imageUsage[image] {
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
