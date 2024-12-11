package github

import (
	"context"
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type testMode string

const (
	anonymous    testMode = "anonymous"
	individual   testMode = "individual"
	organization testMode = "organization"
	team         testMode = "team"
	enterprise   testMode = "enterprise"
)

type testAccConfig struct {
	// Target configuration
	baseUrl string

	// Auth configuration
	authMode testMode
	owner    string
	username string
	token    string

	// Enterprise configuration
	enterpriseSlug string

	// Global test configuration
	testPublicRepository              string
	testPublicRepositoryOwner         string
	testPublicReleaseId               int
	testPublicTemplateRepository      string
	testPublicTemplateRepositoryOwner string
	testGHActionsAppInstallationId    int

	// User test configuration
	testUserRepository string

	// Org test configuration
	testOrgUser               string
	testOrgRepository         string
	testOrgTemplateRepository string
	testOrgAppInstallationId  int

	// External test configuration
	testExternalUser      string
	testExternalUserToken string
	testExternalUser2     string

	// Test options
	testAdvancedSecurity bool
}

var testAccConf *testAccConfig

// providerFactories are used to instantiate a provider during acceptance testing.
// The factory function will be invoked for every Terraform CLI command executed
// to create a provider server to which the CLI can reattach.
var providerFactories = map[string]func() (*schema.Provider, error){
	"github": func() (*schema.Provider, error) {
		return Provider(), nil
	},
}

func TestMain(m *testing.M) {
	authMode := testMode(os.Getenv("GITHUB_TEST_AUTH_MODE"))
	if len(authMode) == 0 {
		authMode = anonymous
	}

	config := testAccConfig{
		baseUrl:                           os.Getenv("GITHUB_BASE_URL"),
		authMode:                          authMode,
		testPublicRepository:              "terraform-provider-github",
		testPublicRepositoryOwner:         "integrations",
		testPublicReleaseId:               186531906,
		testPublicTemplateRepository:      "template-repository",
		testPublicTemplateRepositoryOwner: "template-repository",
		testGHActionsAppInstallationId:    15368,
		testUserRepository:                os.Getenv("GITHUB_TEST_USER_REPOSITORY"),
		testOrgUser:                       os.Getenv("GITHUB_TEST_ORG_USER"),
		testOrgRepository:                 os.Getenv("GITHUB_TEST_ORG_REPOSITORY"),
		testOrgTemplateRepository:         os.Getenv("GITHUB_TEST_ORG_TEMPLATE_REPOSITORY"),
		testExternalUser:                  os.Getenv("GITHUB_TEST_EXTERNAL_USER"),
		testExternalUserToken:             os.Getenv("GITHUB_TEST_EXTERNAL_USER_TOKEN"),
		testExternalUser2:                 os.Getenv("GITHUB_TEST_EXTERNAL_USER2"),
		testAdvancedSecurity:              os.Getenv("GITHUB_TEST_ADVANCED_SECURITY") == "true",
	}

	if config.authMode != anonymous {
		config.owner = os.Getenv("GITHUB_OWNER")
		config.username = os.Getenv("GITHUB_USERNAME")
		config.token = os.Getenv("GITHUB_TOKEN")

		if len(config.owner) == 0 {
			fmt.Println("GITHUB_OWNER environment variable not set")
			os.Exit(1)
		}

		if len(config.username) == 0 {
			fmt.Println("GITHUB_USERNAME environment variable not set")
			os.Exit(1)
		}

		if len(config.token) == 0 {
			fmt.Println("GITHUB_TOKEN environment variable not set")
			os.Exit(1)
		}
	}

	if config.authMode == enterprise {
		config.enterpriseSlug = os.Getenv("GITHUB_ENTERPRISE_SLUG")

		if len(config.enterpriseSlug) == 0 {
			fmt.Println("GITHUB_ENTERPRISE_SLUG environment variable not set")
			os.Exit(1)
		}
	}

	i, err := strconv.Atoi(os.Getenv("GITHUB_TEST_ORG_APP_INSTALLATION_ID"))
	if err == nil {
		config.testOrgAppInstallationId = i
	}

	testAccConf = &config

	if testAccConf.authMode != anonymous {
		meta, err := getTestMeta()
		if err != nil {
			fmt.Println("could not get test meta for sweepers: %w", err)
			os.Exit(1)
		}

		resource.TestMain(m)

		resource.AddTestSweepers("github_repository", &resource.Sweeper{
			Name: "github_repository",
			F:    getSweepRepositoriesFunc(meta),
		})
	}

	m.Run()
}

func getTestMeta() (*Owner, error) {
	config := Config{
		Token:   testAccConf.token,
		Owner:   testAccConf.owner,
		BaseURL: testAccConf.baseUrl,
	}

	meta, err := config.Meta()
	if err != nil {
		return nil, fmt.Errorf("error getting GitHub meta parameter")
	}

	return meta.(*Owner), nil
}

func getSweepRepositoriesFunc(meta *Owner) func(string) error {
	return func(prefix string) error {
		client := meta.v3client
		owner := meta.name

		repos, _, err := client.Repositories.ListByUser(context.TODO(), owner, nil)
		if err != nil {
			return err
		}

		for _, r := range repos {
			if name := r.GetName(); strings.HasPrefix(name, prefix) {
				log.Printf("[DEBUG] Destroying Repository %s", name)

				if _, err := client.Repositories.Delete(context.TODO(), owner, name); err != nil {
					return err
				}
			}
		}

		return nil
	}
}

func skipUnauthenticated(t *testing.T) {
	if testAccConf.authMode == anonymous {
		t.Skip("Skipping as test mode not authenticated")
	}
}

func skipUnlessHasOrgs(t *testing.T) {
	orgModes := []testMode{organization, team, enterprise}

	if !slices.Contains(orgModes, testAccConf.authMode) {
		t.Skip("Skipping as test mode doesn't have orgs")
	}
}

func skipUnlessHasPaidOrgs(t *testing.T) {
	orgModes := []testMode{team, enterprise}

	if !slices.Contains(orgModes, testAccConf.authMode) {
		t.Skip("Skipping as test mode doesn't have orgs")
	}
}

func skipUnlessMode(t *testing.T, testModes ...testMode) {
	if !slices.Contains(testModes, testAccConf.authMode) {
		t.Skip("Skipping as not supported test mode")
	}
}
