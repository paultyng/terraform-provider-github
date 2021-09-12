package github

import (
	"fmt"
	"log"
	"os"
	"testing"
)

var testCollaborator string = os.Getenv("GITHUB_TEST_COLLABORATOR")
var isEnterprise string = os.Getenv("ENTERPRISE_ACCOUNT")
var testOrganization string = testOrganizationFunc()
var testOwner string = os.Getenv("GITHUB_OWNER")
var testToken string = os.Getenv("GITHUB_TOKEN")
var testBaseURLGHES string = os.Getenv("GHES_BASE_URL")

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("GITHUB_TOKEN"); v == "" {
		t.Fatal("GITHUB_TOKEN must be set for acceptance tests")
	}
	if v := os.Getenv("GITHUB_ORGANIZATION"); v == "" && os.Getenv("GITHUB_OWNER") == "" {
		t.Fatal("GITHUB_ORGANIZATION or GITHUB_OWNER must be set for acceptance tests")
	}
	if v := os.Getenv("GITHUB_TEST_USER"); v == "" {
		t.Fatal("GITHUB_TEST_USER must be set for acceptance tests")
	}
	if v := os.Getenv("GITHUB_TEST_COLLABORATOR"); v == "" {
		t.Fatal("GITHUB_TEST_COLLABORATOR must be set for acceptance tests")
	}
	if v := os.Getenv("GITHUB_TEMPLATE_REPOSITORY"); v == "" {
		t.Fatal("GITHUB_TEMPLATE_REPOSITORY must be set for acceptance tests")
	}
	if v := os.Getenv("GITHUB_TEMPLATE_REPOSITORY_RELEASE_ID"); v == "" {
		t.Fatal("GITHUB_TEMPLATE_REPOSITORY_RELEASE_ID must be set for acceptance tests")
	}
}

func skipUnlessMode(t *testing.T, providerMode string) {
	switch providerMode {
	case anonymous:
		if os.Getenv("GITHUB_BASE_URL") != "" &&
			os.Getenv("GITHUB_BASE_URL") != "https://api.github.com/" {
			t.Log("anonymous mode not supported for GHES deployments")
			break
		}

		if os.Getenv("GITHUB_TOKEN") == "" {
			return
		} else {
			t.Log("GITHUB_TOKEN environment variable should be empty")
		}
	case individual:
		if os.Getenv("GITHUB_TOKEN") != "" && os.Getenv("GITHUB_OWNER") != "" {
			return
		} else {
			t.Log("GITHUB_TOKEN and GITHUB_OWNER environment variables should be set")
		}
	case organization:
		if os.Getenv("GITHUB_TOKEN") != "" && os.Getenv("GITHUB_ORGANIZATION") != "" {
			return
		} else {
			t.Log("GITHUB_TOKEN and GITHUB_ORGANIZATION environment variables should be set")
		}
	case enterprise:
		if os.Getenv("GITHUB_TOKEN") != "" && os.Getenv("ENTERPRISE_ACCOUNT") != "" {
			return
		} else {
			t.Log("GITHUB_TOKEN and ENTERPRISE_ACCOUNT environment variables should be set")
		}
	}

	t.Skipf("Skipping %s which requires %s mode", t.Name(), providerMode)
}

func testAccCheckOrganization() error {

	baseURL := os.Getenv("GITHUB_BASE_URL")
	token := os.Getenv("GITHUB_TOKEN")

	owner := os.Getenv("GITHUB_OWNER")
	if owner == "" {
		organization := os.Getenv("GITHUB_ORGANIZATION")
		if organization == "" {
			return fmt.Errorf("Neither `GITHUB_OWNER` or `GITHUB_ORGANIZATION` set in environment")
		}
		owner = organization
	}

	config := Config{
		BaseURL: baseURL,
		Token:   token,
		Owner:   owner,
	}

	meta, err := config.Meta()
	if err != nil {
		return err
	}
	if !meta.(*Owner).IsOrganization {
		return fmt.Errorf("GITHUB_OWNER %q is a user, not an organization", meta.(*Owner).name)
	}
	return nil
}

func OwnerOrOrgEnvDefaultFunc() (interface{}, error) {
	if organization := os.Getenv("GITHUB_ORGANIZATION"); organization != "" {
		log.Printf("[DEBUG] Selecting owner %s from GITHUB_ORGANIZATION environment variable", organization)
		return organization, nil
	}
	owner := os.Getenv("GITHUB_OWNER")
	log.Printf("[DEBUG] Selecting owner %s from GITHUB_OWNER environment variable", owner)
	return owner, nil
}

func testOrganizationFunc() string {
	organization := os.Getenv("GITHUB_ORGANIZATION")
	if organization == "" {
		organization = os.Getenv("GITHUB_TEST_ORGANIZATION")
	}
	return organization
}

func testOwnerFunc() string {
	owner := os.Getenv("GITHUB_OWNER")
	if owner == "" {
		owner = os.Getenv("GITHUB_TEST_OWNER")
	}
	return owner
}

const anonymous = "anonymous"
const individual = "individual"
const organization = "organization"
const enterprise = "enterprise"
