package models

import "time"

// Space is a Cloud Foundry space
type Space struct {
	Guid          string    `json:"guid"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Name          string    `json:"name"`
	Relationships struct {
		Organization struct {
			Data struct {
				Guid string `json:"guid"`
			} `json:"data"`
		} `json:"organization"`
		Quota struct {
			Data interface{} `json:"data"`
		} `json:"quota"`
	} `json:"relationships"`
	Links struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
		Features struct {
			Href string `json:"href"`
		} `json:"features"`
		Organization struct {
			Href string `json:"href"`
		} `json:"organization"`
		ApplyManifest struct {
			Href   string `json:"href"`
			Method string `json:"method"`
		} `json:"apply_manifest"`
	} `json:"links"`
	Metadata struct {
		Labels struct {
		} `json:"labels"`
		Annotations struct {
		} `json:"annotations"`
	} `json:"metadata"`
}
