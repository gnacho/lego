package dondominio

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvAPIUser, EnvAPIPassword, EnvAPIURL).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAPIUser:     "user",
				EnvAPIPassword: "secret",
			},
		},
		{
			desc: "missing API user",
			envVars: map[string]string{
				EnvAPIUser:     "",
				EnvAPIPassword: "secret",
			},
			expected: "dondominio: some credentials information are missing: DONDOMINIO_API_USER",
		},
		{
			desc: "missing API password",
			envVars: map[string]string{
				EnvAPIUser:     "user",
				EnvAPIPassword: "",
			},
			expected: "dondominio: some credentials information are missing: DONDOMINIO_API_PASSWORD",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "dondominio: some credentials information are missing: DONDOMINIO_API_USER,DONDOMINIO_API_PASSWORD",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()

			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.client)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc        string
		apiUser     string
		apiPassword string
		expected    string
	}{
		{
			desc:        "success",
			apiUser:     "user",
			apiPassword: "secret",
		},
		{
			desc:        "missing API user",
			apiUser:     "",
			apiPassword: "secret",
			expected:    "dondominio: credentials missing",
		},
		{
			desc:        "missing API password",
			apiUser:     "user",
			apiPassword: "",
			expected:    "dondominio: credentials missing",
		},
		{
			desc:     "missing credentials",
			expected: "dondominio: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIUser = test.apiUser
			config.APIPassword = test.apiPassword

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.client)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestLivePresent(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present(t.Context(), envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.CleanUp(t.Context(), envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func mockBuilder() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.APIUser = "user"
			config.APIPassword = "secret"
			config.APIURL = server.URL
			config.HTTPClient = server.Client()

			return NewDNSProviderConfig(config)
		},
		servermock.CheckHeader().
			WithContentTypeFromURLEncoded(),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("POST /service/dnslist/",
			servermock.ResponseFromInternal("dnslist.json"),
			servermock.CheckForm().Strict().
				With("serviceName", "example.com").
				With("pageLength", "1000").
				With("apiuser", "user").
				With("apipasswd", "secret"),
		).
		Route("POST /service/dnsdelete/",
			servermock.ResponseFromInternal("dnsdelete.json"),
			servermock.CheckForm().Strict().
				With("serviceName", "example.com").
				With("entityID", "SRV0000214277-000396073").
				With("apiuser", "user").
				With("apipasswd", "secret"),
		).
		Route("POST /service/dnscreate/",
			servermock.ResponseFromInternal("dnscreate.json"),
			servermock.CheckForm().Strict().
				With("serviceName", "example.com").
				With("name", "_acme-challenge.example.com").
				With("type", "TXT").
				With("value", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY").
				With("apiuser", "user").
				With("apipasswd", "secret"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_Present_api_error(t *testing.T) {
	provider := mockBuilder().
		Route("POST /service/dnslist/",
			servermock.ResponseFromInternal("error.json"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.EqualError(t, err, `dondominio: list records: API error 201: Login invalid`)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("POST /service/dnslist/",
			servermock.ResponseFromInternal("dnslist.json"),
			servermock.CheckForm().Strict().
				With("serviceName", "example.com").
				With("pageLength", "1000").
				With("apiuser", "user").
				With("apipasswd", "secret"),
		).
		Route("POST /service/dnsdelete/",
			servermock.ResponseFromInternal("dnsdelete.json"),
			servermock.CheckForm().Strict().
				With("serviceName", "example.com").
				With("entityID", "SRV0000214277-000396073").
				With("apiuser", "user").
				With("apipasswd", "secret"),
		).
		Build(t)

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp_no_matching_record(t *testing.T) {
	provider := mockBuilder().
		Route("POST /service/dnslist/",
			servermock.ResponseFromInternal("dnslist_empty.json"),
		).
		Build(t)

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_Timeout(t *testing.T) {
	provider := mockBuilder().Build(t)

	timeout, interval := provider.Timeout()

	assert.Equal(t, defaultPropagationTimeout, timeout)
	assert.Equal(t, defaultPollingInterval, interval)
}
