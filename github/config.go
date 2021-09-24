package github

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"path"

	"github.com/google/go-github/v39/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/logging"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type Config struct {
	Token    string
	Owner    string
	BaseURL  string
	Insecure bool
}

type Owner struct {
	name           string
	id             int64
	v3client       *github.Client
	v4client       *githubv4.Client
	StopContext    context.Context
	IsOrganization bool
}

func RateLimitedHTTPClient(client *http.Client) *http.Client {

	client.Transport = NewEtagTransport(client.Transport)
	client.Transport = NewRateLimitTransport(client.Transport)
	client.Transport = logging.NewTransport("Github", client.Transport)

	return client

}

func (c *Config) AuthenticatedHTTPClient() *http.Client {

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.Token},
	)
	client := oauth2.NewClient(ctx, ts)

	return RateLimitedHTTPClient(client)

}

func (c *Config) Anonymous() bool {
	return c.Token == ""
}

func (c *Config) AnonymousHTTPClient() *http.Client {
	client := &http.Client{Transport: &http.Transport{}}
	return RateLimitedHTTPClient(client)
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

	v3client, err := github.NewEnterpriseClient(uv3.String(), "", client)
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
// https://godoc.org/github.com/hashicorp/terraform-plugin-sdk/helper/schema#ConfigureFunc
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

	if c.Anonymous() {
		log.Printf("[DEBUG] No token present; configuring anonymous owner.")
		return &owner, nil
	} else {
		_, err = c.ConfigureOwner(&owner)
		if err != nil {
			return &owner, err
		}
		log.Printf("[DEBUG] Token present; configuring authenticated owner: %s", owner.name)
		return &owner, nil
	}
}
