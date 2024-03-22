package cf

import (
	"github.com/darmiel/go-cf-client/pkg/models"
	"strings"
)

// ListOrganizationsOptions specifies criteria for fetching organizations,
// including pagination and optional filtering by names or GUIDFilters.
type ListOrganizationsOptions struct {
	PaginationOptions

	// NameFilters is an optional list of organization names to filter by
	NameFilters []string

	// GUIDFilters is an optional list of organization GUIDFilters to filter by
	GUIDFilters []string
}

// ListOrganizations fetches a list of organizations based on the provided fetch options,
// which include pagination and filters by names and GUIDFilters.
func (req *CloudFoundryClient) ListOrganizations(options ListOrganizationsOptions) ([]models.Organization, error) {
	queryParams := createParams(options.PaginationOptions)
	if options.NameFilters != nil {
		queryParams["names"] = strings.Join(options.NameFilters, ",")
	}
	if options.GUIDFilters != nil {
		queryParams["guids"] = strings.Join(options.GUIDFilters, ",")
	}
	return GetPaginated[models.Organization](req, "/v3/organizations", WithQueryParams(queryParams))
}
