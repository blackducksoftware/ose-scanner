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
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

const hangtimeBeforeTimingOutOnTheHub = 10 * time.Second

type myjar struct {
	// why is a jar a map of string->cookie1,cookie2,...?
	jar map[string][]*http.Cookie
}

// unused ?
func (p *myjar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	p.jar[u.Host] = cookies
}

// Is this supposed to be a "GetCookies" method?
func (p *myjar) Cookies(u *url.URL) []*http.Cookie {
	return p.jar[u.Host]
}

// NewHubServer creates a new connection to a Hub server
func NewHubServer(config *HubConfig) *hubServer {

	return &hubServer{
		client: &http.Client{
			Timeout:   hangtimeBeforeTimingOutOnTheHub,
			Transport: config.Wire,
		},
		config:   config,
		loggedIn: false,
	}
}

// Login performs a login to the hub server. Note an explicit logout is required.
func (h *hubServer) Login() bool {
	// check if the Config entry is initialized
	if h.config == nil {
		log.Printf("ERROR in hubServer no configuration available.\n")
		return false
	}

	log.Printf("Login attempt for %s\n", h.config.Url)
	u, err := url.ParseRequestURI(h.config.Url)
	if err != nil {
		log.Printf("ERROR : url.ParseRequestURI: %s\n", err)
		return false
	}

	resource := "/j_spring_security_check"
	u.Path = resource
	data := url.Values{}
	data.Add("j_username", h.config.User)
	data.Add("j_password", h.config.Password)

	jar := &myjar{}
	jar.jar = make(map[string][]*http.Cookie)
	h.client.Jar = jar

	urlStr := fmt.Sprintf("%v", u)
	req, err := http.NewRequest("POST", urlStr, bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Printf("ERROR NewRequest: %s\n", err)
		return false
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded") // needed this the prevent 401 Unauthorized

	resp, err := h.client.Do(req)
	if err != nil {
		log.Printf("ERROR client.do %s\n", err)
		return false
	}
	resp.Body.Close()
	if resp.StatusCode != 204 {
		log.Printf("ERROR: resp status: %s (%d)\n", resp.Status, resp.StatusCode)
		return false
	}
	h.loggedIn = true
	return true
}

// Logout logs the user out of the hub server
func (h *hubServer) Logout() bool {
	// check if the Config entry is initialized
	if h.config == nil {
		log.Printf("ERROR in hubServer no configuration available.\n")
		return false
	}

	log.Printf("Logout attempt on %s when logged in is %v\n", h.config.Url, h.loggedIn)
	u, err := url.ParseRequestURI(h.config.Url)
	if err != nil {
		log.Printf("ERROR : url.ParseRequestURI: %s\n", err)
		return false
	}

	resource := "/j_spring_security_logout"
	u.Path = resource

	urlStr := fmt.Sprintf("%v", u)

	resp, err := h.client.Get(urlStr)
	if err != nil {
		log.Printf("ERROR client.Get %s\n", err)
		return false
	}

	resp.Body.Close()
	if resp.StatusCode != 204 {
		log.Printf("ERROR: resp status: %s (%d)\n", resp.Status, resp.StatusCode)
		return false
	}
	return true
}

// GetCodeLocation takes a fully qualified apiUrl to a Code Location (aka scan record) and returns the result
func (h *hubServer) GetCodeLocation(apiUrl string) (*CodeLocationStruct, bool) {

	log.Println(apiUrl)

	var codeLocation CodeLocationStruct

	buf := h.getHubRestEndPointJson(apiUrl)
	if buf.Len() == 0 {
		log.Printf("Error no response for url: %s\n", apiUrl)
		return &codeLocation, false
	}

	if err := json.Unmarshal([]byte(buf.String()), &codeLocation); err != nil {
		log.Printf("ERROR Unmarshall error: %s\n", err)
		return &codeLocation, false
	}
	return &codeLocation, true
}

// FindCodeLocations takes a search condition and finds the corresponding code location (aka scan record)
func (h *hubServer) FindCodeLocations(searchCriterea string) *CodeLocationsStruct {
	searchStr := url.QueryEscape(searchCriterea)
	getStr := fmt.Sprintf("%s/api/codelocations/?q=%s&limit=5000", h.config.Url, searchStr)
	log.Println(getStr)

	var codeLocations CodeLocationsStruct

	buf := h.getHubRestEndPointJson(getStr)
	if buf.Len() == 0 {
		log.Printf("Error no response for url: %s\n", getStr)
	}

	if err := json.Unmarshal([]byte(buf.String()), &codeLocations); err != nil {
		log.Printf("ERROR Unmarshall error: %s\n", err)
	}
	return &codeLocations
}

// FindCodeLocationScanSummaries takes an API url and returns code location (aka scan record) summary
func (h *hubServer) FindCodeLocationScanSummaries(url string) *codeLocationsScanSummariesStruct {

	log.Println(url)

	var codeLocationsScanSummaries codeLocationsScanSummariesStruct

	buf := h.getHubRestEndPointJson(url)
	if buf.Len() == 0 {
		log.Printf("Error no response for url: %s\n", url)
	}

	if err := json.Unmarshal([]byte(buf.String()), &codeLocationsScanSummaries); err != nil {
		log.Printf("ERROR Unmarshall error: %s\n", err)
	}
	return &codeLocationsScanSummaries
}

// GetScanSummary takes a scanID and returns a scan summary
func (h *hubServer) GetScanSummary(scanId string) (*ScanSummaryStruct, bool) {
	apiUrl := fmt.Sprintf("%s/api/scan-summaries/%s", h.config.Url, scanId)

	log.Println(apiUrl)

	var scanSummary ScanSummaryStruct

	buf := h.getHubRestEndPointJson(apiUrl)
	if buf.Len() == 0 {
		log.Printf("Error no response for url: %s\n", apiUrl)
		return &scanSummary, false
	}

	if err := json.Unmarshal([]byte(buf.String()), &scanSummary); err != nil {
		log.Printf("ERROR Unmarshall error: %s\n", err)
		return &scanSummary, false
	}

	return &scanSummary, true
}

// FindProjects takes a project name and returns a project summary
func (h *hubServer) FindProjects(projectName string) *projectsStruct {
	searchCriterea := "name:" + projectName
	searchStr := url.QueryEscape(searchCriterea)
	getStr := fmt.Sprintf("%s/api/projects/?q=%s&limit=5000", h.config.Url, searchStr)
	log.Println(getStr)

	var projects projectsStruct

	buf := h.getHubRestEndPointJson(getStr)
	if buf.Len() == 0 {
		log.Printf("Error no response for url: %s\n", getStr)
	}

	if err := json.Unmarshal([]byte(buf.String()), &projects); err != nil {
		log.Printf("ERROR Unmarshall error: %s\n", err)
	}
	return &projects
}

// GetProjectVersion takes an api project reference and returns a project version
func (h *hubServer) GetProjectVersion(apiUrl string) (*projectVersionStruct, bool) {

	log.Println(apiUrl)

	var projectVersion projectVersionStruct

	buf := h.getHubRestEndPointJson(apiUrl)
	if buf.Len() == 0 {
		log.Printf("Error no response for url: %s\n", apiUrl)
		return &projectVersion, false
	}

	if err := json.Unmarshal([]byte(buf.String()), &projectVersion); err != nil {
		log.Printf("ERROR Unmarshall error: %s\n", err)
		return &projectVersion, false
	}
	return &projectVersion, true
}

// FindProjectVersions takes a project ID and a search query for a version and returns a Hub record for the projectVersion
func (h *hubServer) FindProjectVersions(projectId string, projectVersion string) *projectVersionsStruct {
	searchCriterea := "versionName:" + projectVersion
	searchStr := url.QueryEscape(searchCriterea)
	getStr := fmt.Sprintf("%s/api/projects/%s/versions?q=%s&limit=5000", h.config.Url, projectId, searchStr)
	log.Println(getStr)

	var projectVersions projectVersionsStruct

	buf := h.getHubRestEndPointJson(getStr)
	if buf.Len() == 0 {
		log.Printf("Error no response for url: %s\n", getStr)
	}

	if err := json.Unmarshal([]byte(buf.String()), &projectVersions); err != nil {
		log.Printf("ERROR Unmarshall error: %s\n", err)
	}
	return &projectVersions
}

// GetRiskProfile takes an API endpoint and returns a risk profile for the project version
func (h *hubServer) GetRiskProfile(apiUrl string) (*riskProfileStruct, bool) {

	log.Println(apiUrl)

	var riskProfile riskProfileStruct

	buf := h.getHubRestEndPointJson(apiUrl)
	if buf.Len() == 0 {
		log.Printf("Error no response for url: %s\n", apiUrl)
		return &riskProfile, false
	}

	if err := json.Unmarshal([]byte(buf.String()), &riskProfile); err != nil {
		log.Printf("ERROR Unmarshall error: %s\n", err)
		return &riskProfile, false
	}
	return &riskProfile, true
}

// GetPolicyStatus takes an API endpint for a project version and returns policy information
func (h *hubServer) GetPolicyStatus(apiUrl string) (*policyStatusStruct, bool) {

	log.Println(apiUrl)

	var policyStatus policyStatusStruct

	buf := h.getHubRestEndPointJson(apiUrl)
	if buf.Len() == 0 {
		log.Printf("Error no response for url: %s\n", apiUrl)
		return &policyStatus, false
	}

	if err := json.Unmarshal([]byte(buf.String()), &policyStatus); err != nil {
		log.Printf("ERROR Unmarshall error: %s\n", err)
		return &policyStatus, false
	}
	return &policyStatus, true
}

func (h *hubServer) getHubRestEndPointJson(restEndPointUrl string) *bytes.Buffer {

	buf := new(bytes.Buffer)
	resp, err := h.client.Get(restEndPointUrl)
	if err != nil {
		log.Printf("ERROR in client.url : %s get: %s\n", restEndPointUrl, err)
		return buf
	}

	defer resp.Body.Close()

	log.Printf("Endpoint status %s\n", resp.Status)

	if resp.StatusCode != 200 {
		log.Printf("ERROR return status : %s url:%s\n", resp.Status, restEndPointUrl)
		return buf
	}

	if _, err := buf.ReadFrom(resp.Body); err != nil {
		log.Printf("ERROR reading from response: %s url: %s\n", err, restEndPointUrl)
		return buf
	}

	return buf
}
