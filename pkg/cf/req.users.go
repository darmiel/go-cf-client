package cf

import (
	"github.com/darmiel/go-cf-client/internal/util"
	"github.com/darmiel/go-cf-client/pkg/models"
	"net/http"
	"strings"
)

// CreateUserOptions specifies options for creating a user, including labels and annotations
// that can be applied to the user. These are optional maps that store metadata.
type CreateUserOptions struct {
	// Labels is an optional map of labels to apply to the user. Labels are key-value pairs
	// that can be used to organize and categorize users.
	Labels map[string]string

	// Annotations is an optional map of annotations to apply to the user. Annotations are
	// key-value pairs that can store additional metadata that can be used by tools and libraries.
	Annotations map[string]string
}

// CreateUser creates a new user with the specified GUID and options including labels and annotations.
// It returns the created user or an error if the creation fails.
func (req *CloudFoundryClient) CreateUser(guid string, options CreateUserOptions) (*models.User, error) {
	metadata := make(util.KV)
	if options.Labels != nil {
		metadata["labels"] = options.Labels
	}
	if options.Annotations != nil {
		metadata["annotations"] = options.Annotations
	}
	body := util.KV{
		"guid": guid,
	}
	if len(metadata) > 0 {
		body["metadata"] = metadata
	}
	return PostResult[models.User](req, "/v3/users", WithBody(body))
}

// GetUser fetches a user by their GUID. It returns the user if found or an error otherwise.
func (req *CloudFoundryClient) GetUser(userGUID string) (*models.User, error) {
	return GetResult[models.User](req, "/v3/users/"+userGUID)
}

// ListUsersOptions specifies options for listing users with various filters.
type ListUsersOptions struct {
	// GUIDFilters is a list of user GUIDFilters to filter by. Providing multiple GUIDFilters
	// will return users that match any of the specified GUIDFilters.
	GUIDFilters []string

	// UsernameFilters is a list of exact usernames to filter by. Providing multiple
	// usernames will return users that match any of the specified usernames.
	// This filter is mutually exclusive with PartialUsernameFilters.
	UsernameFilters []string

	// PartialUsernameFilters is a list of partial username strings to search by.
	// Users that contain any of the provided strings in their username will be returned.
	// This filter is mutually exclusive with UsernameFilters.
	PartialUsernameFilters []string

	// OriginFilters is a list of user origin sources to filter by. OriginFilters could
	// be sources like 'uaa', 'ldap', etc., indicating where the user was authenticated from.
	// Users authenticated from any of the specified origins will be returned.
	OriginFilters []string

	// OrderBy determines the attribute by which the results are sorted. Valid
	// values might include attributes like 'created_at' or 'updated_at', with
	// the possibility to prepend with '-' for descending order. For example,
	// '-created_at' would sort the users by the creation date in descending order.
	OrderBy OrderBy

	// LabelSelector is a query string containing a list of label selector
	// requirements. The syntax of the selector is similar to Kubernetes label
	// selectors. This allows for filtering users based on a set of labels.
	LabelSelector string
}

// ListUsers lists all users that the current user can see, optionally filtered by the provided
// ListUsersOptions. It supports filtering by multiple criteria such as GUIDFilters, usernames, partial usernames,
// origins, and labels, and allows sorting of the results.
func (req *CloudFoundryClient) ListUsers(options ListUsersOptions) ([]models.User, error) {
	queryParams := util.CreateQueryParams(util.Query{
		"guids":             strings.Join(options.GUIDFilters, ","),
		"usernames":         strings.Join(options.UsernameFilters, ","),
		"partial_usernames": strings.Join(options.PartialUsernameFilters, ","),
		"origins":           strings.Join(options.OriginFilters, ","),
		"order_by":          string(options.OrderBy),
		"label_selector":    options.LabelSelector,
	})
	return GetPaginated[models.User](req, "/v3/users", WithQueryParams(queryParams))
}

// UpdateUser updates a user's metadata including labels and annotations based on the provided GUID.
// It returns the updated user or an error if the update fails.
func (req *CloudFoundryClient) UpdateUser(guid string, metadata util.KV) (*models.User, error) {
	body := util.KV{
		"metadata": metadata,
	}
	return PatchResult[models.User](req, "/v3/users/"+guid, WithBody(body))
}

// DeleteUser deletes a user by their GUID, along with all roles associated with them.
// It returns an error if the deletion fails.
func (req *CloudFoundryClient) DeleteUser(userGUID string) error {
	return req.DeleteAndExpectStatus("/v3/users/"+userGUID, http.StatusAccepted)
}
