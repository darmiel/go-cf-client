package btp

import (
	"btp-service/pkg/models"
)

// ListSpacesOptions are the options for listing spaces
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

// ListSpaces returns a list of spaces the user has access to
func (req *Requester) ListSpaces(options ListSpacesOptions) ([]models.Space, error) {
	params := createParams(options.PerPageOptions)
	return GetPaginated[models.Space](req, "/v3/spaces", WithQueryParams(params))
}

// GetSpace returns a space by GUID
func (req *Requester) GetSpace(guid string) (*models.Space, error) {
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
func (req *Requester) UpdateSpace(guid string, options UpdateSpaceOptions) (*models.Space, error) {
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

func (req *Requester) CreateSpace(name, orgGUID string, options CreateSpaceOptions) (*models.Space, error) {
	body := map[string]any{
		"name": name,
		"relationships": map[string]any{
			"organization": map[string]any{
				"data": map[string]string{
					"guid": orgGUID,
				},
			},
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
