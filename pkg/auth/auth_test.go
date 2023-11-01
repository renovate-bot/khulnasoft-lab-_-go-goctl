package auth

import (
	"testing"

	"github.com/khulnasoft-lab/go-goctl/v2/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestTokenForHost(t *testing.T) {
	tests := []struct {
		name                  string
		host                  string
		githubToken           string
		githubEnterpriseToken string
		goctlToken               string
		goctlEnterpriseToken     string
		config                *config.Config
		wantToken             string
		wantSource            string
		wantNotFound          bool
	}{
		{
			name:         "token for github.com with no env tokens and no config token",
			host:         "github.com",
			config:       testNoHostsConfig(),
			wantToken:    "",
			wantSource:   "oauth_token",
			wantNotFound: true,
		},
		{
			name:         "token for enterprise.com with no env tokens and no config token",
			host:         "enterprise.com",
			config:       testNoHostsConfig(),
			wantToken:    "",
			wantSource:   "oauth_token",
			wantNotFound: true,
		},
		{
			name:        "token for github.com with GOCTL_TOKEN, GITHUB_TOKEN, and config token",
			host:        "github.com",
			goctlToken:     "GOCTL_TOKEN",
			githubToken: "GITHUB_TOKEN",
			config:      testHostsConfig(),
			wantToken:   "GOCTL_TOKEN",
			wantSource:  "GOCTL_TOKEN",
		},
		{
			name:        "token for github.com with GITHUB_TOKEN, and config token",
			host:        "github.com",
			githubToken: "GITHUB_TOKEN",
			config:      testHostsConfig(),
			wantToken:   "GITHUB_TOKEN",
			wantSource:  "GITHUB_TOKEN",
		},
		{
			name:       "token for github.com with config token",
			host:       "github.com",
			config:     testHostsConfig(),
			wantToken:  "xxxxxxxxxxxxxxxxxxxx",
			wantSource: "oauth_token",
		},
		{
			name:                  "token for enterprise.com with GOCTL_ENTERPRISE_TOKEN, GITHUB_ENTERPRISE_TOKEN, and config token",
			host:                  "enterprise.com",
			goctlEnterpriseToken:     "GOCTL_ENTERPRISE_TOKEN",
			githubEnterpriseToken: "GITHUB_ENTERPRISE_TOKEN",
			config:                testHostsConfig(),
			wantToken:             "GOCTL_ENTERPRISE_TOKEN",
			wantSource:            "GOCTL_ENTERPRISE_TOKEN",
		},
		{
			name:                  "token for enterprise.com with GITHUB_ENTERPRISE_TOKEN, and config token",
			host:                  "enterprise.com",
			githubEnterpriseToken: "GITHUB_ENTERPRISE_TOKEN",
			config:                testHostsConfig(),
			wantToken:             "GITHUB_ENTERPRISE_TOKEN",
			wantSource:            "GITHUB_ENTERPRISE_TOKEN",
		},
		{
			name:       "token for enterprise.com with config token",
			host:       "enterprise.com",
			config:     testHostsConfig(),
			wantToken:  "yyyyyyyyyyyyyyyyyyyy",
			wantSource: "oauth_token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GITHUB_TOKEN", tt.githubToken)
			t.Setenv("GITHUB_ENTERPRISE_TOKEN", tt.githubEnterpriseToken)
			t.Setenv("GOCTL_TOKEN", tt.goctlToken)
			t.Setenv("GOCTL_ENTERPRISE_TOKEN", tt.goctlEnterpriseToken)
			token, source := tokenForHost(tt.config, tt.host)
			assert.Equal(t, tt.wantToken, token)
			assert.Equal(t, tt.wantSource, source)
		})
	}
}

func TestDefaultHost(t *testing.T) {
	tests := []struct {
		name         string
		config       *config.Config
		goctlHost       string
		wantHost     string
		wantSource   string
		wantNotFound bool
	}{
		{
			name:       "GOCTL_HOST if set",
			config:     testHostsConfig(),
			goctlHost:     "test.com",
			wantHost:   "test.com",
			wantSource: "GOCTL_HOST",
		},
		{
			name:       "authenticated host if only one",
			config:     testSingleHostConfig(),
			wantHost:   "enterprise.com",
			wantSource: "hosts",
		},
		{
			name:         "default host if more than one authenticated host",
			config:       testHostsConfig(),
			wantHost:     "github.com",
			wantSource:   "default",
			wantNotFound: true,
		},
		{
			name:         "default host if no authenticated host",
			config:       testNoHostsConfig(),
			wantHost:     "github.com",
			wantSource:   "default",
			wantNotFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.goctlHost != "" {
				t.Setenv("GOCTL_HOST", tt.goctlHost)
			}
			host, source := defaultHost(tt.config)
			assert.Equal(t, tt.wantHost, host)
			assert.Equal(t, tt.wantSource, source)
		})
	}
}

func TestKnownHosts(t *testing.T) {
	tests := []struct {
		name      string
		config    *config.Config
		goctlHost    string
		goctlToken   string
		wantHosts []string
	}{
		{
			name:      "no known hosts",
			config:    testNoHostsConfig(),
			wantHosts: []string{},
		},
		{
			name:      "includes GOCTL_HOST",
			config:    testNoHostsConfig(),
			goctlHost:    "test.com",
			wantHosts: []string{"test.com"},
		},
		{
			name:      "includes authenticated hosts",
			config:    testHostsConfig(),
			wantHosts: []string{"github.com", "enterprise.com"},
		},
		{
			name:      "includes default host if environment auth token",
			config:    testNoHostsConfig(),
			goctlToken:   "TOKEN",
			wantHosts: []string{"github.com"},
		},
		{
			name:      "deduplicates hosts",
			config:    testHostsConfig(),
			goctlHost:    "test.com",
			goctlToken:   "TOKEN",
			wantHosts: []string{"test.com", "github.com", "enterprise.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.goctlHost != "" {
				t.Setenv("GOCTL_HOST", tt.goctlHost)
			}
			if tt.goctlToken != "" {
				t.Setenv("GOCTL_TOKEN", tt.goctlToken)
			}
			hosts := knownHosts(tt.config)
			assert.Equal(t, tt.wantHosts, hosts)
		})
	}
}

func TestIsEnterprise(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		wantOut bool
	}{
		{
			name:    "github",
			host:    "github.com",
			wantOut: false,
		},
		{
			name:    "localhost",
			host:    "github.localhost",
			wantOut: false,
		},
		{
			name:    "enterprise",
			host:    "mygithub.com",
			wantOut: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := isEnterprise(tt.host)
			assert.Equal(t, tt.wantOut, out)
		})
	}
}

func TestNormalizeHostname(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		wantHost string
	}{
		{
			name:     "github domain",
			host:     "test.github.com",
			wantHost: "github.com",
		},
		{
			name:     "capitalized",
			host:     "GitHub.com",
			wantHost: "github.com",
		},
		{
			name:     "localhost domain",
			host:     "test.github.localhost",
			wantHost: "github.localhost",
		},
		{
			name:     "enterprise domain",
			host:     "mygithub.com",
			wantHost: "mygithub.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalized := normalizeHostname(tt.host)
			assert.Equal(t, tt.wantHost, normalized)
		})
	}
}

func testNoHostsConfig() *config.Config {
	var data = ``
	return config.ReadFromString(data)
}

func testSingleHostConfig() *config.Config {
	var data = `
hosts:
  enterprise.com:
    user: user2
    oauth_token: yyyyyyyyyyyyyyyyyyyy
    git_protocol: https
`
	return config.ReadFromString(data)
}

func testHostsConfig() *config.Config {
	var data = `
hosts:
  github.com:
    user: user1
    oauth_token: xxxxxxxxxxxxxxxxxxxx
    git_protocol: ssh
  enterprise.com:
    user: user2
    oauth_token: yyyyyyyyyyyyyyyyyyyy
    git_protocol: https
`
	return config.ReadFromString(data)
}
