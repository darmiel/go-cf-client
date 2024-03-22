package cf

import (
	"fmt"
	"github.com/darmiel/go-cf-client/internal/util"
	"github.com/darmiel/go-cf-client/pkg/models"
	"net/http"
	"strings"
)

// Role is a Cloud Foundry role (e.g. organization_user, space_manager, ...)
type Role string

//goland:noinspection GoUnusedConst
const (
	OrganizationUserRole           Role = "organization_user"
	OrganizationAuditorRole             = "organization_auditor"
	OrganizationManagerRole             = "organization_manager"
	OrganizationBillingManagerRole      = "organization_billing_manager"
	OrganizationSpaceAuditorRole        = "space_auditor"
	SpaceDeveloperRole                  = "space_developer"
	SpaceManagerRole                    = "space_manager"
	SpaceSupporterRole                  = "space_supporter"
)

type OrderBy string

//goland:noinspection GoUnusedConst
const (
	OrderByCreatedAtAsc  = "created_at"
	OrderByCreatedAtDesc = "-created_at"
	OrderByUpdatedAtAsc  = "updated_at"
	OrderByUpdatedAtDesc = "-updated_at"
)

var (
	CreateRoleTargetMissingErr = fmt.Errorf("either UserGUID or Username must be provided")
	InvalidRoleErr             = fmt.Errorf("invalid role")
)

type CreateRoleOptions struct {
	// UserGUID is the GUID of the user to assign the role to
	UserGUID string

	// Username is the name of the user to assign the role to
	// this requires the `set_roles_by_username` feature flag to be enabled
	Username string
}

// CreateRole creates a role for a user in an organization or space
// The role must be one of the following:
// - organization_user
// - organization_auditor
// - organization_manager
// - organization_billing_manager
// - space_auditor
// - space_developer
// - space_manager
// - space_supporter
// The targetRelationGUID must be the GUID of the organization or space
// The options must include either UserGUID or Username
func (req *CloudFoundryClient) CreateRole(
	role Role,
	spaceOrOrganizationGUID string,
	options CreateRoleOptions,
) (*models.Role, error) {
	userData := make(util.KV)
	if options.UserGUID != "" {
		userData["guid"] = options.UserGUID
	} else if options.Username != "" {
		userData["username"] = options.Username
	} else {
		return nil, CreateRoleTargetMissingErr
	}

	relationships := util.KV{
		"user": util.Data(userData),
	}

	if strings.HasPrefix(string(role), "organization_") {
		relationships["organization"] = util.DataGUID(spaceOrOrganizationGUID)
	} else if strings.HasPrefix(string(role), "space_") {
		relationships["space"] = util.DataGUID(spaceOrOrganizationGUID)
	} else {
		return nil, InvalidRoleErr
	}

	data := util.KV{
		"type":          string(role),
		"relationships": relationships,
	}
	fmt.Println("payload:", data)
	return PostResult[models.Role](req, "/v3/roles", WithBody(data))
}

// GetRole fetches a role by GUID
func (req *CloudFoundryClient) GetRole(roleGUID string) (*models.Role, error) {
	return GetResult[models.Role](req, "/v3/roles/"+roleGUID)
}

// ListRoleOptions specifies criteria for fetching roles,
// including pagination and optional filtering by type, user GUID, organization GUID, and space GUID
type ListRoleOptions struct {
	PaginationOptions

	// RoleGUIDFilters is an optional list of role GUIDFilters to filter by
	RoleGUIDFilters []string

	// RoleTypeFilters is an optional list of role types to filter by
	RoleTypeFilters []string

	// SpaceGUIDFilters is an optional list of space GUIDFilters to filter by
	SpaceGUIDFilters []string

	// OrganizationGUIDFilters is an optional list of organization GUIDFilters to filter by
	OrganizationGUIDFilters []string

	// UserGUIDFilters is an optional list of user GUIDFilters to filter by
	UserGUIDFilters []string

	// OrderBy is an optional value to sort by
	OrderBy OrderBy
}

// ListRole fetches a list of roles based on the provided fetch options
func (req *CloudFoundryClient) ListRole(options ListRoleOptions) ([]models.Role, error) {
	queryParams := createParams(options.PaginationOptions)
	if options.RoleGUIDFilters != nil {
		queryParams["guids"] = strings.Join(options.RoleGUIDFilters, ",")
	}
	if options.RoleTypeFilters != nil {
		queryParams["types"] = strings.Join(options.RoleTypeFilters, ",")
	}
	if options.SpaceGUIDFilters != nil {
		queryParams["space_guids"] = strings.Join(options.SpaceGUIDFilters, ",")
	}
	if options.OrganizationGUIDFilters != nil {
		queryParams["organization_guids"] = strings.Join(options.OrganizationGUIDFilters, ",")
	}
	if options.UserGUIDFilters != nil {
		queryParams["user_guids"] = strings.Join(options.UserGUIDFilters, ",")
	}
	if options.OrderBy != "" {
		queryParams["order_by"] = string(options.OrderBy)
	}
	return GetPaginated[models.Role](req, "/v3/roles", WithQueryParams(queryParams))
}

// DeleteRole deletes a role by GUID
func (req *CloudFoundryClient) DeleteRole(roleGUID string) error {
	return req.DeleteAndExpectStatus("/v3/roles/"+roleGUID, http.StatusAccepted)
}
