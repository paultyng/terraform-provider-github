package github

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
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
			"retryable_errors": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeInt},
				Optional: true,
				DefaultFunc: func() (interface{}, error) {
					defaultErrors := []int{500, 502, 503, 504}
					errorInterfaces := make([]interface{}, len(defaultErrors))
					for i, v := range defaultErrors {
						errorInterfaces[i] = v
					}
					return errorInterfaces, nil
				},
				Description: descriptions["retryable_errors"],
			},
			"max_retries": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     3,
				Description: descriptions["max_retries"],
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
			"write_delay_ms": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1000,
				Description: descriptions["write_delay_ms"],
			},
			"read_delay_ms": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: descriptions["read_delay_ms"],
			},
			"retry_delay_ms": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1000,
				Description: descriptions["retry_delay_ms"],
			},
			"parallel_requests": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: descriptions["parallel_requests"],
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
			},
			"gh_user": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"GITHUB_USER", "GIT_USER", "GIT_USERNAME"}, nil),
				Description: descriptions["gh_user"],
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"github_enterprise_actions_permissions":                                 resourceGithubActionsEnterprisePermissions(),
			"github_actions_environment_secret":                                     resourceGithubActionsEnvironmentSecret(),
			"github_actions_environment_variable":                                   resourceGithubActionsEnvironmentVariable(),
			"github_actions_organization_oidc_subject_claim_customization_template": resourceGithubActionsOrganizationOIDCSubjectClaimCustomizationTemplate(),
			"github_actions_organization_permissions":                               resourceGithubActionsOrganizationPermissions(),
			"github_actions_organization_secret":                                    resourceGithubActionsOrganizationSecret(),
			"github_actions_organization_variable":                                  resourceGithubActionsOrganizationVariable(),
			"github_actions_organization_secret_repositories":                       resourceGithubActionsOrganizationSecretRepositories(),
			"github_actions_repository_access_level":                                resourceGithubActionsRepositoryAccessLevel(),
			"github_actions_repository_oidc_subject_claim_customization_template":   resourceGithubActionsRepositoryOIDCSubjectClaimCustomizationTemplate(),
			"github_actions_repository_permissions":                                 resourceGithubActionsRepositoryPermissions(),
			"github_actions_runner_group":                                           resourceGithubActionsRunnerGroup(),
			"github_actions_secret":                                                 resourceGithubActionsSecret(),
			"github_actions_variable":                                               resourceGithubActionsVariable(),
			"github_app_installation_repositories":                                  resourceGithubAppInstallationRepositories(),
			"github_app_installation_repository":                                    resourceGithubAppInstallationRepository(),
			"github_branch":                                                         resourceGithubBranch(),
			"github_branch_default":                                                 resourceGithubBranchDefault(),
			"github_branch_protection":                                              resourceGithubBranchProtection(),
			"github_branch_protection_v3":                                           resourceGithubBranchProtectionV3(),
			"github_codespaces_organization_secret":                                 resourceGithubCodespacesOrganizationSecret(),
			"github_codespaces_organization_secret_repositories":                    resourceGithubCodespacesOrganizationSecretRepositories(),
			"github_codespaces_secret":                                              resourceGithubCodespacesSecret(),
			"github_codespaces_user_secret":                                         resourceGithubCodespacesUserSecret(),
			"github_dependabot_organization_secret":                                 resourceGithubDependabotOrganizationSecret(),
			"github_dependabot_organization_secret_repositories":                    resourceGithubDependabotOrganizationSecretRepositories(),
			"github_dependabot_secret":                                              resourceGithubDependabotSecret(),
			"github_emu_group_mapping":                                              resourceGithubEMUGroupMapping(),
			"github_issue":                                                          resourceGithubIssue(),
			"github_issue_label":                                                    resourceGithubIssueLabel(),
			"github_issue_labels":                                                   resourceGithubIssueLabels(),
			"github_membership":                                                     resourceGithubMembership(),
			"github_organization_block":                                             resourceOrganizationBlock(),
			"github_organization_custom_role":                                       resourceGithubOrganizationCustomRole(),
			"github_organization_project":                                           resourceGithubOrganizationProject(),
			"github_organization_security_manager":                                  resourceGithubOrganizationSecurityManager(),
			"github_organization_ruleset":                                           resourceGithubOrganizationRuleset(),
			"github_organization_settings":                                          resourceGithubOrganizationSettings(),
			"github_organization_webhook":                                           resourceGithubOrganizationWebhook(),
			"github_project_card":                                                   resourceGithubProjectCard(),
			"github_project_column":                                                 resourceGithubProjectColumn(),
			"github_release":                                                        resourceGithubRelease(),
			"github_repository":                                                     resourceGithubRepository(),
			"github_repository_autolink_reference":                                  resourceGithubRepositoryAutolinkReference(),
			"github_repository_dependabot_security_updates":                         resourceGithubRepositoryDependabotSecurityUpdates(),
			"github_repository_collaborator":                                        resourceGithubRepositoryCollaborator(),
			"github_repository_collaborators":                                       resourceGithubRepositoryCollaborators(),
			"github_repository_deploy_key":                                          resourceGithubRepositoryDeployKey(),
			"github_repository_deployment_branch_policy":                            resourceGithubRepositoryDeploymentBranchPolicy(),
			"github_repository_environment":                                         resourceGithubRepositoryEnvironment(),
			"github_repository_environment_deployment_policy":                       resourceGithubRepositoryEnvironmentDeploymentPolicy(),
			"github_repository_file":                                                resourceGithubRepositoryFile(),
			"github_repository_milestone":                                           resourceGithubRepositoryMilestone(),
			"github_repository_project":                                             resourceGithubRepositoryProject(),
			"github_repository_pull_request":                                        resourceGithubRepositoryPullRequest(),
			"github_repository_ruleset":                                             resourceGithubRepositoryRuleset(),
			"github_repository_topics":                                              resourceGithubRepositoryTopics(),
			"github_repository_webhook":                                             resourceGithubRepositoryWebhook(),
			"github_team":                                                           resourceGithubTeam(),
			"github_team_members":                                                   resourceGithubTeamMembers(),
			"github_team_membership":                                                resourceGithubTeamMembership(),
			"github_team_repository":                                                resourceGithubTeamRepository(),
			"github_team_settings":                                                  resourceGithubTeamSettings(),
			"github_team_sync_group_mapping":                                        resourceGithubTeamSyncGroupMapping(),
			"github_user_gpg_key":                                                   resourceGithubUserGpgKey(),
			"github_user_invitation_accepter":                                       resourceGithubUserInvitationAccepter(),
			"github_user_ssh_key":                                                   resourceGithubUserSshKey(),
			"github_enterprise_organization":                                        resourceGithubEnterpriseOrganization(),
			"github_enterprise_actions_runner_group":                                resourceGithubActionsEnterpriseRunnerGroup(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"github_actions_environment_secrets":                                    dataSourceGithubActionsEnvironmentSecrets(),
			"github_actions_environment_variables":                                  dataSourceGithubActionsEnvironmentVariables(),
			"github_actions_organization_oidc_subject_claim_customization_template": dataSourceGithubActionsOrganizationOIDCSubjectClaimCustomizationTemplate(),
			"github_actions_organization_public_key":                                dataSourceGithubActionsOrganizationPublicKey(),
			"github_actions_organization_registration_token":                        dataSourceGithubActionsOrganizationRegistrationToken(),
			"github_actions_organization_secrets":                                   dataSourceGithubActionsOrganizationSecrets(),
			"github_actions_organization_variables":                                 dataSourceGithubActionsOrganizationVariables(),
			"github_actions_public_key":                                             dataSourceGithubActionsPublicKey(),
			"github_actions_registration_token":                                     dataSourceGithubActionsRegistrationToken(),
			"github_actions_repository_oidc_subject_claim_customization_template":   dataSourceGithubActionsRepositoryOIDCSubjectClaimCustomizationTemplate(),
			"github_actions_secrets":                                                dataSourceGithubActionsSecrets(),
			"github_actions_variables":                                              dataSourceGithubActionsVariables(),
			"github_app":                                                            dataSourceGithubApp(),
			"github_app_token":                                                      dataSourceGithubAppToken(),
			"github_branch":                                                         dataSourceGithubBranch(),
			"github_branch_protection_rules":                                        dataSourceGithubBranchProtectionRules(),
			"github_collaborators":                                                  dataSourceGithubCollaborators(),
			"github_codespaces_organization_public_key":                             dataSourceGithubCodespacesOrganizationPublicKey(),
			"github_codespaces_organization_secrets":                                dataSourceGithubCodespacesOrganizationSecrets(),
			"github_codespaces_public_key":                                          dataSourceGithubCodespacesPublicKey(),
			"github_codespaces_secrets":                                             dataSourceGithubCodespacesSecrets(),
			"github_codespaces_user_public_key":                                     dataSourceGithubCodespacesUserPublicKey(),
			"github_codespaces_user_secrets":                                        dataSourceGithubCodespacesUserSecrets(),
			"github_dependabot_organization_public_key":                             dataSourceGithubDependabotOrganizationPublicKey(),
			"github_dependabot_organization_secrets":                                dataSourceGithubDependabotOrganizationSecrets(),
			"github_dependabot_public_key":                                          dataSourceGithubDependabotPublicKey(),
			"github_dependabot_secrets":                                             dataSourceGithubDependabotSecrets(),
			"github_external_groups":                                                dataSourceGithubExternalGroups(),
			"github_ip_ranges":                                                      dataSourceGithubIpRanges(),
			"github_issue_labels":                                                   dataSourceGithubIssueLabels(),
			"github_membership":                                                     dataSourceGithubMembership(),
			"github_organization":                                                   dataSourceGithubOrganization(),
			"github_organization_custom_role":                                       dataSourceGithubOrganizationCustomRole(),
			"github_organization_external_identities":                               dataSourceGithubOrganizationExternalIdentities(),
			"github_organization_ip_allow_list":                                     dataSourceGithubOrganizationIpAllowList(),
			"github_organization_team_sync_groups":                                  dataSourceGithubOrganizationTeamSyncGroups(),
			"github_organization_teams":                                             dataSourceGithubOrganizationTeams(),
			"github_organization_webhooks":                                          dataSourceGithubOrganizationWebhooks(),
			"github_ref":                                                            dataSourceGithubRef(),
			"github_release":                                                        dataSourceGithubRelease(),
			"github_repositories":                                                   dataSourceGithubRepositories(),
			"github_repository":                                                     dataSourceGithubRepository(),
			"github_repository_autolink_references":                                 dataSourceGithubRepositoryAutolinkReferences(),
			"github_repository_branches":                                            dataSourceGithubRepositoryBranches(),
			"github_repository_environments":                                        dataSourceGithubRepositoryEnvironments(),
			"github_repository_deploy_keys":                                         dataSourceGithubRepositoryDeployKeys(),
			"github_repository_deployment_branch_policies":                          dataSourceGithubRepositoryDeploymentBranchPolicies(),
			"github_repository_file":                                                dataSourceGithubRepositoryFile(),
			"github_repository_milestone":                                           dataSourceGithubRepositoryMilestone(),
			"github_repository_pull_request":                                        dataSourceGithubRepositoryPullRequest(),
			"github_repository_pull_requests":                                       dataSourceGithubRepositoryPullRequests(),
			"github_repository_teams":                                               dataSourceGithubRepositoryTeams(),
			"github_repository_webhooks":                                            dataSourceGithubRepositoryWebhooks(),
			"github_rest_api":                                                       dataSourceGithubRestApi(),
			"github_ssh_keys":                                                       dataSourceGithubSshKeys(),
			"github_team":                                                           dataSourceGithubTeam(),
			"github_tree":                                                           dataSourceGithubTree(),
			"github_user":                                                           dataSourceGithubUser(),
			"github_user_external_identity":                                         dataSourceGithubUserExternalIdentity(),
			"github_users":                                                          dataSourceGithubUsers(),
			"github_enterprise":                                                     dataSourceGithubEnterprise(),
		},
	}

	p.ConfigureContextFunc = providerConfigure(p)

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
		"write_delay_ms": "Amount of time in milliseconds to sleep in between writes to GitHub API. " +
			"Defaults to 1000ms or 1s if not set.",
		"read_delay_ms": "Amount of time in milliseconds to sleep in between non-write requests to GitHub API. " +
			"Defaults to 0ms if not set.",
		"retry_delay_ms": "Amount of time in milliseconds to sleep in between requests to GitHub API after an error response. " +
			"Defaults to 1000ms or 1s if not set, the max_retries must be set to greater than zero.",
		"parallel_requests": "Allow the provider to make parallel API calls to GitHub. " +
			"You may want to set it to true when you have a private Github Enterprise without strict rate limits. " +
			"Although, it is not possible to enable this setting on github.com " +
			"because we enforce the respect of github.com's best practices to avoid hitting abuse rate limits" +
			"Defaults to false if not set",
		"retryable_errors": "Allow the provider to retry after receiving an error status code, the max_retries should be set for this to work" +
			"Defaults to [500, 502, 503, 504]",
		"max_retries": "Number of times to retry a request after receiving an error status code" +
			"Defaults to 3",
		"gh_user": "the username to retrieve token from GitHub CLI. Defaults to GITHUB_USER, GIT_USER, GIT_USERNAME environemnt variables or none.",
	}
}

func providerConfigure(p *schema.Provider) schema.ConfigureContextFunc {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
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
			log.Printf("[INFO] Selecting organization attribute as owner: %s", org)
			owner = org
		}

		if appAuth, ok := d.Get("app_auth").([]interface{}); ok && len(appAuth) > 0 && appAuth[0] != nil {
			appAuthAttr := appAuth[0].(map[string]interface{})

			var appID, appInstallationID, appPemFile string

			if v, ok := appAuthAttr["id"].(string); ok && v != "" {
				appID = v
			} else {
				return nil, wrapErrors([]error{fmt.Errorf("app_auth.id must be set and contain a non-empty value")})
			}

			if v, ok := appAuthAttr["installation_id"].(string); ok && v != "" {
				appInstallationID = v
			} else {
				return nil, wrapErrors([]error{fmt.Errorf("app_auth.installation_id must be set and contain a non-empty value")})
			}

			if v, ok := appAuthAttr["pem_file"].(string); ok && v != "" {
				// The Go encoding/pem package only decodes PEM formatted blocks
				// that contain new lines. Some platforms, like Terraform Cloud,
				// do not support new lines within Environment Variables.
				// Any occurrence of \n in the `pem_file` argument's value
				// (explicit value, or default value taken from
				// GITHUB_APP_PEM_FILE Environment Variable) is replaced with an
				// actual new line character before decoding.
				appPemFile = strings.Replace(v, `\n`, "\n", -1)
			} else {
				return nil, wrapErrors([]error{fmt.Errorf("app_auth.pem_file must be set and contain a non-empty value")})
			}

			appToken, err := GenerateOAuthTokenFromApp(baseURL, appID, appInstallationID, appPemFile)
			if err != nil {
				return nil, wrapErrors([]error{err})
			}

			token = appToken
		}

		isGithubDotCom, err := regexp.MatchString("^"+regexp.QuoteMeta("https://api.github.com"), baseURL)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		if token == "" {
			ghUser := d.Get("gh_user").(string)
			ghAuthToken, err := tokenFromGhCli(baseURL, isGithubDotCom, ghUser)
			if err != nil {
				return nil, diag.FromErr(fmt.Errorf("gh auth token: %w", err))
			}
			token = ghAuthToken
		}

		writeDelay := d.Get("write_delay_ms").(int)
		if writeDelay <= 0 {
			return nil, wrapErrors([]error{fmt.Errorf("write_delay_ms must be greater than 0ms")})
		}
		log.Printf("[INFO] Setting write_delay_ms to %d", writeDelay)

		readDelay := d.Get("read_delay_ms").(int)
		if readDelay < 0 {
			return nil, wrapErrors([]error{fmt.Errorf("read_delay_ms must be greater than or equal to 0ms")})
		}
		log.Printf("[DEBUG] Setting read_delay_ms to %d", readDelay)

		retryDelay := d.Get("read_delay_ms").(int)
		if retryDelay < 0 {
			return nil, diag.FromErr(fmt.Errorf("retry_delay_ms must be greater than or equal to 0ms"))
		}
		log.Printf("[DEBUG] Setting retry_delay_ms to %d", retryDelay)

		maxRetries := d.Get("max_retries").(int)
		if maxRetries < 0 {
			return nil, diag.FromErr(fmt.Errorf("max_retries must be greater than or equal to 0"))
		}
		log.Printf("[DEBUG] Setting max_retries to %d", maxRetries)
		retryableErrors := make(map[int]bool)
		if maxRetries > 0 {
			reParam := d.Get("retryable_errors").([]interface{})
			if len(reParam) == 0 {
				retryableErrors = getDefaultRetriableErrors()
			} else {
				for _, status := range reParam {
					retryableErrors[status.(int)] = true
				}
			}

			log.Printf("[DEBUG] Setting retriableErrors to %v", retryableErrors)
		}

		parallelRequests := d.Get("parallel_requests").(bool)

		if parallelRequests && isGithubDotCom {
			return nil, wrapErrors([]error{fmt.Errorf("parallel_requests cannot be true when connecting to public github")})
		}
		log.Printf("[DEBUG] Setting parallel_requests to %t", parallelRequests)

		config := Config{
			Token:            token,
			BaseURL:          baseURL,
			Insecure:         insecure,
			Owner:            owner,
			WriteDelay:       time.Duration(writeDelay) * time.Millisecond,
			ReadDelay:        time.Duration(readDelay) * time.Millisecond,
			RetryDelay:       time.Duration(retryDelay) * time.Millisecond,
			RetryableErrors:  retryableErrors,
			MaxRetries:       maxRetries,
			ParallelRequests: parallelRequests,
		}

		meta, err := config.Meta()
		if err != nil {
			return nil, wrapErrors([]error{err})
		}

		return meta, nil
	}
}

// See https://github.com/integrations/terraform-provider-github/issues/1822
func tokenFromGhCli(baseURL string, isGithubDotCom bool, ghUser string) (string, error) {
	ghCliPath := os.Getenv("GH_PATH")
	if ghCliPath == "" {
		ghCliPath = "gh"
	}
	hostname := ""
	if isGithubDotCom {
		hostname = "github.com"
	} else {
		parsedURL, err := url.Parse(baseURL)
		if err != nil {
			return "", fmt.Errorf("parse %s: %w", baseURL, err)
		}
		hostname = parsedURL.Host
	}
	// GitHub CLI uses different base URLs in ~/.config/gh/hosts.yml, so when
	// we're using the standard base path of this provider, it doesn't align
	// with the way `gh` CLI stores the credentials. The following doesn't work:
	//
	// $ gh auth token --hostname api.github.com
	// > no oauth token
	//
	// ... but the following does work correctly
	//
	// $ gh auth token --hostname github.com
	// > gh..<valid token>
	hostname = strings.TrimPrefix(hostname, "api.")
	args := []string{"auth", "token", "--hostname", hostname}
	if ghUser != "" {
		// If the user is set, pass it to the `gh` CLI to retrieve the token. This is
		// useful when the user has multiple accounts on the development machine and
		// wants to use a specific one. This is especially useful for GitHub Enterprise
		// Managed User (EMU) accounts.
		args = append(args, "--user", ghUser)
	}
	out, err := exec.Command(ghCliPath, args...).Output()
	if err != nil {
		// GH CLI is either not installed or there was no `gh auth login` command issued,
		// which is fine. don't return the error to keep the flow going
		return "", nil
	}

	log.Printf("[INFO] Using the token from GitHub CLI")
	return strings.TrimSpace(string(out)), nil
}
