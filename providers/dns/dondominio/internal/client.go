package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-acme/lego/v5/internal/errutils"
	"github.com/go-acme/lego/v5/internal/useragent"
)

// DefaultBaseURL is the default endpoint of the DonDominio API.
const DefaultBaseURL = "https://simple-api.dondominio.net"

const (
	TXTType           = "TXT"
	dnsListPageLength = 1000
)

// Client the DonDominio API client.
// https://dondominio.dev/en/api/docs/api/
type Client struct {
	apiUser     string
	apiPassword string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(apiURL, apiUser, apiPassword string) (*Client, error) {
	if apiUser == "" || apiPassword == "" {
		return nil, errors.New("credentials missing")
	}

	if apiURL == "" {
		apiURL = DefaultBaseURL
	}

	baseURL, err := url.Parse(apiURL)
	if err != nil {
		return nil, fmt.Errorf("parse base URL: %w", err)
	}

	return &Client{
		apiUser:     apiUser,
		apiPassword: apiPassword,
		BaseURL:     baseURL,
		HTTPClient:  &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// ListRecords returns all the DNS records of a service.
// https://dondominio.dev/en/api/docs/api/#dns-zone-list-service-dnslist
func (c *Client) ListRecords(ctx context.Context, serviceName string) ([]Record, error) {
	endpoint := c.BaseURL.JoinPath("service/dnslist/")

	values := url.Values{}
	values.Set("serviceName", serviceName)
	values.Set("pageLength", strconv.Itoa(dnsListPageLength))

	req, err := c.newRequest(ctx, endpoint, values)
	if err != nil {
		return nil, err
	}

	result := new(listRecordsResponse)

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result.DNS, nil
}

// CreateRecord creates a DNS record.
// The `ttl` parameter is never sent: the API rejects the request when it is present.
// https://dondominio.dev/en/api/docs/api/#dns-zone-create-service-dnscreate
func (c *Client) CreateRecord(ctx context.Context, serviceName, name, value string) error {
	endpoint := c.BaseURL.JoinPath("service/dnscreate/")

	values := url.Values{}
	values.Set("serviceName", serviceName)
	values.Set("name", name)
	values.Set("type", TXTType)
	values.Set("value", value)

	req, err := c.newRequest(ctx, endpoint, values)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

// DeleteRecord deletes a DNS record by its entity ID.
// https://dondominio.dev/en/api/docs/api/#dns-zone-delete-service-dnsdelete
func (c *Client) DeleteRecord(ctx context.Context, serviceName, entityID string) error {
	endpoint := c.BaseURL.JoinPath("service/dnsdelete/")

	values := url.Values{}
	values.Set("serviceName", serviceName)
	values.Set("entityID", entityID)

	req, err := c.newRequest(ctx, endpoint, values)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		raw, _ := io.ReadAll(resp.Body)

		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	response := &APIResponse{}

	err = json.Unmarshal(raw, response)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	if !response.Success {
		return fmt.Errorf("API error %d: %s", response.ErrorCode, response.ErrorCodeMsg)
	}

	if result == nil {
		return nil
	}

	err = json.Unmarshal(response.ResponseData, result)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, response.ResponseData, err)
	}

	return nil
}

func (c *Client) newRequest(ctx context.Context, endpoint *url.URL, values url.Values) (*http.Request, error) {
	// The DonDominio API expects the credentials as form values of every request.
	values.Set("apiuser", c.apiUser)
	values.Set("apipasswd", c.apiPassword)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), strings.NewReader(values.Encode()))
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return req, nil
}
