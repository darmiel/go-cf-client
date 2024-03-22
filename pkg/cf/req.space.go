package cf

import (
	"github.com/darmiel/go-cf-client/internal/util"
	"github.com/darmiel/go-cf-client/pkg/models"
	"strings"
)

// ListSpacesOptions are the options for listing spaces
type ListSpacesOptions struct {
	PaginationOptions

	// Names is a list of space names to filter by
	Names []string

	// GUIDs is a list of space GUIDs to filter by
	GUIDs []string

	// OrganizationGUIDs is a list of organization GUIDs to filter by
	OrganizationGUIDs []string

	// Labels is a list of space labels to filter by
	Labels []string
}

// ListSpaces returns a list of spaces the user has access to
func (req *CloudFoundryClient) ListSpaces(options ListSpacesOptions) ([]models.Space, error) {
	params := createParams(options.PaginationOptions)
	return GetPaginated[models.Space](req, "/v3/spaces", WithQueryParams(params))
}

// GetSpace returns a space by GUID
func (req *CloudFoundryClient) GetSpace(guid string) (*models.Space, error) {
	return GetResult[models.Space](req, "/v3/spaces/"+guid)
}

// UpdateSpaceOptions are the options for updating a space
type UpdateSpaceOptions struct {
	// Name is the new name of the space
	Name string

	// Labels is a map of labels to assign to the space
	Labels map[string]string

	// Annotations is a map of annotations to assign to the space
	Annotations map[string]string
}

// UpdateSpace updates a space by GUID
// You can update the name, labels, and annotations
func (req *CloudFoundryClient) UpdateSpace(guid string, options UpdateSpaceOptions) (*models.Space, error) {
	body := map[string]any{}
	if options.Name != "" {
		body["name"] = options.Name
	}

	metadata := map[string]any{}
	if options.Labels != nil {
		metadata["labels"] = options.Labels
	}
	if options.Annotations != nil {
		metadata["annotations"] = options.Annotations
	}
	if len(metadata) > 0 {
		body["metadata"] = metadata
	}

	return PatchResult[models.Space](req, "/v3/spaces/"+guid, WithBody(body))
}

// CreateSpaceOptions are the options for creating a space
type CreateSpaceOptions struct {
	// Labels is a map of labels to assign to the space
	Labels map[string]string

	// Annotations is a map of annotations to assign to the space
	Annotations map[string]string
}

// CreateSpace creates a space with the specified name and organization GUID
// You can also specify labels and annotations
func (req *CloudFoundryClient) CreateSpace(name, orgGUID string, options CreateSpaceOptions) (*models.Space, error) {
	body := util.KV{
		"name": name,
		"relationships": util.KV{
			"organization": util.DataGUID(orgGUID),
		},
	}
	if options.Labels != nil {
		body["labels"] = options.Labels
	}
	if options.Annotations != nil {
		body["annotations"] = options.Annotations
	}
	return PostResult[models.Space](req, "/v3/spaces", WithBody(body))
}

// ListUsersForSpaceOptions specifies the options for listing users for a space.
type ListUsersForSpaceOptions struct {
	PaginationOptions

	// GUIDFilters is a comma-delimited list of user GUIDFilters to filter by.
	GUIDFilters []string

	// UsernameFilters is a comma-delimited list of usernames to filter by. Mutually exclusive with PartialUsernameFilters.
	UsernameFilters []string

	// PartialUsernameFilters is a comma-delimited list of strings to search by. Mutually exclusive with UsernameFilters.
	PartialUsernameFilters []string

	// OriginFilters is a comma-delimited list of user origins to filter by.
	OriginFilters []string

	// OrderBy specifies the value to sort by. Defaults to ascending; prepend with - to sort descending.
	OrderBy string

	// LabelSelector contains a list of label selector requirements.
	LabelSelector string
}

// ListUsersForSpace lists all users with a role in the specified space.
// This method supports filtering by user GUIDFilters, usernames (or partial usernames), origins,
// as well as pagination and sorting options.
func (req *CloudFoundryClient) ListUsersForSpace(spaceGUID string, options ListUsersForSpaceOptions) ([]models.User, error) {
	queryParams := util.CreateQueryParams(util.Query{
		"guids":             strings.Join(options.GUIDFilters, ","),
		"usernames":         strings.Join(options.UsernameFilters, ","),
		"partial_usernames": strings.Join(options.PartialUsernameFilters, ","),
		"origins":           strings.Join(options.OriginFilters, ","),
		"order_by":          options.OrderBy,
		"label_selector":    options.LabelSelector,
	}, options.PaginationOptions.PerPage)
	return GetPaginated[models.User](req, "/v3/spaces/"+spaceGUID+"/users", WithQueryParams(queryParams))
}
