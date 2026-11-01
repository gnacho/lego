// Package dondominio implements a DNS provider for solving the DNS-01 challenge using DonDominio.
package dondominio

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/platform/env"
	"github.com/go-acme/lego/v5/providers/dns/dondominio/internal"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
)

// Environment variables names.
const (
	envNamespace = "DONDOMINIO_"

	EnvAPIUser     = envNamespace + "API_USER"
	EnvAPIPassword = envNamespace + "API_PASSWORD"
	EnvAPIURL      = envNamespace + "API_URL"

	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

const (
	defaultPropagationTimeout = 5 * time.Minute
	defaultPollingInterval    = 10 * time.Second
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIUser     string
	APIPassword string
	APIURL      string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		APIURL:             env.GetOrDefaultString(EnvAPIURL, internal.DefaultBaseURL),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, defaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, defaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for DonDominio.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIUser, EnvAPIPassword)
	if err != nil {
		return nil, fmt.Errorf("dondominio: %w", err)
	}

	config := NewDefaultConfig()
	config.APIUser = values[EnvAPIUser]
	config.APIPassword = values[EnvAPIPassword]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for DonDominio.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("dondominio: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.APIURL, config.APIUser, config.APIPassword)
	if err != nil {
		return nil, fmt.Errorf("dondominio: %w", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	client.HTTPClient = clientdebug.Wrap(client.HTTPClient)

	return &DNSProvider{
		config: config,
		client: client,
	}, nil
}

// Present creates a TXT record using the specified parameters.
// The DonDominio API has no upsert: any previous TXT record with the same name is deleted first.
func (d *DNSProvider) Present(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("dondominio: could not find zone for domain %q: %w", domain, err)
	}

	serviceName := dns01.UnFqdn(authZone)

	err = d.removeRecords(ctx, serviceName, dns01.UnFqdn(info.EffectiveFQDN))
	if err != nil {
		return fmt.Errorf("dondominio: %w", err)
	}

	err = d.client.CreateRecord(ctx, serviceName, dns01.UnFqdn(info.EffectiveFQDN), info.Value)
	if err != nil {
		return fmt.Errorf("dondominio: create record: %w", err)
	}

	return nil
}

// CleanUp removes the TXT records matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dns01.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("dondominio: could not find zone for domain %q: %w", domain, err)
	}

	err = d.removeRecords(ctx, dns01.UnFqdn(authZone), dns01.UnFqdn(info.EffectiveFQDN))
	if err != nil {
		return fmt.Errorf("dondominio: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) removeRecords(ctx context.Context, serviceName, recordName string) error {
	records, err := d.client.ListRecords(ctx, serviceName)
	if err != nil {
		return fmt.Errorf("list records: %w", err)
	}

	for _, record := range records {
		if record.Type != internal.TXTType || !strings.EqualFold(record.Name, recordName) {
			continue
		}

		err = d.client.DeleteRecord(ctx, serviceName, record.EntityID)
		if err != nil {
			return fmt.Errorf("delete record %s: %w", record.EntityID, err)
		}
	}

	return nil
}
