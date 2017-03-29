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
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

type myjar struct {
	jar map[string][]*http.Cookie
}

func (p *myjar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	p.jar[u.Host] = cookies
}

func (p *myjar) Cookies(u *url.URL) []*http.Cookie {
	return p.jar[u.Host]
}

type HubConfig struct {
	Url      string `json:"url"`
	Host     string `json:"hubhost"`
	Port     string `json:"port"`
	Scheme   string `json:"scheme"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type codeLocationStruct struct {
	Type                 string    `json:"type"`
	Name                 string    `json:"name"`
	URL                  string    `json:"url"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
	MappedProjectVersion string    `json:"mappedProjectVersion"`
	Meta                 struct {
		Allow []string `json:"allow"`
		Href  string   `json:"href"`
		Links []struct {
			Rel  string `json:"rel"`
			Href string `json:"href"`
		} `json:"links"`
	} `json:"_meta"`
}

type codeLocationsStruct struct {
	TotalCount int `json:"totalCount"`
	Items      []struct {
		Type                 string    `json:"type"`
		URL                  string    `json:"url"`
		CreatedAt            time.Time `json:"createdAt"`
		UpdatedAt            time.Time `json:"updatedAt"`
		MappedProjectVersion string    `json:"mappedProjectVersion"`
		Meta                 struct {
			Allow []string `json:"allow"`
			Href  string   `json:"href"`
			Links []struct {
				Rel  string `json:"rel"`
				Href string `json:"href"`
			} `json:"links"`
		} `json:"_meta"`
	} `json:"items"`
	Meta struct {
		Allow []string      `json:"allow"`
		Href  string        `json:"href"`
		Links []interface{} `json:"links"`
	} `json:"_meta"`
	AppliedFilters []interface{} `json:"appliedFilters"`
}

type scanSummaryStruct struct {
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	Meta      struct {
		Allow []string `json:"allow"`
		Href  string   `json:"href"`
		Links []struct {
			Rel  string `json:"rel"`
			Href string `json:"href"`
		} `json:"links"`
	} `json:"_meta"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type codeLocationsScanSummariesStruct struct {
	TotalCount int `json:"totalCount"`
	Items      []struct {
		Status    string    `json:"status"`
		CreatedAt time.Time `json:"createdAt"`
		Meta      struct {
			Allow []string `json:"allow"`
			Href  string   `json:"href"`
			Links []struct {
				Rel  string `json:"rel"`
				Href string `json:"href"`
			} `json:"links"`
		} `json:"_meta"`
		UpdatedAt time.Time `json:"updatedAt"`
	} `json:"items"`
	Meta struct {
		Allow []string      `json:"allow"`
		Href  string        `json:"href"`
		Links []interface{} `json:"links"`
	} `json:"_meta"`
	AppliedFilters []interface{} `json:"appliedFilters"`
}

type projectsStruct struct {
	TotalCount int `json:"totalCount"`
	Items      []struct {
		Name                    string `json:"name"`
		ProjectLevelAdjustments bool   `json:"projectLevelAdjustments"`
		Source                  string `json:"source"`
		Meta                    struct {
			Allow []string `json:"allow"`
			Href  string   `json:"href"`
			Links []struct {
				Rel  string `json:"rel"`
				Href string `json:"href"`
			} `json:"links"`
		} `json:"_meta"`
	} `json:"items"`
}

type projectVersionStruct struct {
	VersionName  string `json:"versionName"`
	Phase        string `json:"phase"`
	Distribution string `json:"distribution"`
	Source       string `json:"source"`
	Meta         struct {
		Allow []string `json:"allow"`
		Href  string   `json:"href"`
		Links []struct {
			Rel  string `json:"rel"`
			Href string `json:"href"`
		} `json:"links"`
	} `json:"_meta"`
}

type projectVersionsStruct struct {
	TotalCount int `json:"totalCount"`
	Items      []struct {
		VersionName  string `json:"versionName"`
		Phase        string `json:"phase"`
		Distribution string `json:"distribution"`
		Source       string `json:"source"`
		Meta         struct {
			Allow []string `json:"allow"`
			Href  string   `json:"href"`
			Links []struct {
				Rel  string `json:"rel"`
				Href string `json:"href"`
			} `json:"links"`
		} `json:"_meta"`
	} `json:"items"`
}

type riskProfileStruct struct {
	Categories struct {
		VERSION struct {
			HIGH    int `json:"HIGH"`
			MEDIUM  int `json:"MEDIUM"`
			LOW     int `json:"LOW"`
			OK      int `json:"OK"`
			UNKNOWN int `json:"UNKNOWN"`
		} `json:"VERSION"`
		VULNERABILITY struct {
			HIGH    int `json:"HIGH"`
			MEDIUM  int `json:"MEDIUM"`
			LOW     int `json:"LOW"`
			OK      int `json:"OK"`
			UNKNOWN int `json:"UNKNOWN"`
		} `json:"VULNERABILITY"`
		ACTIVITY struct {
			HIGH    int `json:"HIGH"`
			MEDIUM  int `json:"MEDIUM"`
			LOW     int `json:"LOW"`
			OK      int `json:"OK"`
			UNKNOWN int `json:"UNKNOWN"`
		} `json:"ACTIVITY"`
		LICENSE struct {
			HIGH    int `json:"HIGH"`
			MEDIUM  int `json:"MEDIUM"`
			LOW     int `json:"LOW"`
			OK      int `json:"OK"`
			UNKNOWN int `json:"UNKNOWN"`
		} `json:"LICENSE"`
		OPERATIONAL struct {
			HIGH    int `json:"HIGH"`
			MEDIUM  int `json:"MEDIUM"`
			LOW     int `json:"LOW"`
			OK      int `json:"OK"`
			UNKNOWN int `json:"UNKNOWN"`
		} `json:"OPERATIONAL"`
	} `json:"categories"`
	Meta struct {
		Allow []string `json:"allow"`
		Href  string   `json:"href"`
		Links []struct {
			Rel  string `json:"rel"`
			Href string `json:"href"`
		} `json:"links"`
	} `json:"_meta"`
}

type policyStatusStruct struct {
	OverallStatus                string    `json:"overallStatus"`
	UpdatedAt                    time.Time `json:"updatedAt"`
	ComponentVersionStatusCounts []struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	} `json:"componentVersionStatusCounts"`
	Meta struct {
		Allow []string      `json:"allow"`
		Href  string        `json:"href"`
		Links []interface{} `json:"links"`
	} `json:"_meta"`
}

type HubServer struct {
	client *http.Client
	Config *HubConfig
}

func (h *HubServer) login() bool {
	// check if the Config entry is initialized
	if h.Config == nil {
		log.Printf("ERROR in HubServer no configuration available.\n")
		return false
	}

	log.Printf("Login attempt for %s\n", h.Config.Url)
	u, err := url.ParseRequestURI(h.Config.Url)
	if err != nil {
		log.Printf("ERROR : url.ParseRequestURI: %s\n", err)
		return false
	}

	resource := "/j_spring_security_check"
	u.Path = resource
	data := url.Values{}
	data.Add("j_username", h.Config.User)
	data.Add("j_password", h.Config.Password)

	h.client = &http.Client{}

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
	return true
}

func (h *HubServer) getCodeLocation(apiUrl string) (*codeLocationStruct, bool) {

	log.Println(apiUrl)

	var codeLocation codeLocationStruct

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

func (h *HubServer) findCodeLocations(searchCriterea string) *codeLocationsStruct {
	searchStr := url.QueryEscape(searchCriterea)
	getStr := fmt.Sprintf("%s/api/codelocations/?q=%s&limit=5000", h.Config.Url, searchStr)
	log.Println(getStr)

	var codeLocations codeLocationsStruct

	buf := h.getHubRestEndPointJson(getStr)
	if buf.Len() == 0 {
		log.Printf("Error no response for url: %s\n", getStr)
	}

	if err := json.Unmarshal([]byte(buf.String()), &codeLocations); err != nil {
		log.Printf("ERROR Unmarshall error: %s\n", err)
	}
	return &codeLocations
}

func (h *HubServer) findCodeLocationScanSummaries(url string) *codeLocationsScanSummariesStruct {

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

func (h *HubServer) getScanSummary(scanId string) (*scanSummaryStruct, bool) {
	apiUrl := fmt.Sprintf("%s/api/scan-summaries/%s", h.Config.Url, scanId)

	log.Println(apiUrl)

	var scanSummary scanSummaryStruct

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

func (h *HubServer) findProjects(projectName string) *projectsStruct {
	searchCriterea := "name:" + projectName
	searchStr := url.QueryEscape(searchCriterea)
	getStr := fmt.Sprintf("%s/api/projects/?q=%s&limit=5000", h.Config.Url, searchStr)
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

func (h *HubServer) getProjectVersion(apiUrl string) (*projectVersionStruct, bool) {

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

func (h *HubServer) findProjectVersions(projectId string, projectVersion string) *projectVersionsStruct {
	searchCriterea := "versionName:" + projectVersion
	searchStr := url.QueryEscape(searchCriterea)
	getStr := fmt.Sprintf("%s/api/projects/%s/versions?q=%s&limit=5000", h.Config.Url, projectId, searchStr)
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

func (h *HubServer) getRiskProfile(apiUrl string) (*riskProfileStruct, bool) {

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

func (h *HubServer) getPolicyStatus(apiUrl string) (*policyStatusStruct, bool) {

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

func (h *HubServer) getHubRestEndPointJson(restEndPointUrl string) *bytes.Buffer {

	buf := new(bytes.Buffer)
	resp, err := h.client.Get(restEndPointUrl)
	if err != nil {
		log.Printf("ERROR in client.url : %s get: %s\n", restEndPointUrl, err)
		return buf
	}
	log.Printf("Endpoint status %s\n", resp.Status)

	if resp.StatusCode != 200 {
		log.Printf("ERROR return status : %s url:%s\n", resp.Status, restEndPointUrl)
		return buf
	}

	if _, err := buf.ReadFrom(resp.Body); err != nil {
		log.Printf("ERROR reading from response: %s url: %s\n", err, restEndPointUrl)
		return buf
	}
	defer resp.Body.Close()

	return buf

}
