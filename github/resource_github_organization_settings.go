package github

import (
	"context"
	"log"
	"strconv"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceGithubOrganizationSettings() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubOrganizationSettingsCreateOrUpdate,
		Read:   resourceGithubOrganizationSettingsRead,
		Update: resourceGithubOrganizationSettingsCreateOrUpdate,
		Delete: resourceGithubOrganizationSettingsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"billing_email": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The billing email address for the organization.",
			},
			"company": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The company name for the organization.",
			},
			"email": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The email address for the organization.",
			},
			"twitter_username": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Twitter username for the organization.",
			},
			"location": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The location for the organization.",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name for the organization.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description for the organization.",
			},
			"has_organization_projects": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether or not organization projects are enabled for the organization.",
			},
			"has_repository_projects": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether or not repository projects are enabled for the organization.",
			},
			"default_repository_permission": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "read",
				Description:      "The default permission for organization members to create new repositories. Can be one of 'read', 'write', 'admin' or 'none'.",
				ValidateDiagFunc: toDiagFunc(validation.StringInSlice([]string{"read", "write", "admin", "none"}, false), "default_repository_permission"),
			},
			"members_can_create_repositories": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether or not organization members can create new repositories.",
			},
			"members_can_create_internal_repositories": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether or not organization members can create new internal repositories. For Enterprise Organizations only.",
			},
			"members_can_create_private_repositories": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether or not organization members can create new private repositories.",
			},
			"members_can_create_public_repositories": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether or not organization members can create new public repositories.",
			},
			"members_can_create_pages": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether or not organization members can create new pages.",
			},
			"members_can_create_public_pages": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether or not organization members can create new public pages.",
			},
			"members_can_create_private_pages": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether or not organization members can create new private pages.",
			},
			"members_can_fork_private_repositories": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether or not organization members can fork private repositories.",
			},
			"web_commit_signoff_required": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether or not commit signatures are required for commits to the organization.",
			},
			"blog": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The blog URL for the organization.",
			},
			"advanced_security_enabled_for_new_repositories": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: " Whether or not advanced security is enabled for new repositories.",
			},
			"dependabot_alerts_enabled_for_new_repositories": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether or not dependabot alerts are enabled for new repositories.",
			},
			"dependabot_security_updates_enabled_for_new_repositories": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: " Whether or not dependabot security updates are enabled for new repositories.",
			},
			"dependency_graph_enabled_for_new_repositories": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether or not dependency graph is enabled for new repositories.",
			},
			"secret_scanning_enabled_for_new_repositories": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether or not secret scanning is enabled for new repositories.",
			},
			"secret_scanning_push_protection_enabled_for_new_repositories": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether or not secret scanning push protection is enabled for new repositories.",
			},
		},
	}
}

func resourceGithubOrganizationSettingsCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}
	client := meta.(*Owner).v3client
	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	org := meta.(*Owner).name

	settings := github.Organization{
		BillingEmail:                       github.String(d.Get("billing_email").(string)),
		Company:                            github.String(d.Get("company").(string)),
		Email:                              github.String(d.Get("email").(string)),
		TwitterUsername:                    github.String(d.Get("twitter_username").(string)),
		Location:                           github.String(d.Get("location").(string)),
		Name:                               github.String(d.Get("name").(string)),
		Description:                        github.String(d.Get("description").(string)),
		HasOrganizationProjects:            github.Bool(d.Get("has_organization_projects").(bool)),
		HasRepositoryProjects:              github.Bool(d.Get("has_repository_projects").(bool)),
		DefaultRepoPermission:              github.String(d.Get("default_repository_permission").(string)),
		MembersCanCreateRepos:              github.Bool(d.Get("members_can_create_repositories").(bool)),
		MembersCanCreatePrivateRepos:       github.Bool(d.Get("members_can_create_private_repositories").(bool)),
		MembersCanCreatePublicRepos:        github.Bool(d.Get("members_can_create_public_repositories").(bool)),
		MembersCanCreatePages:              github.Bool(d.Get("members_can_create_pages").(bool)),
		MembersCanCreatePublicPages:        github.Bool(d.Get("members_can_create_public_pages").(bool)),
		MembersCanCreatePrivatePages:       github.Bool(d.Get("members_can_create_private_pages").(bool)),
		MembersCanForkPrivateRepos:         github.Bool(d.Get("members_can_fork_private_repositories").(bool)),
		WebCommitSignoffRequired:           github.Bool(d.Get("web_commit_signoff_required").(bool)),
		Blog:                               github.String(d.Get("blog").(string)),
		AdvancedSecurityEnabledForNewRepos: github.Bool(d.Get("advanced_security_enabled_for_new_repositories").(bool)),
		DependabotAlertsEnabledForNewRepos: github.Bool(d.Get("dependabot_alerts_enabled_for_new_repositories").(bool)),
		DependabotSecurityUpdatesEnabledForNewRepos:    github.Bool(d.Get("dependabot_security_updates_enabled_for_new_repositories").(bool)),
		DependencyGraphEnabledForNewRepos:              github.Bool(d.Get("dependency_graph_enabled_for_new_repositories").(bool)),
		SecretScanningEnabledForNewRepos:               github.Bool(d.Get("secret_scanning_enabled_for_new_repositories").(bool)),
		SecretScanningPushProtectionEnabledForNewRepos: github.Bool(d.Get("secret_scanning_push_protection_enabled_for_new_repositories").(bool)),
	}

	enterpriseSettings := github.Organization{
		BillingEmail:                       github.String(d.Get("billing_email").(string)),
		Company:                            github.String(d.Get("company").(string)),
		Email:                              github.String(d.Get("email").(string)),
		TwitterUsername:                    github.String(d.Get("twitter_username").(string)),
		Location:                           github.String(d.Get("location").(string)),
		Name:                               github.String(d.Get("name").(string)),
		Description:                        github.String(d.Get("description").(string)),
		HasOrganizationProjects:            github.Bool(d.Get("has_organization_projects").(bool)),
		HasRepositoryProjects:              github.Bool(d.Get("has_repository_projects").(bool)),
		DefaultRepoPermission:              github.String(d.Get("default_repository_permission").(string)),
		MembersCanCreateRepos:              github.Bool(d.Get("members_can_create_repositories").(bool)),
		MembersCanCreateInternalRepos:      github.Bool(d.Get("members_can_create_internal_repositories").(bool)),
		MembersCanCreatePrivateRepos:       github.Bool(d.Get("members_can_create_private_repositories").(bool)),
		MembersCanCreatePublicRepos:        github.Bool(d.Get("members_can_create_public_repositories").(bool)),
		MembersCanCreatePages:              github.Bool(d.Get("members_can_create_pages").(bool)),
		MembersCanCreatePublicPages:        github.Bool(d.Get("members_can_create_public_pages").(bool)),
		MembersCanCreatePrivatePages:       github.Bool(d.Get("members_can_create_private_pages").(bool)),
		MembersCanForkPrivateRepos:         github.Bool(d.Get("members_can_fork_private_repositories").(bool)),
		WebCommitSignoffRequired:           github.Bool(d.Get("web_commit_signoff_required").(bool)),
		Blog:                               github.String(d.Get("blog").(string)),
		AdvancedSecurityEnabledForNewRepos: github.Bool(d.Get("advanced_security_enabled_for_new_repositories").(bool)),
		DependabotAlertsEnabledForNewRepos: github.Bool(d.Get("dependabot_alerts_enabled_for_new_repositories").(bool)),
		DependabotSecurityUpdatesEnabledForNewRepos:    github.Bool(d.Get("dependabot_security_updates_enabled_for_new_repositories").(bool)),
		DependencyGraphEnabledForNewRepos:              github.Bool(d.Get("dependency_graph_enabled_for_new_repositories").(bool)),
		SecretScanningEnabledForNewRepos:               github.Bool(d.Get("secret_scanning_enabled_for_new_repositories").(bool)),
		SecretScanningPushProtectionEnabledForNewRepos: github.Bool(d.Get("secret_scanning_push_protection_enabled_for_new_repositories").(bool)),
	}

	enterpriseSettingsNoFork := github.Organization{
		BillingEmail:                       github.String(d.Get("billing_email").(string)),
		Company:                            github.String(d.Get("company").(string)),
		Email:                              github.String(d.Get("email").(string)),
		TwitterUsername:                    github.String(d.Get("twitter_username").(string)),
		Location:                           github.String(d.Get("location").(string)),
		Name:                               github.String(d.Get("name").(string)),
		Description:                        github.String(d.Get("description").(string)),
		HasOrganizationProjects:            github.Bool(d.Get("has_organization_projects").(bool)),
		HasRepositoryProjects:              github.Bool(d.Get("has_repository_projects").(bool)),
		DefaultRepoPermission:              github.String(d.Get("default_repository_permission").(string)),
		MembersCanCreateRepos:              github.Bool(d.Get("members_can_create_repositories").(bool)),
		MembersCanCreateInternalRepos:      github.Bool(d.Get("members_can_create_internal_repositories").(bool)),
		MembersCanCreatePrivateRepos:       github.Bool(d.Get("members_can_create_private_repositories").(bool)),
		MembersCanCreatePublicRepos:        github.Bool(d.Get("members_can_create_public_repositories").(bool)),
		MembersCanCreatePages:              github.Bool(d.Get("members_can_create_pages").(bool)),
		MembersCanCreatePublicPages:        github.Bool(d.Get("members_can_create_public_pages").(bool)),
		MembersCanCreatePrivatePages:       github.Bool(d.Get("members_can_create_private_pages").(bool)),
		WebCommitSignoffRequired:           github.Bool(d.Get("web_commit_signoff_required").(bool)),
		Blog:                               github.String(d.Get("blog").(string)),
		AdvancedSecurityEnabledForNewRepos: github.Bool(d.Get("advanced_security_enabled_for_new_repositories").(bool)),
		DependabotAlertsEnabledForNewRepos: github.Bool(d.Get("dependabot_alerts_enabled_for_new_repositories").(bool)),
		DependabotSecurityUpdatesEnabledForNewRepos:    github.Bool(d.Get("dependabot_security_updates_enabled_for_new_repositories").(bool)),
		DependencyGraphEnabledForNewRepos:              github.Bool(d.Get("dependency_graph_enabled_for_new_repositories").(bool)),
		SecretScanningEnabledForNewRepos:               github.Bool(d.Get("secret_scanning_enabled_for_new_repositories").(bool)),
		SecretScanningPushProtectionEnabledForNewRepos: github.Bool(d.Get("secret_scanning_push_protection_enabled_for_new_repositories").(bool)),
	}

	orgPlan, _, err := client.Organizations.Edit(ctx, org, nil)
	if err != nil {
		return err
	}

	if orgPlan.GetPlan().GetName() == "enterprise" {
		if _, ok := d.GetOk("members_can_fork_private_repositories"); !ok {
			orgSettings, _, err := client.Organizations.Edit(ctx, org, &enterpriseSettingsNoFork)
			if err != nil {
				return err
			}
			id := strconv.FormatInt(orgSettings.GetID(), 10)
			d.SetId(id)
		} else if _, ok := d.GetOk("members_can_fork_private_repositories"); ok {
			orgSettings, _, err := client.Organizations.Edit(ctx, org, &enterpriseSettings)
			if err != nil {
				return err
			}
			id := strconv.FormatInt(orgSettings.GetID(), 10)
			d.SetId(id)
		}
	} else {
		orgSettings, _, err := client.Organizations.Edit(ctx, org, &settings)
		if err != nil {
			return err
		}
		id := strconv.FormatInt(orgSettings.GetID(), 10)
		d.SetId(id)
	}

	return resourceGithubOrganizationSettingsRead(d, meta)
}

func resourceGithubOrganizationSettingsRead(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}
	client := meta.(*Owner).v3client
	ctx := context.Background()
	org := meta.(*Owner).name

	orgSettings, _, err := client.Organizations.Get(ctx, org)
	if err != nil {
		return err
	}

	if err = d.Set("billing_email", orgSettings.GetBillingEmail()); err != nil {
		return err
	}
	if err = d.Set("company", orgSettings.GetCompany()); err != nil {
		return err
	}
	if err = d.Set("email", orgSettings.GetEmail()); err != nil {
		return err
	}
	if err = d.Set("twitter_username", orgSettings.GetTwitterUsername()); err != nil {
		return err
	}
	if err = d.Set("location", orgSettings.GetLocation()); err != nil {
		return err
	}
	if err = d.Set("name", orgSettings.GetName()); err != nil {
		return err
	}
	if err = d.Set("description", orgSettings.GetDescription()); err != nil {
		return err
	}
	if err = d.Set("has_organization_projects", orgSettings.GetHasOrganizationProjects()); err != nil {
		return err
	}
	if err = d.Set("has_repository_projects", orgSettings.GetHasRepositoryProjects()); err != nil {
		return err
	}
	if err = d.Set("default_repository_permission", orgSettings.GetDefaultRepoPermission()); err != nil {
		return err
	}
	if err = d.Set("members_can_create_repositories", orgSettings.GetMembersCanCreateRepos()); err != nil {
		return err
	}
	if err = d.Set("members_can_create_internal_repositories", orgSettings.GetMembersCanCreateInternalRepos()); err != nil {
		return err
	}
	if err = d.Set("members_can_create_private_repositories", orgSettings.GetMembersCanCreatePrivateRepos()); err != nil {
		return err
	}
	if err = d.Set("members_can_create_public_repositories", orgSettings.GetMembersCanCreatePublicRepos()); err != nil {
		return err
	}
	if err = d.Set("members_can_create_pages", orgSettings.GetMembersCanCreatePages()); err != nil {
		return err
	}
	if err = d.Set("members_can_create_public_pages", orgSettings.GetMembersCanCreatePublicPages()); err != nil {
		return err
	}
	if err = d.Set("members_can_create_private_pages", orgSettings.GetMembersCanCreatePrivatePages()); err != nil {
		return err
	}
	if err = d.Set("members_can_fork_private_repositories", orgSettings.GetMembersCanForkPrivateRepos()); err != nil {
		return err
	}
	if err = d.Set("web_commit_signoff_required", orgSettings.GetWebCommitSignoffRequired()); err != nil {
		return err
	}
	if err = d.Set("blog", orgSettings.GetBlog()); err != nil {
		return err
	}
	if err = d.Set("advanced_security_enabled_for_new_repositories", orgSettings.GetAdvancedSecurityEnabledForNewRepos()); err != nil {
		return err
	}
	if err = d.Set("dependabot_alerts_enabled_for_new_repositories", orgSettings.GetDependabotAlertsEnabledForNewRepos()); err != nil {
		return err
	}
	if err = d.Set("dependabot_security_updates_enabled_for_new_repositories", orgSettings.GetDependabotSecurityUpdatesEnabledForNewRepos()); err != nil {
		return err
	}
	if err = d.Set("dependency_graph_enabled_for_new_repositories", orgSettings.GetDependencyGraphEnabledForNewRepos()); err != nil {
		return err
	}
	if err = d.Set("secret_scanning_enabled_for_new_repositories", orgSettings.GetSecretScanningEnabledForNewRepos()); err != nil {
		return err
	}
	if err = d.Set("secret_scanning_push_protection_enabled_for_new_repositories", orgSettings.GetSecretScanningPushProtectionEnabledForNewRepos()); err != nil {
		return err
	}
	return nil
}

func resourceGithubOrganizationSettingsDelete(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client
	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	org := meta.(*Owner).name

	// This will set org settings to default values
	settings := github.Organization{
		BillingEmail:                       github.String("email@example.com"),
		Company:                            github.String(""),
		Email:                              github.String(""),
		TwitterUsername:                    github.String(""),
		Location:                           github.String(""),
		Name:                               github.String(""),
		Description:                        github.String(""),
		HasOrganizationProjects:            github.Bool(true),
		HasRepositoryProjects:              github.Bool(true),
		DefaultRepoPermission:              github.String("read"),
		MembersCanCreateRepos:              github.Bool(true),
		MembersCanCreatePrivateRepos:       github.Bool(true),
		MembersCanCreatePublicRepos:        github.Bool(true),
		MembersCanCreatePages:              github.Bool(false),
		MembersCanCreatePublicPages:        github.Bool(true),
		MembersCanCreatePrivatePages:       github.Bool(true),
		MembersCanForkPrivateRepos:         github.Bool(false),
		WebCommitSignoffRequired:           github.Bool(false),
		Blog:                               github.String(""),
		AdvancedSecurityEnabledForNewRepos: github.Bool(false),
		DependabotAlertsEnabledForNewRepos: github.Bool(false),
		DependabotSecurityUpdatesEnabledForNewRepos:    github.Bool(false),
		DependencyGraphEnabledForNewRepos:              github.Bool(false),
		SecretScanningEnabledForNewRepos:               github.Bool(false),
		SecretScanningPushProtectionEnabledForNewRepos: github.Bool(false),
	}

	enterpriseSettings := github.Organization{
		BillingEmail:                       github.String("email@example.com"),
		Company:                            github.String(""),
		Email:                              github.String(""),
		TwitterUsername:                    github.String(""),
		Location:                           github.String(""),
		Name:                               github.String(""),
		Description:                        github.String(""),
		HasOrganizationProjects:            github.Bool(true),
		HasRepositoryProjects:              github.Bool(true),
		DefaultRepoPermission:              github.String("read"),
		MembersCanCreateRepos:              github.Bool(true),
		MembersCanCreatePrivateRepos:       github.Bool(true),
		MembersCanCreateInternalRepos:      github.Bool(true),
		MembersCanCreatePublicRepos:        github.Bool(true),
		MembersCanCreatePages:              github.Bool(false),
		MembersCanCreatePublicPages:        github.Bool(true),
		MembersCanCreatePrivatePages:       github.Bool(true),
		MembersCanForkPrivateRepos:         github.Bool(false),
		WebCommitSignoffRequired:           github.Bool(false),
		Blog:                               github.String(""),
		AdvancedSecurityEnabledForNewRepos: github.Bool(false),
		DependabotAlertsEnabledForNewRepos: github.Bool(false),
		DependabotSecurityUpdatesEnabledForNewRepos:    github.Bool(false),
		DependencyGraphEnabledForNewRepos:              github.Bool(false),
		SecretScanningEnabledForNewRepos:               github.Bool(false),
		SecretScanningPushProtectionEnabledForNewRepos: github.Bool(false),
	}

	enterpriseSettingsNoFork := github.Organization{
		BillingEmail:                       github.String("email@example.com"),
		Company:                            github.String(""),
		Email:                              github.String(""),
		TwitterUsername:                    github.String(""),
		Location:                           github.String(""),
		Name:                               github.String(""),
		Description:                        github.String(""),
		HasOrganizationProjects:            github.Bool(true),
		HasRepositoryProjects:              github.Bool(true),
		DefaultRepoPermission:              github.String("read"),
		MembersCanCreateRepos:              github.Bool(true),
		MembersCanCreatePrivateRepos:       github.Bool(true),
		MembersCanCreateInternalRepos:      github.Bool(true),
		MembersCanCreatePublicRepos:        github.Bool(true),
		MembersCanCreatePages:              github.Bool(false),
		MembersCanCreatePublicPages:        github.Bool(true),
		MembersCanCreatePrivatePages:       github.Bool(true),
		WebCommitSignoffRequired:           github.Bool(false),
		Blog:                               github.String(""),
		AdvancedSecurityEnabledForNewRepos: github.Bool(false),
		DependabotAlertsEnabledForNewRepos: github.Bool(false),
		DependabotSecurityUpdatesEnabledForNewRepos:    github.Bool(false),
		DependencyGraphEnabledForNewRepos:              github.Bool(false),
		SecretScanningEnabledForNewRepos:               github.Bool(false),
		SecretScanningPushProtectionEnabledForNewRepos: github.Bool(false),
	}

	log.Printf("[DEBUG] Reverting Organization Settings to default values: %s", org)
	orgPlan, _, err := client.Organizations.Edit(ctx, org, nil)
	if err != nil {
		return err
	}
	if orgPlan.GetPlan().GetName() == "enterprise" {
		if _, ok := d.GetOk("members_can_fork_private_repositories"); !ok {
			_, _, err := client.Organizations.Edit(ctx, org, &enterpriseSettingsNoFork)
			if err != nil {
				return err
			}
		} else if _, ok := d.GetOk("members_can_fork_private_repositories"); ok {
			_, _, err := client.Organizations.Edit(ctx, org, &enterpriseSettings)
			if err != nil {
				return err
			}
		}
	} else {
		_, _, err := client.Organizations.Edit(ctx, org, &settings)
		if err != nil {
			return err
		}
	}

	return nil
}
