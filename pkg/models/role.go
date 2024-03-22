package models

import "time"

type Role struct {
	Guid          string    `json:"guid"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Type          string    `json:"type"`
	Relationships struct {
		User struct {
			Data struct {
				Guid string `json:"guid"`
			} `json:"data"`
		} `json:"user"`
		Organization struct {
			Data struct {
				Guid string `json:"guid"`
			} `json:"data"`
		} `json:"organization"`
		Space struct {
			Data struct {
				Guid string `json:"guid"`
			} `json:"data"`
		} `json:"space"`
	} `json:"relationships"`
	Links struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
		User struct {
			Href string `json:"href"`
		} `json:"user"`
		Organization struct {
			Href string `json:"href"`
		} `json:"organization"`
	} `json:"links"`
}

// GetUserID returns the GUID of the user associated with the role
func (r Role) GetUserID() string {
	return r.Relationships.User.Data.Guid
}

// IsOrganizationRole returns true if the role is associated with an organization
func (r Role) IsOrganizationRole() bool {
	return r.Relationships.Organization.Data.Guid != ""
}

// GetOrganizationID returns the GUID of the organization associated with the role
func (r Role) GetOrganizationID() (string, bool) {
	if !r.IsOrganizationRole() {
		return "", false
	}
	return r.Relationships.Organization.Data.Guid, true
}

// IsSpaceRole returns true if the role is associated with a space
func (r Role) IsSpaceRole() bool {
	return r.Relationships.Space.Data.Guid != ""
}

// GetSpaceID returns the GUID of the space associated with the role
func (r Role) GetSpaceID() (string, bool) {
	if !r.IsSpaceRole() {
		return "", false
	}
	return r.Relationships.Space.Data.Guid, true
}
