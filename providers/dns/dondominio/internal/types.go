package internal

import "encoding/json"

// APIResponse is the envelope of every DonDominio API response.
// https://dondominio.dev/en/api/docs/api/#simple-post-interface
type APIResponse struct {
	Success      bool            `json:"success"`
	ErrorCode    int             `json:"errorCode"`
	ErrorCodeMsg string          `json:"errorCodeMsg"`
	ResponseData json.RawMessage `json:"responseData"`
}

// Record is a DNS record of a service.
type Record struct {
	EntityID string `json:"entityID"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Value    string `json:"value"`
}

// listRecordsResponse is the `responseData` of the `service/dnslist/` action.
// https://dondominio.dev/en/api/docs/api/#dns-zone-list-service-dnslist
type listRecordsResponse struct {
	QueryInfo QueryInfo `json:"queryInfo"`
	DNS       []Record  `json:"dns"`
}

// QueryInfo describes the pagination of a list response.
type QueryInfo struct {
	Page       int `json:"page"`
	PageLength int `json:"pageLength"`
	Results    int `json:"results"`
	Total      int `json:"total"`
}
