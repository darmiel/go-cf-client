package btp

import (
	"btp-service/pkg/models"
	"strings"
)

type ListOrganizationsOptions struct {
	PerPageOptions

	// Names is a list of organization names to filter by
	Names []string

	// GUIDs is a list of organization GUIDs to filter by
	GUIDs []string
}

// ListOrganizations returns a list of organizations
func (req *Requester) ListOrganizations(options ListOrganizationsOptions) ([]models.Organization, error) {
	params := createParams(options.PerPageOptions)
	if options.Names != nil {
		params["names"] = strings.Join(options.Names, ",")
	}
	if options.GUIDs != nil {
		params["guids"] = strings.Join(options.GUIDs, ",")
	}
	return GetPaginated[models.Organization](req, "/v3/organizations", WithQueryParams(params))
}
