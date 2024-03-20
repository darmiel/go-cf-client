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

// TokenTimeJitter is the time to subtract from the token expiration time to ensure that the token is refreshed before it expires
const (
	TokenTimeJitter = 5 * time.Second
	MaxPages        = 100

	// MaxPerPage is the maximum number of items to return per page
	// according to https://v3-apidocs.cloudfoundry.org/version/3.158.0/index.html#list-organizations
	// the maximum per_page is 5000
	MaxPerPage = 5000
)

// Config is the configuration for the Cloud Foundry client
type Config struct {
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

// TokenResponse is the response from the token endpoint and will be returned after a successful authentication
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	JTI          string `json:"jti"`
}

// Requester is a struct that manages the token and refreshes it if necessary
type Requester struct {
	token  *TokenResponse
	time   time.Time
	config *Config
}

// LoadConfig loads the config from the `config.json` file
func LoadConfig() (res *Config, err error) {
	data, err := os.ReadFile("config.json")
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &res)
	return
}

// createRequest is a wrapper around resty's R method which automatically refreshes the token if necessary
func (req *Requester) createRequest() (*resty.Request, error) {
	// refresh the token if necessary
	if req.IsExpired() {
		if err := req.Refresh(); err != nil {
			return nil, err
		}
	}
	// fill in authentication headers
	return resty.New().R().
		SetAuthScheme(req.token.TokenType).
		SetAuthToken(req.token.AccessToken), nil
}

type (
	RelativePath string
	AbsolutePath string
)

// getURL returns the full URL for the given path
// :param path: The path to the endpoint. This can be a string, AbsolutePath or RelativePath.
// :return: The full URL
func (c *Config) getURL(path any) string {
	switch t := path.(type) {
	case AbsolutePath:
		return string(t)
	case RelativePath:
		return c.APIEndpoint + string(t)
	case string:
		return c.getURL(RelativePath(t))
	}
	panic("invalid path type")
}

func applyCallbacks(r *resty.Request, callbacks ...func(r *resty.Request)) {
	for _, c := range callbacks {
		c(r)
	}
}

func WithCallbacks(callbacks ...func(r *resty.Request)) func(r *resty.Request) {
	return func(r *resty.Request) {
		applyCallbacks(r, callbacks...)
	}
}

// CFError is a struct that represents an error response from the server
type CFError struct {
	Detail string `json:"detail"`
	Title  string `json:"title"`
	Code   int    `json:"code"`
}

// Error returns a string representation of the error
func (c CFError) Error() string {
	return fmt.Sprintf("CF-Error[%d] %s: %s", c.Code, c.Title, c.Detail)
}

// getErrorFromBody returns an error from the given body
func getErrorFromBody(body []byte) error {
	multiError := struct {
		Errors []CFError `json:"errors"`
	}{}
	if err := json.Unmarshal(body, &multiError); err != nil {
		var singleError CFError
		if err = json.Unmarshal(body, &singleError); err != nil {
			// if we can't parse the error, just return the raw body
			return fmt.Errorf("api error. body: %s", body)
		}
		multiError = struct {
			Errors []CFError `json:"errors"`
		}{
			Errors: []CFError{singleError},
		}
	}
	var errors []string
	for _, e := range multiError.Errors {
		errors = append(errors, e.Error())
	}
	return fmt.Errorf("api error. errors: %s", strings.Join(errors, ", "))
}

// Execute is a wrapper around createRequest which automatically sets the endpoint
// :param method: The HTTP method to use
// :param path: The path to the endpoint. This can be a string, AbsolutePath or RelativePath
// :param callback: One or more optional callbacks that will be called with the request object before it is executed
// :return: The response from the server
func (req *Requester) Execute(method string, path any, callback ...func(r *resty.Request)) (*resty.Response, error) {
	// createRequest automatically fills in authentication headers and refreshes the token if necessary
	r, err := req.createRequest()
	if err != nil {
		return nil, err
	}
	applyCallbacks(r, callback...)
	resp, err := r.Execute(method, req.config.getURL(path))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return nil, getErrorFromBody(resp.Body())
	}
	return resp, nil
}

// href is a struct that represents a href in a paginated response
type href struct {
	Href string `json:"href"`
}

// PaginatedResponse is a struct that represents a paginated response from the server
type PaginatedResponse[T any] struct {
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

// ExecuteResult is a wrapper around Execute which automatically sets the result type
// :param req: The requester to use
// :param method: The HTTP method to use
// :param path: The path to the endpoint. This can be a string, AbsolutePath or RelativePath
// :param callback: One or more optional callbacks that will be called with the request object before it is executed
// :return: The response from the server, parsed as the given type
func ExecuteResult[T any](req *Requester, method string, path any, callback ...func(r *resty.Request)) (*T, error) {
	resp, err := req.Execute(method, path, WithResult[T](), WithCallbacks(callback...))
	if err != nil {
		return nil, err
	}
	return resp.Result().(*T), nil
}

// ExecutePaginated is a wrapper around Execute which automatically fetches all pages of a paginated response
// :param req: The requester to use
// :param method: The HTTP method to use
// :param path: The path to the endpoint. This can be a string, AbsolutePath or RelativePath
// :param callback: One or more optional callbacks that will be called with the request object before it is executed
// :return: The resources from all pages
func ExecutePaginated[T any](req *Requester, method string, path any, callback ...func(r *resty.Request)) ([]T, error) {
	var result []T

	currentPath := path
	for i := 0; i < MaxPages; i++ {
		paginated, err := ExecuteResult[PaginatedResponse[T]](req, method, currentPath, callback...)
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

// Get is a wrapper around Execute which automatically sets the method to GET
// :param path: The path to the endpoint. This can be a string, AbsolutePath or RelativePath
// :param callback: One or more optional callbacks that will be called with the request object before it is executed
// :return: The response from the server
func (req *Requester) Get(path string, callback ...func(r *resty.Request)) (*resty.Response, error) {
	return req.Execute(resty.MethodGet, path, callback...)
}

// GetPaginated is a wrapper around ExecutePaginated which automatically sets the method to GET
// :param path: The path to the endpoint. This can be a string, AbsolutePath or RelativePath
// :param callback: One or more optional callbacks that will be called with the request object before it is executed
// :return: The resources from all pages
func GetPaginated[T any](req *Requester, path string, callback ...func(r *resty.Request)) ([]T, error) {
	return ExecutePaginated[T](req, resty.MethodGet, path, callback...)
}

// GetResult is a wrapper around ExecuteResult which automatically sets the method to GET
// :param path: The path to the endpoint. This can be a string, AbsolutePath or RelativePath
// :param callback: One or more optional callbacks that will be called with the request object before it is executed
// :return: The response from the server, parsed as the given type
func GetResult[T any](req *Requester, path string, callback ...func(r *resty.Request)) (*T, error) {
	return ExecuteResult[T](req, resty.MethodGet, path, callback...)
}

// Post is a wrapper around Execute which automatically sets the method to POST
// :param path: The path to the endpoint. This can be a string, AbsolutePath or RelativePath
// :param callback: One or more optional callbacks that will be called with the request object before it is executed
// :return: The response from the server
func (req *Requester) Post(path string, callback ...func(r *resty.Request)) (*resty.Response, error) {
	return req.Execute(resty.MethodPost, path, callback...)
}

// PostResult is a wrapper around ExecuteResult which automatically sets the method to Post
// :param path: The path to the endpoint. This can be a string, AbsolutePath or RelativePath
// :param callback: One or more optional callbacks that will be called with the request object before it is executed
// :return: The response from the server, parsed as the given type
func PostResult[T any](req *Requester, path string, callback ...func(r *resty.Request)) (*T, error) {
	return ExecuteResult[T](req, resty.MethodPost, path, callback...)
}

// PatchResult is a wrapper around ExecuteResult which automatically sets the method to Patch
// :param path: The path to the endpoint. This can be a string, AbsolutePath or RelativePath
// :param callback: One or more optional callbacks that will be called with the request object before it is executed
// :return: The response from the server, parsed as the given type
func PatchResult[T any](req *Requester, path string, callback ...func(r *resty.Request)) (*T, error) {
	return ExecuteResult[T](req, resty.MethodPatch, path, callback...)
}

// WithQueryParams is a callback that sets the query parameters for a request
func WithQueryParams(params map[string]string) func(r *resty.Request) {
	return func(r *resty.Request) {
		for key, value := range params {
			r.SetQueryParam(key, value)
		}
	}
}

func WithBody(body any) func(r *resty.Request) {
	return func(r *resty.Request) {
		r.SetBody(body)
	}
}

func WithResult[T any]() func(r *resty.Request) {
	return func(r *resty.Request) {
		r.SetResult(new(T))
	}
}

// createParams creates the query parameters for a paginated request
func createParams(ppo PerPageOptions) map[string]string {
	m := make(map[string]string)
	if ppo.PerPage > 0 {
		m["per_page"] = strconv.Itoa(util.Clamp(ppo.PerPage, 1, MaxPerPage))
	} else {
		m["per_page"] = strconv.Itoa(MaxPerPage)
	}
	return m
}

// PerPageOptions is a struct that represents the options for the number of items to return per page
type PerPageOptions struct {
	// PerPage is the number of organizations to return per page
	PerPage int
}
