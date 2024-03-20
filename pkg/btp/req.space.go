package btp

import "btp-service/pkg/models"

type ListSpacesOptions struct {
	PerPageOptions

	// Names is a list of space names to filter by
	Names []string

	// GUIDs is a list of space GUIDs to filter by
	GUIDs []string

	// OrganizationGUIDs is a list of organization GUIDs to filter by
	OrganizationGUIDs []string

	// Labels is a list of space labels to filter by
	Labels []string
}

func (req *Requester) ListSpaces(options ListSpacesOptions) ([]models.Space, error) {
	params := createParams(options.PerPageOptions)
	return GetPaginated[models.Space](req, "/v3/spaces", WithQueryParams(params))
}
