package github

import (
	"context"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type Config struct {
	Token            string
	Owner            string
	BaseURL          string
	Insecure         bool
	WriteDelay       time.Duration
	ReadDelay        time.Duration
	RetryDelay       time.Duration
	RetryableErrors  map[int]bool
	MaxRetries       int
	ParallelRequests bool
}

type Owner struct {
	name           string
	id             int64
	v3client       *github.Client
	v4client       *githubv4.Client
	StopContext    context.Context
	IsOrganization bool
}

func RateLimitedHTTPClient(client *http.Client, writeDelay time.Duration, readDelay time.Duration, retryDelay time.Duration, parallelRequests bool, retryableErrors map[int]bool, maxRetries int) *http.Client {

	client.Transport = NewEtagTransport(client.Transport)
	client.Transport = NewRateLimitTransport(client.Transport, WithWriteDelay(writeDelay), WithReadDelay(readDelay), WithParallelRequests(parallelRequests))
	client.Transport = logging.NewSubsystemLoggingHTTPTransport("GitHub", client.Transport)
	client.Transport = newPreviewHeaderInjectorTransport(map[string]string{
		// TODO: remove when Stone Crop preview is moved to general availability in the GraphQL API
		"Accept": "application/vnd.github.stone-crop-preview+json",
	}, client.Transport)

	if maxRetries > 0 {
		client.Transport = NewRetryTransport(client.Transport, WithRetryDelay(retryDelay), WithRetryableErrors(retryableErrors), WithMaxRetries(maxRetries))
	}

	return client
}

func (c *Config) AuthenticatedHTTPClient() *http.Client {

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.Token},
	)
	client := oauth2.NewClient(ctx, ts)

	return RateLimitedHTTPClient(client, c.WriteDelay, c.ReadDelay, c.RetryDelay, c.ParallelRequests, c.RetryableErrors, c.MaxRetries)
}

func (c *Config) Anonymous() bool {
	return c.Token == ""
}

func (c *Config) AnonymousHTTPClient() *http.Client {
	client := &http.Client{Transport: &http.Transport{}}
	return RateLimitedHTTPClient(client, c.WriteDelay, c.ReadDelay, c.RetryDelay, c.ParallelRequests, c.RetryableErrors, c.MaxRetries)
}

func (c *Config) NewGraphQLClient(client *http.Client) (*githubv4.Client, error) {

	uv4, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	if uv4.String() != "https://api.github.com/" {
		uv4.Path = path.Join(uv4.Path, "api/graphql/")
	} else {
		uv4.Path = path.Join(uv4.Path, "graphql")
	}

	return githubv4.NewEnterpriseClient(uv4.String(), client), nil
}

func (c *Config) NewRESTClient(client *http.Client) (*github.Client, error) {

	uv3, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	if uv3.String() != "https://api.github.com/" {
		uv3.Path = uv3.Path + "api/v3/"
	}

	v3client, err := github.NewClient(client).WithEnterpriseURLs(uv3.String(), "")
	if err != nil {
		return nil, err
	}

	v3client.BaseURL = uv3

	return v3client, nil
}

func (c *Config) ConfigureOwner(owner *Owner) (*Owner, error) {
	ctx := context.Background()
	owner.name = c.Owner
	if owner.name == "" {
		if c.Anonymous() {
			return owner, nil
		}
		// Discover authenticated user
		user, _, err := owner.v3client.Users.Get(ctx, "")
		if err != nil {
			return nil, err
		}
		owner.name = user.GetLogin()
	} else {
		remoteOrg, _, err := owner.v3client.Organizations.Get(ctx, owner.name)
		if err == nil {
			if remoteOrg != nil {
				owner.id = remoteOrg.GetID()
				owner.IsOrganization = true
			}
		}
	}

	return owner, nil
}

// Meta returns the meta parameter that is passed into subsequent resources
// https://godoc.org/github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema#ConfigureFunc
func (c *Config) Meta() (interface{}, error) {

	var client *http.Client
	if c.Anonymous() {
		client = c.AnonymousHTTPClient()
	} else {
		client = c.AuthenticatedHTTPClient()
	}

	v3client, err := c.NewRESTClient(client)
	if err != nil {
		return nil, err
	}

	v4client, err := c.NewGraphQLClient(client)
	if err != nil {
		return nil, err
	}

	var owner Owner
	owner.v4client = v4client
	owner.v3client = v3client
	owner.StopContext = context.Background()

	_, err = c.ConfigureOwner(&owner)
	if err != nil {
		return &owner, err
	}
	return &owner, nil
}

type previewHeaderInjectorTransport struct {
	rt             http.RoundTripper
	previewHeaders map[string]string
}

func newPreviewHeaderInjectorTransport(headers map[string]string, rt http.RoundTripper) *previewHeaderInjectorTransport {
	return &previewHeaderInjectorTransport{
		rt:             rt,
		previewHeaders: headers,
	}
}

func (injector *previewHeaderInjectorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for name, value := range injector.previewHeaders {
		header := req.Header.Get(name)
		if header == "" {
			header = value
		} else {
			header = strings.Join([]string{header, value}, ",")
		}
		req.Header.Set(name, header)
	}
	return injector.rt.RoundTrip(req)
}
