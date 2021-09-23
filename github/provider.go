package github

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func Provider() terraform.ResourceProvider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("GITHUB_TOKEN", nil),
				Description: descriptions["token"],
			},
			"owner": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("GITHUB_OWNER", nil),
				Description: descriptions["owner"],
			},
			"organization": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("GITHUB_ORGANIZATION", nil),
				Description: descriptions["organization"],
				Deprecated:  "Use owner (or GITHUB_OWNER) instead of organization (or GITHUB_ORGANIZATION)",
			},
			"base_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("GITHUB_BASE_URL", "https://api.github.com/"),
				Description: descriptions["base_url"],
			},
			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: descriptions["insecure"],
			},
			"app_auth": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: descriptions["app_auth"],
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							DefaultFunc: schema.EnvDefaultFunc("GITHUB_APP_ID", nil),
							Description: descriptions["app_auth.id"],
						},
						"installation_id": {
							Type:        schema.TypeString,
							Required:    true,
							DefaultFunc: schema.EnvDefaultFunc("GITHUB_APP_INSTALLATION_ID", nil),
							Description: descriptions["app_auth.installation_id"],
						},
						"pem_file": {
							Type:        schema.TypeString,
							Required:    true,
							Sensitive:   true,
							DefaultFunc: schema.EnvDefaultFunc("GITHUB_APP_PEM_FILE", nil),
							Description: descriptions["app_auth.pem_file"],
						},
					},
				},
				ConflictsWith: []string{"token"},
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"github_actions_environment_secret":               resourceGithubActionsEnvironmentSecret(),
			"github_actions_organization_secret":              resourceGithubActionsOrganizationSecret(),
			"github_actions_organization_secret_repositories": resourceGithubActionsOrganizationSecretRepositories(),
			"github_actions_runner_group":                     resourceGithubActionsRunnerGroup(),
			"github_actions_secret":                           resourceGithubActionsSecret(),
			"github_app_installation_repository":              resourceGithubAppInstallationRepository(),
			"github_branch":                                   resourceGithubBranch(),
			"github_branch_protection":                        resourceGithubBranchProtection(),
			"github_branch_protection_v3":                     resourceGithubBranchProtectionV3(),
			"github_issue_label":                              resourceGithubIssueLabel(),
			"github_membership":                               resourceGithubMembership(),
			"github_organization_block":                       resourceOrganizationBlock(),
			"github_organization_project":                     resourceGithubOrganizationProject(),
			"github_organization_webhook":                     resourceGithubOrganizationWebhook(),
			"github_project_card":                             resourceGithubProjectCard(),
			"github_project_column":                           resourceGithubProjectColumn(),
			"github_repository_collaborator":                  resourceGithubRepositoryCollaborator(),
			"github_repository_deploy_key":                    resourceGithubRepositoryDeployKey(),
			"github_repository_environment":                   resourceGithubRepositoryEnvironment(),
			"github_repository_file":                          resourceGithubRepositoryFile(),
			"github_repository_milestone":                     resourceGithubRepositoryMilestone(),
			"github_repository_project":                       resourceGithubRepositoryProject(),
			"github_repository_pull_request":                  resourceGithubRepositoryPullRequest(),
			"github_repository_webhook":                       resourceGithubRepositoryWebhook(),
			"github_repository":                               resourceGithubRepository(),
			"github_team_membership":                          resourceGithubTeamMembership(),
			"github_team_repository":                          resourceGithubTeamRepository(),
			"github_team_sync_group_mapping":                  resourceGithubTeamSyncGroupMapping(),
			"github_team":                                     resourceGithubTeam(),
			"github_user_gpg_key":                             resourceGithubUserGpgKey(),
			"github_user_invitation_accepter":                 resourceGithubUserInvitationAccepter(),
			"github_user_ssh_key":                             resourceGithubUserSshKey(),
			"github_branch_default":                           resourceGithubBranchDefault(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"github_actions_public_key":            dataSourceGithubActionsPublicKey(),
			"github_branch":                        dataSourceGithubBranch(),
			"github_collaborators":                 dataSourceGithubCollaborators(),
			"github_ip_ranges":                     dataSourceGithubIpRanges(),
			"github_membership":                    dataSourceGithubMembership(),
			"github_organization":                  dataSourceGithubOrganization(),
			"github_organization_team_sync_groups": dataSourceGithubOrganizationTeamSyncGroups(),
			"github_organization_teams":            dataSourceGithubOrganizationTeams(),
			"github_release":                       dataSourceGithubRelease(),
			"github_repositories":                  dataSourceGithubRepositories(),
			"github_repository":                    dataSourceGithubRepository(),
			"github_repository_file":               dataSourceGithubRepositoryFile(),
			"github_repository_milestone":          dataSourceGithubRepositoryMilestone(),
			"github_repository_pull_request":       dataSourceGithubRepositoryPullRequest(),
			"github_repository_pull_requests":      dataSourceGithubRepositoryPullRequests(),
			"github_team":                          dataSourceGithubTeam(),
			"github_user":                          dataSourceGithubUser(),
		},
	}

	p.ConfigureFunc = providerConfigure(p)

	return p
}

var descriptions map[string]string

func init() {
	descriptions = map[string]string{
		"token": "The OAuth token used to connect to GitHub. Anonymous mode is enabled if both `token` and " +
			"`app_auth` are not set.",

		"base_url": "The GitHub Base API URL",

		"insecure": "Enable `insecure` mode for testing purposes",

		"owner": "The GitHub owner name to manage. " +
			"Use this field instead of `organization` when managing individual accounts.",

		"organization": "The GitHub organization name to manage. " +
			"Use this field instead of `owner` when managing organization accounts.",

		"app_auth": "The GitHub App credentials used to connect to GitHub. Conflicts with " +
			"`token`. Anonymous mode is enabled if both `token` and `app_auth` are not set.",
		"app_auth.id":              "The GitHub App ID.",
		"app_auth.installation_id": "The GitHub App installation instance ID.",
		"app_auth.pem_file":        "The GitHub App PEM file contents.",
	}
}

func providerConfigure(p *schema.Provider) schema.ConfigureFunc {
	return func(d *schema.ResourceData) (interface{}, error) {
		owner := d.Get("owner").(string)
		baseURL := d.Get("base_url").(string)
		token := d.Get("token").(string)
		insecure := d.Get("insecure").(bool)

		// BEGIN backwards compatibility
		// OwnerOrOrgEnvDefaultFunc used to be the default value for both
		// 'owner' and 'organization'. This meant that if 'owner' and
		// 'GITHUB_OWNER' were set, 'GITHUB_OWNER' would be used as the default
		// value of 'organization' and therefore override 'owner'.
		//
		// This seems undesirable (an environment variable should not override
		// an explicitly set value in a provider block), but is necessary
		// for backwards compatibility. We could remove this backwards compatibility
		// code in a future major release.
		env, _ := OwnerOrOrgEnvDefaultFunc()
		if env.(string) != "" {
			owner = env.(string)
		}
		// END backwards compatibility

		org := d.Get("organization").(string)
		if org != "" {
			log.Printf("[DEBUG] Selecting organization attribute as owner: %s", org)
			owner = org
		}

		if appAuth, ok := d.Get("app_auth").([]interface{}); ok && len(appAuth) > 0 && appAuth[0] != nil {
			appAuthAttr := appAuth[0].(map[string]interface{})

			var appID, appInstallationID, appPemFile string

			if v, ok := appAuthAttr["id"].(string); ok && v != "" {
				appID = v
			} else {
				return nil, fmt.Errorf("app_auth.id must be set and contain a non-empty value")
			}

			if v, ok := appAuthAttr["installation_id"].(string); ok && v != "" {
				appInstallationID = v
			} else {
				return nil, fmt.Errorf("app_auth.installation_id must be set and contain a non-empty value")
			}

			if v, ok := appAuthAttr["pem_file"].(string); ok && v != "" {
				appPemFile = v
			} else {
				return nil, fmt.Errorf("app_auth.pem_file must be set and contain a non-empty value")
			}

			appToken, err := GenerateOAuthTokenFromApp(baseURL, appID, appInstallationID, appPemFile)
			if err != nil {
				return nil, err
			}

			token = appToken
		}

		config := Config{
			Token:    token,
			BaseURL:  baseURL,
			Insecure: insecure,
			Owner:    owner,
		}

		meta, err := config.Meta()
		if err != nil {
			return nil, err
		}

		meta.(*Owner).StopContext = p.StopContext()

		return meta, nil
	}
}
