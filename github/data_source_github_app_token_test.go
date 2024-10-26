package github

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestAccGithubAppTokenDataSource(t *testing.T) {
	expectedAccessToken := "W+2e/zjiMTweDAr2b35toCF+h29l7NW92rJIPvFrCJQK"

	owner := "test-owner"

	pemData, err := os.ReadFile(testGitHubAppPrivateKeyFile)
	assert.Nil(t, err)

	t.Run("creates a application token without error", func(t *testing.T) {
		ts := githubApiMock([]*mockResponse{
			{
				ExpectedUri: fmt.Sprintf("/api/v3/app/installations/%s/access_tokens", testGitHubAppInstallationID),
				ExpectedHeaders: map[string]string{
					"Accept": "application/vnd.github.v3+json",
				},
				ResponseBody: fmt.Sprintf(`{"token": "%s"}`, expectedAccessToken),
				StatusCode:   201,
			},
		})
		defer ts.Close()

		httpCl := http.DefaultClient
		httpCl.Transport = http.DefaultTransport

		client := github.NewClient(httpCl)
		u, _ := url.Parse(ts.URL + "/")
		client.BaseURL = u

		meta := &Owner{
			name:     owner,
			v3client: client,
		}

		testSchema := map[string]*schema.Schema{
			"app_id":          {Type: schema.TypeString},
			"installation_id": {Type: schema.TypeString},
			"pem_file":        {Type: schema.TypeString},
			"token":           {Type: schema.TypeString},
		}

		schema := schema.TestResourceDataRaw(t, testSchema, map[string]interface{}{
			"app_id":          testGitHubAppID,
			"installation_id": testGitHubAppInstallationID,
			"pem_file":        string(pemData),
			"token":           "",
		})

		err := dataSourceGithubAppTokenRead(schema, meta)
		assert.Nil(t, err)
		assert.Equal(t, expectedAccessToken, schema.Get("token"))
	})
}
