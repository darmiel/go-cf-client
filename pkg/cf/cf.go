package cf

import (
	"encoding/json"
	"fmt"
	"github.com/darmiel/go-cf-client/internal/util"
	"github.com/go-resty/resty/v2"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	// TokenExpirySafetyMargin is the lastAuthTime to subtract from the authToken expiration lastAuthTime to ensure
	// that the authToken is refreshed before it expires
	TokenExpirySafetyMargin = 5 * time.Second

	// MaxPaginationPages is the maximum number of pages to fetch in a paginated request
	MaxPaginationPages = 100

	// MaxItemsPerPage is the maximum number of items to return per page
	// according to https://v3-apidocs.cloudfoundry.org/version/3.158.0/index.html#list-organizations
	// the maximum per_page is 5000
	MaxItemsPerPage = 5000
)

type (
	// RelativePath is a type that represents a relative path
	RelativePath string

	// AbsolutePath is a type that represents an absolute path
	AbsolutePath string

	// RequestModifier is a type that represents a request modifier
	// You can use this type to modify the request before it is executed
	RequestModifier func(r *resty.Request)
)

// CloudFoundryConfig is the configuration for the Cloud Foundry httpClient
type CloudFoundryConfig struct {
	Username     string `json:"username,omitempty"`
	Password     string `json:"password,omitempty"`
	Organization string `json:"organization,omitempty"`
	APIEndpoint  string `json:"api_endpoint,omitempty"`

	// Authorization and Authentication
	AuthEndpoint      string `json:"auth_endpoint,omitempty"`
	OAuthClientID     string `json:"oauth_client_id,omitempty"`
	OAuthClientSecret string `json:"oauth_client_secret,omitempty"`

	UAAEndpoint string `json:"uaa_endpoint,omitempty"`
}

// AuthTokenInfo is the response from the authToken endpoint and will be returned after a successful authentication
type AuthTokenInfo struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	JTI          string `json:"jti"`
}

// CloudFoundryClient is a struct that manages the authToken and refreshes it if necessary
type CloudFoundryClient struct {
	authToken    *AuthTokenInfo
	lastAuthTime time.Time
	config       *CloudFoundryConfig
	httpClient   *resty.Client
}

// GetTokenInfo returns a copy of the authToken info
func (req *CloudFoundryClient) GetTokenInfo() AuthTokenInfo {
	return *req.authToken
}

// CloudFoundryError is a struct that represents an error response from the server
type CloudFoundryError struct {
	Detail string `json:"detail"`
	Title  string `json:"title"`
	Code   int    `json:"code"`
}

// Error returns a string representation of the error
func (c CloudFoundryError) Error() string {
	return fmt.Sprintf("CF-Error[%d] %s: %s", c.Code, c.Title, c.Detail)
}

// href is a struct that represents a href in a paginated response
type href struct {
	Href string `json:"href"`
}

// CloudFoundryPaginatedResult is a struct that represents a paginated response from the server
type CloudFoundryPaginatedResult[T any] struct {
	Pagination struct {
		TotalResults int   `json:"total_results"`
		TotalPages   int   `json:"total_pages"`
		First        *href `json:"first"`
		Last         *href `json:"last"`
		Next         *href `json:"next"`
		Previous     *href `json:"previous"`
	}
	Resources []T
}

// PaginationOptions is a struct that represents the options for the number of items to return per page
type PaginationOptions struct {
	// PerPage is the number of organizations to return per page
	PerPage int
}

// LoadCloudFoundryConfig loads the config from the `config.json` file
func LoadCloudFoundryConfig(fileName string) (res *CloudFoundryConfig, err error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &res)
	return
}

// newAuthenticatedRequest is a wrapper around resty's R method which automatically refreshes the authToken if necessary
func (req *CloudFoundryClient) newAuthenticatedRequest() (*resty.Request, error) {
	// refresh the authToken if necessary
	if req.TokenIsExpired() {
		if err := req.RefreshToken(); err != nil {
			return nil, err
		}
	}
	// fill in authentication headers
	return req.httpClient.R().
		SetAuthScheme(req.authToken.TokenType).
		SetAuthToken(req.authToken.AccessToken), nil
}

// resolveEndpointURL returns the full URL for the given path
// :param path: The path to the endpoint. This can be a string, AbsolutePath or RelativePath.
// :return: The full URL
func (cfg *CloudFoundryConfig) resolveEndpointURL(path any) string {
	switch t := path.(type) {
	case AbsolutePath:
		return string(t)
	case RelativePath:
		return cfg.APIEndpoint + string(t)
	case string:
		return cfg.resolveEndpointURL(RelativePath(t))
	}
	panic("invalid path type")
}

// WithRequestModifiers is a request modifier that applies all given modifiers to the request
func WithRequestModifiers(modifiers ...RequestModifier) RequestModifier {
	return func(r *resty.Request) {
		applyRequestModifiers(r, modifiers...)
	}
}

// WithQueryParams is a request modifier that sets the query parameters for a request
func WithQueryParams(params map[string]string) RequestModifier {
	return func(r *resty.Request) {
		for key, value := range params {
			r.SetQueryParam(key, value)
		}
	}
}

// WithBody is a request modifier that sets the body for a request
func WithBody(body any) RequestModifier {
	return func(r *resty.Request) {
		r.SetBody(body)
	}
}

// WithResult is a request modifier that sets the result type for a request
func WithResult[T any]() RequestModifier {
	return func(r *resty.Request) {
		r.SetResult(new(T))
	}
}

// SendRequest is a wrapper around newAuthenticatedRequest which automatically sets the endpoint
// :param method: The HTTP method to use
// :param path: The path to the endpoint. This can be a string, AbsolutePath or RelativePath
// :param modifier: One or more optional modifiers that will be called with the request object before it is executed
// :return: The response from the server
func (req *CloudFoundryClient) SendRequest(
	method string,
	path any,
	modifiers ...RequestModifier,
) (*resty.Response, error) {
	// newAuthenticatedRequest automatically fills in authentication headers and refreshes the authToken if necessary
	r, err := req.newAuthenticatedRequest()
	if err != nil {
		return nil, err
	}
	applyRequestModifiers(r, modifiers...)
	resp, err := r.Execute(method, req.config.resolveEndpointURL(path))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return nil, parseErrorResponse(resp.Body())
	}
	return resp, nil
}

// SendRequestAndParseResult is a wrapper around SendRequest which automatically sets the result type
// :param req: The requester to use
// :param method: The HTTP method to use
// :param path: The path to the endpoint. This can be a string, AbsolutePath or RelativePath
// :param modifier: One or more optional modifiers that will be called with the request object before it is executed
// :return: The response from the server, parsed as the given type
func SendRequestAndParseResult[T any](
	req *CloudFoundryClient,
	method string,
	path any,
	modifiers ...RequestModifier,
) (*T, error) {
	resp, err := req.SendRequest(method, path, WithResult[T](), WithRequestModifiers(modifiers...))
	if err != nil {
		return nil, err
	}
	return resp.Result().(*T), nil
}

// FetchAllPages is a wrapper around SendRequest which automatically fetches all pages of a paginated response
// :param req: The requester to use
// :param method: The HTTP method to use
// :param path: The path to the endpoint. This can be a string, AbsolutePath or RelativePath
// :param modifier: One or more optional modifiers that will be called with the request object before it is executed
// :return: The resources from all pages
func FetchAllPages[T any](
	req *CloudFoundryClient,
	method string,
	path any,
	modifiers ...RequestModifier,
) ([]T, error) {
	var result []T

	currentPath := path
	for i := 0; i < MaxPaginationPages; i++ {
		paginated, err := SendRequestAndParseResult[CloudFoundryPaginatedResult[T]](
			req, method, currentPath, modifiers...,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, paginated.Resources...)

		if paginated.Pagination.Next != nil {
			currentPath = AbsolutePath(paginated.Pagination.Next.Href)
		} else {
			break
		}
	}

	return result, nil
}

// Get is a wrapper around SendRequest which automatically sets the method to GET
// :param path: The path to the endpoint. This can be a string, AbsolutePath or RelativePath
// :param modifiers: One or more optional modifiers that will be called with the request object before it is executed
// :return: The response from the server
func (req *CloudFoundryClient) Get(path string, modifiers ...RequestModifier) (*resty.Response, error) {
	return req.SendRequest(resty.MethodGet, path, modifiers...)
}

// GetPaginated is a wrapper around FetchAllPages which automatically sets the method to GET
// :param path: The path to the endpoint. This can be a string, AbsolutePath or RelativePath
// :param modifiers: One or more optional modifiers that will be called with the request object before it is executed
// :return: The resources from all pages
func GetPaginated[T any](req *CloudFoundryClient, path string, modifiers ...RequestModifier) ([]T, error) {
	return FetchAllPages[T](req, resty.MethodGet, path, modifiers...)
}

// GetResult is a wrapper around SendRequestAndParseResult which automatically sets the method to GET
// :param path: The path to the endpoint. This can be a string, AbsolutePath or RelativePath
// :param modifiers: One or more optional modifiers that will be called with the request object before it is executed
// :return: The response from the server, parsed as the given type
func GetResult[T any](req *CloudFoundryClient, path string, modifiers ...RequestModifier) (*T, error) {
	return SendRequestAndParseResult[T](req, resty.MethodGet, path, modifiers...)
}

// Post is a wrapper around SendRequest which automatically sets the method to POST
// :param path: The path to the endpoint. This can be a string, AbsolutePath or RelativePath
// :param modifiers: One or more optional modifiers that will be called with the request object before it is executed
// :return: The response from the server
func (req *CloudFoundryClient) Post(path string, modifiers ...RequestModifier) (*resty.Response, error) {
	return req.SendRequest(resty.MethodPost, path, modifiers...)
}

// PostResult is a wrapper around SendRequestAndParseResult which automatically sets the method to Post
// :param path: The path to the endpoint. This can be a string, AbsolutePath or RelativePath
// :param modifiers: One or more optional modifiers that will be called with the request object before it is executed
// :return: The response from the server, parsed as the given type
func PostResult[T any](req *CloudFoundryClient, path string, modifiers ...RequestModifier) (*T, error) {
	return SendRequestAndParseResult[T](req, resty.MethodPost, path, modifiers...)
}

// PatchResult is a wrapper around SendRequestAndParseResult which automatically sets the method to Patch
// :param path: The path to the endpoint. This can be a string, AbsolutePath or RelativePath
// :param modifiers: One or more optional modifiers that will be called with the request object before it is executed
// :return: The response from the server, parsed as the given type
func PatchResult[T any](req *CloudFoundryClient, path string, modifiers ...RequestModifier) (*T, error) {
	return SendRequestAndParseResult[T](req, resty.MethodPatch, path, modifiers...)
}

// DeleteAndExpectStatus is a wrapper around SendRequestAndParseResult which automatically sets the method to Delete
// :param path: The path to the endpoint. This can be a string, AbsolutePath or RelativePath
// :param modifiers: One or more optional modifiers that will be called with the request object before it is executed
// :return: The response from the server, parsed as the given type
func (req *CloudFoundryClient) DeleteAndExpectStatus(path string, expectedStatus int, modifiers ...RequestModifier) error {
	resp, err := req.SendRequest(resty.MethodDelete, path, modifiers...)
	if err != nil {
		return err
	}
	if resp.StatusCode() != expectedStatus {
		return fmt.Errorf("expected %d, got %d", expectedStatus, resp.StatusCode())
	}
	return nil
}

// applyRequestModifiers applies the given modifiers to the request
func applyRequestModifiers(r *resty.Request, modifiers ...RequestModifier) {
	for _, c := range modifiers {
		c(r)
	}
}

// createParams creates the query parameters for a paginated request
func createParams(ppo PaginationOptions) map[string]string {
	m := make(map[string]string)
	if ppo.PerPage > 0 {
		m["per_page"] = strconv.Itoa(util.Clamp(ppo.PerPage, 1, MaxItemsPerPage))
	} else {
		m["per_page"] = strconv.Itoa(MaxItemsPerPage)
	}
	return m
}

// parseErrorResponse returns an error from the given body
func parseErrorResponse(body []byte) error {
	multiError := struct {
		Errors []CloudFoundryError `json:"errors"`
	}{}
	if err := json.Unmarshal(body, &multiError); err != nil {
		var singleError CloudFoundryError
		if err = json.Unmarshal(body, &singleError); err != nil {
			// if we can't parse the error, just return the raw body
			return fmt.Errorf("api error. body: %s", body)
		}
		multiError = struct {
			Errors []CloudFoundryError `json:"errors"`
		}{
			Errors: []CloudFoundryError{singleError},
		}
	}
	var errors []string
	for _, e := range multiError.Errors {
		errors = append(errors, e.Error())
	}
	return fmt.Errorf("api error. errors: %s", strings.Join(errors, ", "))
}
