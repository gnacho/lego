package internal

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient(server.URL, "user", "secret")
			if err != nil {
				return nil, err
			}

			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithContentTypeFromURLEncoded(),
	)
}

func TestNewClient_missing_credentials(t *testing.T) {
	testCases := []struct {
		desc        string
		apiUser     string
		apiPassword string
	}{
		{
			desc:    "missing API user",
			apiUser: "",
		},
		{
			desc:        "missing API password",
			apiPassword: "",
		},
		{
			desc: "missing credentials",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			client, err := NewClient(DefaultBaseURL, test.apiUser, test.apiPassword)

			require.EqualError(t, err, "credentials missing")
			assert.Nil(t, client)
		})
	}
}

func TestClient_ListRecords(t *testing.T) {
	client := mockBuilder().
		Route("POST /service/dnslist/",
			servermock.ResponseFromFixture("dnslist.json"),
			servermock.CheckForm().Strict().
				With("serviceName", "example.com").
				With("pageLength", "1000").
				With("apiuser", "user").
				With("apipasswd", "secret"),
		).
		Build(t)

	records, err := client.ListRecords(t.Context(), "example.com")
	require.NoError(t, err)

	expected := []Record{
		{
			EntityID: "SRV0000214277-000396072",
			Name:     "example.com",
			Type:     "A",
			Value:    "127.0.0.1",
		},
		{
			EntityID: "SRV0000214277-000396073",
			Name:     "_acme-challenge.example.com",
			Type:     "TXT",
			Value:    "old",
		},
		{
			EntityID: "SRV0000214277-000396074",
			Name:     "_acme-challenge.other.example.com",
			Type:     "TXT",
			Value:    "unrelated",
		},
	}

	assert.Equal(t, expected, records)
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /service/dnscreate/",
			servermock.ResponseFromFixture("dnscreate.json"),
			servermock.CheckForm().Strict().
				With("serviceName", "example.com").
				With("name", "_acme-challenge.example.com").
				With("type", "TXT").
				With("value", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY").
				With("apiuser", "user").
				With("apipasswd", "secret"),
		).
		Build(t)

	err := client.CreateRecord(t.Context(), "example.com", "_acme-challenge.example.com", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY")
	require.NoError(t, err)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /service/dnsdelete/",
			servermock.ResponseFromFixture("dnsdelete.json"),
			servermock.CheckForm().Strict().
				With("serviceName", "example.com").
				With("entityID", "SRV0000214277-000396073").
				With("apiuser", "user").
				With("apipasswd", "secret"),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", "SRV0000214277-000396073")
	require.NoError(t, err)
}

func TestClient_api_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /service/dnslist/",
			servermock.ResponseFromFixture("error.json"),
		).
		Build(t)

	_, err := client.ListRecords(t.Context(), "example.com")
	require.EqualError(t, err, "API error 201: Login invalid")
}

func TestClient_invalid_base_url(t *testing.T) {
	_, err := NewClient("://invalid", "user", "secret")
	require.Error(t, err)
}

func TestClient_default_base_url(t *testing.T) {
	client, err := NewClient("", "user", "secret")
	require.NoError(t, err)

	assert.Equal(t, DefaultBaseURL, client.BaseURL.String())
}
