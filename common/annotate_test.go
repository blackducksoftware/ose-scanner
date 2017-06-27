package common

import (
	"testing"
	"encoding/json"
	"strings"
)

func TestMapMerge(t *testing.T) {
	a := map[string]string{"a":"b"}
	b := map[string]string{"c":"d"}
	c := mapMerge(a,b)
	if len(c) != 2{
		t.Fail()
	}

	a = map[string]string{"a":"b"}
	b = map[string]string{"a":"b"}
	c = mapMerge(a,b)
	if len(c) != 1{
		t.Fail()
	}

	a = map[string]string{"a":"b", "c":"d"}
	b = map[string]string{"aa":"bb","cc":"dd"}
	c = mapMerge(a,b)
	if len(c) != 4{
		t.Fail()
	}

	// make sure mutations work right, run w/ -v to make sure logging is working right.
	a = map[string]string{"a":"b", "c":"d"}
	b = map[string]string{"a":"b","c":"e"}
	c = mapMerge(a,b)
	if len(c) != 2{
		t.Fail()
	}
	if c["c"] != "e" {
		t.Fail()
	}

	// make sure mutations work right, run w/ -v to make sure logging is working right.
	a = map[string]string{}
	b = map[string]string{"a":"b","c":"e"}
	c = mapMerge(a,b)
	if len(c) != 2{
		t.Fail()
	}
	if c["c"] != "e" {
		t.Fail()
	}
}

func TestOpenshiftAnnotations(t *testing.T) {

	mockAnnotator := &Annotator{
		HubServer: "",
		ScannerVersion: "none",
	}
	mockImageInfo := ImageInfo{
	}

	newImageInfo := mockAnnotator.UpdateAnnotations(mockImageInfo,"ref",3,4,"https://example.com","random-scanid-1234")

	// explicit test to confirm that the exact openshift guidelines are met, key should be
	// "quality.images.openshift.io/vulnerability.blackduck"
	vulnerabilityAsString := newImageInfo.Annotations["quality.images.openshift.io/vulnerability.blackduck"]
	var vmap map[string]string
	_ = json.Unmarshal([]byte(vulnerabilityAsString),&vmap)
	if !strings.Contains(string(vmap["summary"]),"label:high") {
		t.Fatalf("Failed finding\n( %v ) in \n( %v )","label:high", vmap["summary"])
	}

	vmap = nil
	// policy similar to vuln, w/ 'important' instead of 'high'.

	policyAsString := newImageInfo.Annotations["quality.images.openshift.io/policy.blackduck"]
	_ = json.Unmarshal([]byte(policyAsString),&vmap)
	if !strings.Contains(string(vmap["summary"]),"label:important") {
		t.Fatalf("Failed finding\n( %v ) in \n( %v )","label:important",vmap["summary"])
	}
}
