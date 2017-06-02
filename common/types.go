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
	"net/http"
	"time"
)

type hubServer struct {
	client *http.Client
	config *HubConfig
}

type HubConfig struct {
	Url      string `json:"url"`
	Host     string `json:"hubhost"`
	Port     string `json:"port"`
	Scheme   string `json:"scheme"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type CodeLocationStruct struct {
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

type CodeLocationsStruct struct {
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

type ScanSummaryStruct struct {
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
