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
// ***Note*** Minor fork of the Atomic implementation
package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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


type HubServer struct {
	client *http.Client
	Config *HubConfig
}

type ranking struct {
	HIGH    int `json:"HIGH"`
	MEDIUM  int `json:"MEDIUM"`
	LOW     int `json:"LOW"`
	OK      int `json:"OK"`
	UNKNOWN int `json:"UNKNOWN"`
}

type vulnerabilityBom struct {
	TotalCount int `json:"totalCount"`
	Items      []struct {
		ChannelRelease struct {
			ExternalId                    string `json:"externalId"`
			ExternalNamespace             string `json:"externalNamespace"`
			ExternalNamespaceDistribution bool   `json:"externalNamespaceDistribution"`
			Id                            string `json:"id"`
			Name                          string `json:"name"`
		} `json:"channelRelease"`
		ProducerReleases []struct {
			Id string `json:"id"`
		} `json:"producerReleases"`
	} `json:"items"`
}

type vulnerability struct {
	Id                     string `json:"type"`
	AccessComplexity       string `json:"accessComplexity"`
	AccessVector           string `json:"accessVector"`
	ActualAt               string `json:"actualAt"`
	Authentication         string `json:"authentication"`
	AutoCreated            bool `json:"autoCreated"`
	AvailabilityImpact     string `json:"availabilityImpact"`
	BaseScore              float64 `json:"baseScore"`
	ExploitabilitySubscore float64 `json:"exploitabilitySubscore"`
	ImpactSubscore         float64 `json:"impactSubscore"`
	LastModified           string `json:"lastModified"`
	PublishedDate          string `json:"publishedDate"`
	Severity               string `json:"severity"`
	Solution               string `json:"solution"`
	Source                 string `json:"source"`
	Summary                string `json:"summary"`
	TargetAt               string `json:"targetAt"`
	TechnicalDescription   string `json:"technicalDescription"`
	Title                  string `json:"title"`
	Classifications        []struct {
		ClassificationId int    `json:"classificationId"`
		Description      string `json:"description"`
		Longname         string `json:"longname"`
		Name             string `json:"name"`
	} `json:"classifications"`
	References []struct {
		Content string `json:"content"`
		Href    string `json:"href"`
		Source  string `json:"source"`
		Type    string `json:"type"`
	} `json:"references"`
	RelatedMetrics []struct {
		AccessComplexity       string `json:"accessComplexity"`
		AccessVector           string `json:"accessVector"`
		Authentication         string `json:"authentication"`
		AvailabilityImpact     string `json:"availabilityImpact"`
		BaseScore              float64 `json:"baseScore"`
		ConfidentialityImpact  string `json:"confidentialityImpact"`
		ExploitabilitySubscore float64 `json:"exploitabilitySubscore"`
		generatedOn            string `json:"generatedOn"`
		ImpactSubscore         float64 `json:"impactSubscore"`
		IntegrityImpact        string `json:"integrityImpact"`
		Source                 string `json:"source"`
	} `json:"relatedMetrics"`
	RelatedVulnerabilities []struct {
		Id                  string `json:"relatedVulnerabilities"`
		VulnerabilitySource string `json:"vulnerabilitySource"`
		VulnerabilityUrl    string `json:"vulnerabilityUrl"`
	} `json:"relatedVulnerabilities"`
}

type vulnerabilities struct {
	TotalCount int             `json:"totalCount"`
	Items      []vulnerability `json:"items"`
}

type bomRiskProfile struct {
	NumberOfItems int `json:"numberOfItems"`
	Categories    struct {
		ACTIVITY      ranking `json:"ACTIVITY"`
		LICENSE       ranking `json:"LICENSE"`
		OPERATIONAL   ranking `json:"OPERATIONAL"`
		VERSION       ranking `json:"VERSION"`
		VULNERABILITY ranking `json:"VULNERABILITY"`
	} `json:"categories"`
}

type bomRows struct {
	TotalCount int `json:"totalCount"`
	Items      []struct {
		Activity struct {
			ActivityTrend           string `json:"activityTrend"`
			CommitCount12Month      int    `json:"commitCount12Month"`
			ContributorCount12Month int    `json:"contributorCount12Month"`
		} `json:"activity"`
		ComponentMatchTypes []string `json:"componentMatchTypes"`
		License             struct {
			Licenses []struct {
				Name           string `json:"name"`
				LicenseDisplay string `json:"licenseDisplay"`
			} `json:"licenses"`
		} `json:"license"`
		MatchTypes      []string `json:"matchTypes"`
		ProducerProject struct {
			Name string `json:"name"`
		} `json:producerProject"`
		ProducerRelease struct {
			Version string `json:"version"`
		} `json:"producerRelease"`
		RiskProfile bomRiskProfile `json:"riskProfile"`
		VersionRisk struct {
			NumberOfNewerReleases int    `json:"numberOfNewerReleases"`
			ReleaseDate           string `json:"releaseDate"`
		} `json:"versionRisk"`
	} `json:"items"`
}

func (h *HubServer) login() bool {
	// check if the Config entry is initialized
	if h.Config == nil {
		fmt.Printf("ERROR in HubServer no configuration available.\n")
		return false
	}

	fmt.Println(h.Config.Url)
	u, err := url.ParseRequestURI(h.Config.Url)
	if err != nil {
		fmt.Printf("ERROR : url.ParseRequestURI: %s\n", err)
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
		fmt.Printf("ERROR NewRequest:\n%s\n", err)
		return false
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded") // needed this the prevend 401 Unauthorized

	resp, err := h.client.Do(req)
	if err != nil {
		fmt.Printf("ERROR client.do\n%s\n", err)
		return false
	}
	resp.Body.Close()
	if resp.StatusCode != 204 {
		fmt.Printf("ERROR : resp status : %s\n%d\n", resp.Status, resp.StatusCode)
		return false
	}
	return true
}

type codeLocationsStruct struct {
	TotalCount int `json:"totalCount"`
	Items      []struct {
		Status  string `json:"status"`
		Type    string `json:"type"`
		Url     string `json:"url"`
		Project struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		} `json:"project"`
		Version struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		} `json;"version"`
	} `json:"items"`
}

func (h *HubServer) findCodeLocations(searchCriterea string) *codeLocationsStruct {
	searchStr := url.QueryEscape(searchCriterea)
	getStr := fmt.Sprintf("%s/api/v1/composite/codelocations/?q=%s&limit=1&includeErrors=true", h.Config.Url, searchStr)
	fmt.Println(getStr)

	var codeLocations codeLocationsStruct

	buf := h.getHubRestEndPointJson(getStr)
	if buf.Len() == 0 {
		fmt.Printf("Error no response for url : %s\n", getStr)
	}

	if err := json.Unmarshal([]byte(buf.String()), &codeLocations); err != nil {
		fmt.Printf("ERROR Unmarshall error : %s\n", err)
	}
	return &codeLocations
}

func (h *HubServer) getHubRestEndPointJson(restEndPointUrl string) *bytes.Buffer {

	buf := new(bytes.Buffer)
	resp, err := h.client.Get(restEndPointUrl)
	if err != nil {
		fmt.Printf("ERROR in client.url : %s\n get :\n%s\n", restEndPointUrl, err)
		return buf
	}
	fmt.Println(resp.Status)
	if resp.StatusCode != 200 {
		fmt.Printf("ERROR return status : %s\nurl:%s\n", resp.Status, restEndPointUrl)
		return buf
	}

	if _, err := buf.ReadFrom(resp.Body); err != nil {
		fmt.Printf("ERROR in getProjects : %s\nurl: %s\n", err, restEndPointUrl)
		return buf
	}
	defer resp.Body.Close()

	return buf

}

func (h *HubServer) getBomRows(versionId string, maxRows int) *bomRows {
	getStr := fmt.Sprintf("%s/api/v1/releases/%s/component-bom-entries?limit=%d&sortField=producerProject.name&ascending=true&offset=0&aggregationEntityType=RL&inUseOnly=true", h.Config.Url, versionId, maxRows)

	var bomRows bomRows

	buf := h.getHubRestEndPointJson(getStr)
	if buf.Len() == 0 {
		fmt.Printf("Error no response for url : %s\n", getStr)

	}

	if err := json.Unmarshal([]byte(buf.String()), &bomRows); err != nil {
		fmt.Printf("ERROR Unmarshall error : %s\n", err)
	}
	return &bomRows
}

func (h *HubServer) getVulnerabilityBom(versionId string, maxRows int) *vulnerabilityBom {
	getStr := fmt.Sprintf("%s/api/v1/releases/%s/vulnerability-bom?limit=%s&sortField=producerProject.name&ascending=true&offset=0&aggregationEntityType=RL", h.Config.Url, versionId, maxRows)

	var vulns vulnerabilityBom

	buf := h.getHubRestEndPointJson(getStr)
	if buf.Len() == 0 {
		fmt.Printf("Error getVulnerabilityBom no response for url : %s\n", getStr)

	}

	if err := json.Unmarshal([]byte(buf.String()), &vulns); err != nil {
		fmt.Printf("ERROR getVulnerabilityBom Unmarshall error : %s\n", err)
	}

	return &vulns
}

func (h *HubServer) getVulnerabilities(versionId string, channelReleaseId string, producerReleaseId string, maxRows int) *vulnerabilities {
	getStr := fmt.Sprintf("%s/api/v1/releases/%s/RL/%s/channels/%s/vulnerabilities?limit=%d&sortField=baseScore&offset=0", h.Config.Url, versionId, producerReleaseId, channelReleaseId, maxRows)

	var vulns vulnerabilities

	buf := h.getHubRestEndPointJson(getStr)
	if buf.Len() == 0 {
		fmt.Printf("Error getVulnerabilities no response for url : %s\n", getStr)

	}

	if err := json.Unmarshal([]byte(buf.String()), &vulns); err != nil {
		fmt.Printf("ERROR getVulnerabilities Unmarshall error : %s\n", err)
	}

	return &vulns
}
