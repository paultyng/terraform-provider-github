package github

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/google/go-github/v66/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceGithubRepository() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubRepositoryCreate,
		Read:   resourceGithubRepositoryRead,
		Update: resourceGithubRepositoryUpdate,
		Delete: resourceGithubRepositoryDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				if err := d.Set("auto_init", false); err != nil {
					return nil, err
				}
				return []*schema.ResourceData{d}, nil
			},
		},

		SchemaVersion: 1,
		MigrateState:  resourceGithubRepositoryMigrateState,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: toDiagFunc(validation.StringMatch(regexp.MustCompile(`^[-a-zA-Z0-9_.]{1,100}$`), "must include only alphanumeric characters, underscores or hyphens and consist of 100 characters or less"), "name"),
				Description:      "The name of the repository.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A description of the repository.",
			},
			"homepage_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "URL of a page describing the project.",
			},
			"private": {
				Type:          schema.TypeBool,
				Computed:      true, // is affected by "visibility"
				Optional:      true,
				ConflictsWith: []string{"visibility"},
				Deprecated:    "use visibility instead",
			},
			"visibility": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true, // is affected by "private"
				ValidateDiagFunc: toDiagFunc(validation.StringInSlice([]string{"public", "private", "internal"}, false), "visibility"),
				Description:      "Can be 'public' or 'private'. If your organization is associated with an enterprise account using GitHub Enterprise Cloud or GitHub Enterprise Server 2.20+, visibility can also be 'internal'.",
			},
			"security_and_analysis": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Security and analysis settings for the repository. To use this parameter you must have admin permissions for the repository or be an owner or security manager for the organization that owns the repository.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"advanced_security": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "The advanced security configuration for the repository. If a repository's visibility is 'public', advanced security is always enabled and cannot be changed, so this setting cannot be supplied.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"status": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: toDiagFunc(validation.StringInSlice([]string{"enabled", "disabled"}, false), "status"),
										Description:      "Set to 'enabled' to enable advanced security features on the repository. Can be 'enabled' or 'disabled'.",
									},
								},
							},
						},
						"secret_scanning": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "The secret scanning configuration for the repository.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"status": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: toDiagFunc(validation.StringInSlice([]string{"enabled", "disabled"}, false), "secret_scanning"),
										Description:      "Set to 'enabled' to enable secret scanning on the repository. Can be 'enabled' or 'disabled'. If set to 'enabled', the repository's visibility must be 'public' or 'security_and_analysis[0].advanced_security[0].status' must also be set to 'enabled'.",
									},
								},
							},
						},
						"secret_scanning_push_protection": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "The secret scanning push protection configuration for the repository.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"status": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: toDiagFunc(validation.StringInSlice([]string{"enabled", "disabled"}, false), "secret_scanning_push_protection"),
										Description:      "Set to 'enabled' to enable secret scanning push protection on the repository. Can be 'enabled' or 'disabled'. If set to 'enabled', the repository's visibility must be 'public' or 'security_and_analysis[0].advanced_security[0].status' must also be set to 'enabled'.",
									},
								},
							},
						},
					},
				},
			},
			"has_issues": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Set to 'true' to enable the GitHub Issues features on the repository",
			},
			"has_discussions": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Set to 'true' to enable GitHub Discussions on the repository. Defaults to 'false'.",
			},
			"has_projects": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Set to 'true' to enable the GitHub Projects features on the repository. Per the GitHub documentation when in an organization that has disabled repository projects it will default to 'false' and will otherwise default to 'true'. If you specify 'true' when it has been disabled it will return an error.",
			},
			"has_downloads": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Set to 'true' to enable the (deprecated) downloads features on the repository.",
			},
			"has_wiki": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Set to 'true' to enable the GitHub Wiki features on the repository.",
			},
			"is_template": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Set to 'true' to tell GitHub that this is a template repository.",
			},
			"allow_merge_commit": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Set to 'false' to disable merge commits on the repository.",
			},
			"allow_squash_merge": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Set to 'false' to disable squash merges on the repository.",
			},
			"allow_rebase_merge": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Set to 'false' to disable rebase merges on the repository.",
			},
			"allow_auto_merge": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Set to 'true' to allow auto-merging pull requests on the repository.",
			},
			"allow_forking": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Set to 'true' to allow private forks.",
			},
			"squash_merge_commit_title": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "COMMIT_OR_PR_TITLE",
				Description: "Can be 'PR_TITLE' or 'COMMIT_OR_PR_TITLE' for a default squash merge commit title.",
			},
			"squash_merge_commit_message": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "COMMIT_MESSAGES",
				Description: "Can be 'PR_BODY', 'COMMIT_MESSAGES', or 'BLANK' for a default squash merge commit message.",
			},
			"merge_commit_title": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "MERGE_MESSAGE",
				Description: "Can be 'PR_TITLE' or 'MERGE_MESSAGE' for a default merge commit title.",
			},
			"merge_commit_message": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "PR_TITLE",
				Description: "Can be 'PR_BODY', 'PR_TITLE', or 'BLANK' for a default merge commit message.",
			},
			"delete_branch_on_merge": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Automatically delete head branch after a pull request is merged. Defaults to 'false'.",
			},
			"web_commit_signoff_required": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Require contributors to sign off on web-based commits. Defaults to 'false'.",
			},
			"auto_init": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Set to 'true' to produce an initial commit in the repository.",
			},
			"default_branch": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Can only be set after initial repository creation, and only if the target branch exists",
				Deprecated:  "Use the github_branch_default resource instead",
			},
			"license_template": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Use the name of the template without the extension. For example, 'mit' or 'mpl-2.0'.",
			},
			"gitignore_template": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Use the name of the template without the extension. For example, 'Haskell'.",
			},
			"archived": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Specifies if the repository should be archived. Defaults to 'false'. NOTE Currently, the API does not support unarchiving.",
			},
			"archive_on_destroy": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Set to 'true' to archive the repository instead of deleting on destroy.",
			},
			"pages": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "The repository's GitHub Pages configuration",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source": {
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Description: "The source branch and directory for the rendered Pages site.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"branch": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The repository branch used to publish the site's source files. (i.e. 'main' or 'gh-pages')",
									},
									"path": {
										Type:        schema.TypeString,
										Optional:    true,
										Default:     "/",
										Description: "The repository directory from which the site publishes (Default: '/')",
									},
								},
							},
						},
						"build_type": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          "legacy",
							Description:      "The type the page should be sourced.",
							ValidateDiagFunc: validateValueFunc([]string{"legacy", "workflow"}),
						},
						"cname": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The custom domain for the repository. This can only be set after the repository has been created.",
						},
						"custom_404": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the rendered GitHub Pages site has a custom 404 page",
						},
						"html_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "URL to the repository on the web.",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The GitHub Pages site's build status e.g. building or built.",
						},
						"url": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"topics": {
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Description: "The list of topics of the repository.",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: toDiagFunc(validation.StringMatch(regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,49}$`), "must include only lowercase alphanumeric characters or hyphens and cannot start with a hyphen and consist of 50 characters or less"), "topics"),
				},
			},
			"vulnerability_alerts": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Set to 'true' to enable security alerts for vulnerable dependencies. Enabling requires alerts to be enabled on the owner level. (Note for importing: GitHub enables the alerts on public repos but disables them on private repos by default). Note that vulnerability alerts have not been successfully tested on any GitHub Enterprise instance and may be unavailable in those settings.",
			},
			"ignore_vulnerability_alerts_during_read": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Set to true to not call the vulnerability alerts endpoint so the resource can also be used without admin permissions during read.",
			},
			"full_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "A string of the form 'orgname/reponame'.",
			},
			"html_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL to the repository on the web.",
			},
			"ssh_clone_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL that can be provided to 'git clone' to clone the repository via SSH.",
			},
			"svn_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL that can be provided to 'svn checkout' to check out the repository via GitHub's Subversion protocol emulation.",
			},
			"git_clone_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL that can be provided to 'git clone' to clone the repository anonymously via the git protocol.",
			},
			"http_clone_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL that can be provided to 'git clone' to clone the repository via HTTPS.",
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"primary_language": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"template": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Use a template repository to create this resource.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"include_all_branches": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether the new repository should include all the branches from the template repository (defaults to 'false', which includes only the default branch from the template).",
						},
						"owner": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The GitHub organization or user the template repository is owned by.",
						},
						"repository": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the template repository.",
						},
					},
				},
			},
			"node_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "GraphQL global node id for use with v4 API.",
			},
			"repo_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "GitHub ID for the repository.",
			},
			"allow_update_branch": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: " Set to 'true' to always suggest updating pull request branches.",
			},
		},
		CustomizeDiff: customDiffFunction,
	}
}

func calculateVisibility(d *schema.ResourceData) string {

	if value, ok := d.GetOk("visibility"); ok {
		return value.(string)
	}

	if value, ok := d.GetOk("private"); ok {
		if value.(bool) {
			return "private"
		} else {
			return "public"
		}
	}

	return "public"
}

func tryGetSecurityAndAnalysisSettingStatus(securityAndAnalysis map[string]interface{}, setting string) (bool, string) {
	value, ok := securityAndAnalysis[setting]
	if !ok {
		return false, ""
	}

	asList := value.([]interface{})
	if len(asList) == 0 || asList[0] == nil {
		return false, ""
	}

	return true, asList[0].(map[string]interface{})["status"].(string)
}

func calculateSecurityAndAnalysis(d *schema.ResourceData) *github.SecurityAndAnalysis {
	value, ok := d.GetOk("security_and_analysis")
	if !ok {
		return nil
	}

	asList := value.([]interface{})
	if len(asList) == 0 || asList[0] == nil {
		return nil
	}

	lookup := asList[0].(map[string]interface{})

	var securityAndAnalysis github.SecurityAndAnalysis

	if ok, status := tryGetSecurityAndAnalysisSettingStatus(lookup, "advanced_security"); ok {
		securityAndAnalysis.AdvancedSecurity = &github.AdvancedSecurity{
			Status: github.String(status),
		}
	}
	if ok, status := tryGetSecurityAndAnalysisSettingStatus(lookup, "secret_scanning"); ok {
		securityAndAnalysis.SecretScanning = &github.SecretScanning{
			Status: github.String(status),
		}
	}
	if ok, status := tryGetSecurityAndAnalysisSettingStatus(lookup, "secret_scanning_push_protection"); ok {
		securityAndAnalysis.SecretScanningPushProtection = &github.SecretScanningPushProtection{
			Status: github.String(status),
		}
	}

	return &securityAndAnalysis
}

func resourceGithubRepositoryObject(d *schema.ResourceData) *github.Repository {
	repository := &github.Repository{
		Name:                     github.String(d.Get("name").(string)),
		Description:              github.String(d.Get("description").(string)),
		Homepage:                 github.String(d.Get("homepage_url").(string)),
		Visibility:               github.String(calculateVisibility(d)),
		HasDownloads:             github.Bool(d.Get("has_downloads").(bool)),
		HasIssues:                github.Bool(d.Get("has_issues").(bool)),
		HasDiscussions:           github.Bool(d.Get("has_discussions").(bool)),
		HasProjects:              github.Bool(d.Get("has_projects").(bool)),
		HasWiki:                  github.Bool(d.Get("has_wiki").(bool)),
		IsTemplate:               github.Bool(d.Get("is_template").(bool)),
		AllowMergeCommit:         github.Bool(d.Get("allow_merge_commit").(bool)),
		AllowSquashMerge:         github.Bool(d.Get("allow_squash_merge").(bool)),
		AllowRebaseMerge:         github.Bool(d.Get("allow_rebase_merge").(bool)),
		AllowAutoMerge:           github.Bool(d.Get("allow_auto_merge").(bool)),
		AllowForking:             github.Bool(d.Get("allow_forking").(bool)),
		DeleteBranchOnMerge:      github.Bool(d.Get("delete_branch_on_merge").(bool)),
		WebCommitSignoffRequired: github.Bool(d.Get("web_commit_signoff_required").(bool)),
		AutoInit:                 github.Bool(d.Get("auto_init").(bool)),
		LicenseTemplate:          github.String(d.Get("license_template").(string)),
		GitignoreTemplate:        github.String(d.Get("gitignore_template").(string)),
		Archived:                 github.Bool(d.Get("archived").(bool)),
		Topics:                   expandStringList(d.Get("topics").(*schema.Set).List()),
		AllowUpdateBranch:        github.Bool(d.Get("allow_update_branch").(bool)),
		SecurityAndAnalysis:      calculateSecurityAndAnalysis(d),
	}

	// only configure merge commit if we are in commit merge strategy
	allowMergeCommit, ok := d.Get("allow_merge_commit").(bool)
	if ok {
		if allowMergeCommit {
			repository.MergeCommitTitle = github.String(d.Get("merge_commit_title").(string))
			repository.MergeCommitMessage = github.String(d.Get("merge_commit_message").(string))
		}
	}

	// only configure squash commit if we are in squash merge strategy
	allowSquashMerge, ok := d.Get("allow_squash_merge").(bool)
	if ok {
		if allowSquashMerge {
			repository.SquashMergeCommitTitle = github.String(d.Get("squash_merge_commit_title").(string))
			repository.SquashMergeCommitMessage = github.String(d.Get("squash_merge_commit_message").(string))
		}
	}

	return repository
}

func resourceGithubRepositoryCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	if branchName, hasDefaultBranch := d.GetOk("default_branch"); hasDefaultBranch && (branchName != "main") {
		return fmt.Errorf("cannot set the default branch on a new repository to something other than 'main'")
	}

	repoReq := resourceGithubRepositoryObject(d)
	owner := meta.(*Owner).name

	repoName := repoReq.GetName()
	ctx := context.Background()

	// determine if repository should be private. assume public to start
	isPrivate := false

	// prefer visibility to private flag since private flag is deprecated
	privateKeyword, ok := d.Get("private").(bool)
	if ok {
		isPrivate = privateKeyword
	}

	visibility, ok := d.Get("visibility").(string)
	if ok {
		if visibility == "private" || visibility == "internal" {
			isPrivate = true
		}
	}

	repoReq.Private = github.Bool(isPrivate)

	if template, ok := d.GetOk("template"); ok {
		templateConfigBlocks := template.([]interface{})

		for _, templateConfigBlock := range templateConfigBlocks {
			templateConfigMap, ok := templateConfigBlock.(map[string]interface{})
			if !ok {
				return errors.New("failed to unpack template configuration block")
			}

			templateRepo := templateConfigMap["repository"].(string)
			templateRepoOwner := templateConfigMap["owner"].(string)
			includeAllBranches := templateConfigMap["include_all_branches"].(bool)

			templateRepoReq := github.TemplateRepoRequest{
				Name:               &repoName,
				Owner:              &owner,
				Description:        github.String(d.Get("description").(string)),
				Private:            github.Bool(isPrivate),
				IncludeAllBranches: github.Bool(includeAllBranches),
			}

			repo, _, err := client.Repositories.CreateFromTemplate(ctx,
				templateRepoOwner,
				templateRepo,
				&templateRepoReq,
			)
			if err != nil {
				return err
			}

			d.SetId(*repo.Name)
		}
	} else {
		// Create without a repository template
		var repo *github.Repository
		var err error
		if meta.(*Owner).IsOrganization {
			repo, _, err = client.Repositories.Create(ctx, owner, repoReq)
		} else {
			// Create repository within authenticated user's account
			repo, _, err = client.Repositories.Create(ctx, "", repoReq)
		}
		if err != nil {
			return err
		}
		d.SetId(repo.GetName())
	}

	topics := repoReq.Topics
	if len(topics) > 0 {
		_, _, err := client.Repositories.ReplaceAllTopics(ctx, owner, repoName, topics)
		if err != nil {
			return err
		}
	}

	pages := expandPages(d.Get("pages").([]interface{}))
	if pages != nil {
		_, _, err := client.Repositories.EnablePages(ctx, owner, repoName, pages)
		if err != nil {
			return err
		}
	}

	return resourceGithubRepositoryUpdate(d, meta)
}

func resourceGithubRepositoryRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	owner := meta.(*Owner).name
	repoName := d.Id()

	// When the user has not authenticated the provider, AnonymousHTTPClient is used, therefore owner == "". In this
	// case lookup the owner in the data, and use that, if present.
	if explicitOwner, _, ok := resourceGithubParseFullName(d); ok && owner == "" {
		owner = explicitOwner
	}

	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	if !d.IsNewResource() {
		ctx = context.WithValue(ctx, ctxEtag, d.Get("etag").(string))
	}

	repo, resp, err := client.Repositories.Get(ctx, owner, repoName)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotModified {
				return nil
			}
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[INFO] Removing repository %s/%s from state because it no longer exists in GitHub",
					owner, repoName)
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.Set("etag", resp.Header.Get("ETag"))
	d.Set("name", repoName)
	d.Set("description", repo.GetDescription())
	d.Set("primary_language", repo.GetLanguage())
	d.Set("homepage_url", repo.GetHomepage())
	d.Set("private", repo.GetPrivate())
	d.Set("visibility", repo.GetVisibility())
	d.Set("has_issues", repo.GetHasIssues())
	d.Set("has_discussions", repo.GetHasDiscussions())
	d.Set("has_projects", repo.GetHasProjects())
	d.Set("has_wiki", repo.GetHasWiki())
	d.Set("is_template", repo.GetIsTemplate())
	d.Set("full_name", repo.GetFullName())
	d.Set("default_branch", repo.GetDefaultBranch())
	d.Set("html_url", repo.GetHTMLURL())
	d.Set("ssh_clone_url", repo.GetSSHURL())
	d.Set("svn_url", repo.GetSVNURL())
	d.Set("git_clone_url", repo.GetGitURL())
	d.Set("http_clone_url", repo.GetCloneURL())
	d.Set("archived", repo.GetArchived())
	d.Set("topics", flattenStringList(repo.Topics))
	d.Set("node_id", repo.GetNodeID())
	d.Set("repo_id", repo.GetID())

	// GitHub API doesn't respond following parameters when repository is archived
	if !d.Get("archived").(bool) {
		d.Set("allow_auto_merge", repo.GetAllowAutoMerge())
		d.Set("allow_forking", repo.GetAllowForking())
		d.Set("allow_merge_commit", repo.GetAllowMergeCommit())
		d.Set("allow_rebase_merge", repo.GetAllowRebaseMerge())
		d.Set("allow_squash_merge", repo.GetAllowSquashMerge())
		d.Set("allow_update_branch", repo.GetAllowUpdateBranch())
		d.Set("delete_branch_on_merge", repo.GetDeleteBranchOnMerge())
		d.Set("web_commit_signoff_required", repo.GetWebCommitSignoffRequired())
		d.Set("has_downloads", repo.GetHasDownloads())
		d.Set("merge_commit_message", repo.GetMergeCommitMessage())
		d.Set("merge_commit_title", repo.GetMergeCommitTitle())
		d.Set("squash_merge_commit_message", repo.GetSquashMergeCommitMessage())
		d.Set("squash_merge_commit_title", repo.GetSquashMergeCommitTitle())
	}

	if repo.GetHasPages() {
		pages, _, err := client.Repositories.GetPagesInfo(ctx, owner, repoName)
		if err != nil {
			return err
		}
		if err := d.Set("pages", flattenPages(pages)); err != nil {
			return fmt.Errorf("error setting pages: %w", err)
		}
	}

	if repo.TemplateRepository != nil {
		if err = d.Set("template", []interface{}{
			map[string]interface{}{
				"owner":      repo.TemplateRepository.Owner.Login,
				"repository": repo.TemplateRepository.Name,
			},
		}); err != nil {
			return err
		}
	} else {
		if err = d.Set("template", []interface{}{}); err != nil {
			return err
		}
	}

	if !d.Get("ignore_vulnerability_alerts_during_read").(bool) {
		vulnerabilityAlerts, _, err := client.Repositories.GetVulnerabilityAlerts(ctx, owner, repoName)
		if err != nil {
			return fmt.Errorf("error reading repository vulnerability alerts: %v", err)
		}
		if err = d.Set("vulnerability_alerts", vulnerabilityAlerts); err != nil {
			return err
		}
	}

	if err = d.Set("security_and_analysis", flattenSecurityAndAnalysis(repo.GetSecurityAndAnalysis())); err != nil {
		return err
	}

	return nil
}

func resourceGithubRepositoryUpdate(d *schema.ResourceData, meta interface{}) error {
	// Can only update a repository if it is not archived or the update is to
	// archive the repository (unarchiving is not supported by the GitHub API)
	if d.Get("archived").(bool) && !d.HasChange("archived") {
		log.Printf("[INFO] Skipping update of archived repository")
		return nil
	}

	client := meta.(*Owner).v3client

	repoReq := resourceGithubRepositoryObject(d)

	// handle visibility updates separately from other fields
	repoReq.Visibility = nil

	if !d.HasChange("security_and_analysis") {
		repoReq.SecurityAndAnalysis = nil
		log.Print("[DEBUG] No security_and_analysis update required. Removing this field from the payload.")
	}

	// The documentation for `default_branch` states: "This can only be set
	// after a repository has already been created". However, for backwards
	// compatibility we need to allow terraform configurations that set
	// `default_branch` to "main" when a repository is created.
	if d.HasChange("default_branch") && !d.IsNewResource() {
		repoReq.DefaultBranch = github.String(d.Get("default_branch").(string))
	}

	repoName := d.Id()
	owner := meta.(*Owner).name
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	repo, _, err := client.Repositories.Edit(ctx, owner, repoName, repoReq)
	if err != nil {
		return err
	}
	d.SetId(*repo.Name)

	if d.HasChange("pages") && !d.IsNewResource() {
		opts := expandPagesUpdate(d.Get("pages").([]interface{}))
		if opts != nil {
			pages, res, err := client.Repositories.GetPagesInfo(ctx, owner, repoName)
			if res.StatusCode != http.StatusNotFound && err != nil {
				return err
			}

			if pages == nil {
				_, _, err = client.Repositories.EnablePages(ctx, owner, repoName, &github.Pages{Source: opts.Source, BuildType: opts.BuildType})
			} else {
				_, err = client.Repositories.UpdatePages(ctx, owner, repoName, opts)
			}
			if err != nil {
				return err
			}
		} else {
			_, err := client.Repositories.DisablePages(ctx, owner, repoName)
			if err != nil {
				return err
			}
		}
	}

	if d.HasChange("topics") {
		topics := repoReq.Topics
		_, _, err = client.Repositories.ReplaceAllTopics(ctx, owner, *repo.Name, topics)
		if err != nil {
			return err
		}
		d.SetId(*repo.Name)

		if d.HasChange("topics") {
			topics := repoReq.Topics
			_, _, err = client.Repositories.ReplaceAllTopics(ctx, owner, *repo.Name, topics)
			if err != nil {
				return err
			}
		}
	}

	if d.HasChange("vulnerability_alerts") {
		updateVulnerabilityAlerts := client.Repositories.DisableVulnerabilityAlerts
		if vulnerabilityAlerts, ok := d.GetOk("vulnerability_alerts"); ok && vulnerabilityAlerts.(bool) {
			updateVulnerabilityAlerts = client.Repositories.EnableVulnerabilityAlerts
		}

		_, err = updateVulnerabilityAlerts(ctx, owner, repoName)
		if err != nil {
			return err
		}
	}

	if d.HasChange("visibility") {
		o, n := d.GetChange("visibility")
		repoReq.Visibility = github.String(n.(string))
		log.Printf("[DEBUG] Updating repository visibility from %s to %s", o, n)
		_, resp, err := client.Repositories.Edit(ctx, owner, repoName, repoReq)
		if err != nil {
			if resp.StatusCode != 422 || !strings.Contains(err.Error(), fmt.Sprintf("Visibility is already %s", n.(string))) {
				return err
			}
		}
	} else {
		log.Printf("[DEBUG] No visibility update required. visibility: %s", d.Get("visibility"))
	}

	if d.HasChange("private") {
		o, n := d.GetChange("private")
		repoReq.Private = github.Bool(n.(bool))
		log.Printf("[DEBUG] Updating repository privacy from %v to %v", o, n)
		_, _, err = client.Repositories.Edit(ctx, owner, repoName, repoReq)
		if err != nil {
			if !strings.Contains(err.Error(), "422 Privacy is already set") {
				return err
			}
		}
	} else {
		log.Printf("[DEBUG] No privacy update required. private: %v", d.Get("private"))
	}

	return resourceGithubRepositoryRead(d, meta)
}

func resourceGithubRepositoryDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	repoName := d.Id()
	owner := meta.(*Owner).name
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	archiveOnDestroy := d.Get("archive_on_destroy").(bool)
	if archiveOnDestroy {
		if d.Get("archived").(bool) {
			log.Printf("[DEBUG] Repository already archived, nothing to do on delete: %s/%s", owner, repoName)
			return nil
		} else {
			if err := d.Set("archived", true); err != nil {
				return err
			}
			repoReq := resourceGithubRepositoryObject(d)
			log.Printf("[DEBUG] Archiving repository on delete: %s/%s", owner, repoName)
			_, _, err := client.Repositories.Edit(ctx, owner, repoName, repoReq)
			return err
		}
	}

	log.Printf("[DEBUG] Deleting repository: %s/%s", owner, repoName)
	_, err := client.Repositories.Delete(ctx, owner, repoName)
	return err
}

func expandPages(input []interface{}) *github.Pages {
	if len(input) == 0 || input[0] == nil {
		return nil
	}
	pages := input[0].(map[string]interface{})
	source := &github.PagesSource{
		Branch: github.String("main"),
	}
	if len(pages["source"].([]interface{})) == 1 {
		if pagesSource, ok := pages["source"].([]interface{})[0].(map[string]interface{}); ok {
			if v, ok := pagesSource["branch"].(string); ok {
				source.Branch = github.String(v)
			}
			if v, ok := pagesSource["path"].(string); ok {
				// To set to the root directory "/", leave source.Path unset
				if v != "" && v != "/" {
					source.Path = github.String(v)
				}
			}
		}
	}

	var buildType *string
	if v, ok := pages["build_type"].(string); ok {
		buildType = github.String(v)
	}

	return &github.Pages{Source: source, BuildType: buildType}
}

func expandPagesUpdate(input []interface{}) *github.PagesUpdate {
	if len(input) == 0 || input[0] == nil {
		return nil
	}

	pages := input[0].(map[string]interface{})
	update := &github.PagesUpdate{}

	// Only set the github.PagesUpdate CNAME field if the value is a non-empty string.
	// Leaving the CNAME field unset will remove the custom domain.
	if v, ok := pages["cname"].(string); ok && v != "" {
		update.CNAME = github.String(v)
	}

	// Only set the github.PagesUpdate BuildType field if the value is a non-empty string.
	if v, ok := pages["build_type"].(string); ok && v != "" {
		update.BuildType = github.String(v)
	}

	// To update the GitHub Pages source, the github.PagesUpdate Source field
	// must include the branch name and optionally the subdirectory /docs.
	// e.g. "master" or "master /docs"
	// This is only necessary if the BuildType is "legacy".
	if update.BuildType == nil || *update.BuildType == "legacy" {
		pagesSource := pages["source"].([]interface{})[0].(map[string]interface{})
		sourceBranch := pagesSource["branch"].(string)
		sourcePath := ""
		if v, ok := pagesSource["path"].(string); ok && v != "" {
			sourcePath = v
		}
		update.Source = &github.PagesSource{Branch: &sourceBranch, Path: &sourcePath}
	}

	return update
}

func flattenPages(pages *github.Pages) []interface{} {
	if pages == nil {
		return []interface{}{}
	}

	sourceMap := make(map[string]interface{})
	sourceMap["branch"] = pages.GetSource().GetBranch()
	sourceMap["path"] = pages.GetSource().GetPath()

	pagesMap := make(map[string]interface{})
	pagesMap["source"] = []interface{}{sourceMap}
	pagesMap["build_type"] = pages.GetBuildType()
	pagesMap["url"] = pages.GetURL()
	pagesMap["status"] = pages.GetStatus()
	pagesMap["cname"] = pages.GetCNAME()
	pagesMap["custom_404"] = pages.GetCustom404()
	pagesMap["html_url"] = pages.GetHTMLURL()

	return []interface{}{pagesMap}
}

func flattenRepositoryLicense(repositorylicense *github.RepositoryLicense) []interface{} {
	if repositorylicense == nil {
		return []interface{}{}
	}

	licenseMap := make(map[string]interface{})
	licenseMap["key"] = repositorylicense.GetLicense().GetKey()
	licenseMap["name"] = repositorylicense.GetLicense().GetName()
	licenseMap["url"] = repositorylicense.GetLicense().GetURL()
	licenseMap["spdx_id"] = repositorylicense.GetLicense().GetSPDXID()
	licenseMap["html_url"] = repositorylicense.GetLicense().GetHTMLURL()
	licenseMap["featured"] = repositorylicense.GetLicense().GetFeatured()
	licenseMap["description"] = repositorylicense.GetLicense().GetDescription()
	licenseMap["implementation"] = repositorylicense.GetLicense().GetImplementation()
	licenseMap["permissions"] = repositorylicense.GetLicense().GetPermissions()
	licenseMap["conditions"] = repositorylicense.GetLicense().GetConditions()
	licenseMap["limitations"] = repositorylicense.GetLicense().GetLimitations()
	licenseMap["body"] = repositorylicense.GetLicense().GetBody()

	repositorylicenseMap := make(map[string]interface{})
	repositorylicenseMap["license"] = []interface{}{licenseMap}
	repositorylicenseMap["name"] = repositorylicense.GetName()
	repositorylicenseMap["path"] = repositorylicense.GetPath()
	repositorylicenseMap["sha"] = repositorylicense.GetSHA()
	repositorylicenseMap["size"] = repositorylicense.GetSize()
	repositorylicenseMap["url"] = repositorylicense.GetURL()
	repositorylicenseMap["html_url"] = repositorylicense.GetHTMLURL()
	repositorylicenseMap["git_url"] = repositorylicense.GetGitURL()
	repositorylicenseMap["download_url"] = repositorylicense.GetDownloadURL()
	repositorylicenseMap["type"] = repositorylicense.GetType()
	repositorylicenseMap["content"] = repositorylicense.GetContent()
	repositorylicenseMap["encoding"] = repositorylicense.GetEncoding()

	return []interface{}{repositorylicenseMap}
}

func flattenSecurityAndAnalysis(securityAndAnalysis *github.SecurityAndAnalysis) []interface{} {
	if securityAndAnalysis == nil {
		return []interface{}{}
	}

	securityAndAnalysisMap := make(map[string]interface{})

	advancedSecurity := securityAndAnalysis.GetAdvancedSecurity()
	if advancedSecurity != nil {
		securityAndAnalysisMap["advanced_security"] = []interface{}{map[string]interface{}{
			"status": advancedSecurity.GetStatus(),
		}}
	}

	securityAndAnalysisMap["secret_scanning"] = []interface{}{map[string]interface{}{
		"status": securityAndAnalysis.GetSecretScanning().GetStatus(),
	}}

	securityAndAnalysisMap["secret_scanning_push_protection"] = []interface{}{map[string]interface{}{
		"status": securityAndAnalysis.GetSecretScanningPushProtection().GetStatus(),
	}}

	return []interface{}{securityAndAnalysisMap}
}

// In case full_name can be determined from the data, parses it into an org and repo name proper. For example,
// resourceGithubParseFullName will return "myorg", "myrepo", true when full_name is "myorg/myrepo".
func resourceGithubParseFullName(resourceDataLike interface {
	GetOk(string) (interface{}, bool)
}) (string, string, bool) {
	x, ok := resourceDataLike.GetOk("full_name")
	if !ok {
		return "", "", false
	}
	s, ok := x.(string)
	if !ok || s == "" {
		return "", "", false
	}
	parts := strings.Split(s, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func customDiffFunction(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if diff.HasChange("name") {
		if err := diff.SetNewComputed("full_name"); err != nil {
			return err
		}
	}
	return nil
}
