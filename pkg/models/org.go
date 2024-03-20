package models

import "time"

// Organization is a Cloud Foundry organization
type Organization struct {
	Guid          string    `json:"guid"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Name          string    `json:"name"`
	Suspended     bool      `json:"suspended"`
	Relationships struct {
		Quota struct {
			Data struct {
				Guid string `json:"guid"`
			} `json:"data"`
		} `json:"quota"`
	} `json:"relationships"`
	Metadata struct {
		Labels struct {
		} `json:"labels"`
		Annotations struct {
		} `json:"annotations"`
	} `json:"metadata"`
	Links struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
		Domains struct {
			Href string `json:"href"`
		} `json:"domains"`
		DefaultDomain struct {
			Href string `json:"href"`
		} `json:"default_domain"`
		Quota struct {
			Href string `json:"href"`
		} `json:"quota"`
	} `json:"links"`
}
