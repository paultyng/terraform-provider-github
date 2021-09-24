package github

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/google/go-github/v39/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceGithubRepository() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubRepositoryCreate,
		Read:   resourceGithubRepositoryRead,
		Update: resourceGithubRepositoryUpdate,
		Delete: resourceGithubRepositoryDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("auto_init", false)
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"homepage_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"private": {
				Type:          schema.TypeBool,
				Computed:      true, // is affected by "visibility"
				Optional:      true,
				ConflictsWith: []string{"visibility"},
				Deprecated:    "use visibility instead",
			},
			"visibility": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true, // is affected by "private"
				ValidateFunc: validation.StringInSlice([]string{"public", "private", "internal"}, false),
			},
			"has_issues": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"has_projects": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"has_downloads": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"has_wiki": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"is_template": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"allow_merge_commit": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"allow_squash_merge": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"allow_rebase_merge": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"delete_branch_on_merge": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"auto_init": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"default_branch": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Can only be set after initial repository creation, and only if the target branch exists",
				Deprecated:  "Use the github_branch_default resource instead",
			},
			"license_template": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"gitignore_template": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"archived": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"archive_on_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"pages": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"branch": {
										Type:     schema.TypeString,
										Required: true,
									},
									"path": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "/",
									},
								},
							},
						},
						"cname": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"custom_404": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"html_url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"url": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"topics": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`), "must include only lowercase alphanumeric characters or hyphens and cannot start with a hyphen"),
				},
			},
			"vulnerability_alerts": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"full_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"html_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ssh_clone_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"svn_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"git_clone_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"http_clone_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"template": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"owner": {
							Type:     schema.TypeString,
							Required: true,
						},
						"repository": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"node_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"repo_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
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

func resourceGithubRepositoryObject(d *schema.ResourceData) *github.Repository {
	return &github.Repository{
		Name:                github.String(d.Get("name").(string)),
		Description:         github.String(d.Get("description").(string)),
		Homepage:            github.String(d.Get("homepage_url").(string)),
		Visibility:          github.String(calculateVisibility(d)),
		HasDownloads:        github.Bool(d.Get("has_downloads").(bool)),
		HasIssues:           github.Bool(d.Get("has_issues").(bool)),
		HasProjects:         github.Bool(d.Get("has_projects").(bool)),
		HasWiki:             github.Bool(d.Get("has_wiki").(bool)),
		IsTemplate:          github.Bool(d.Get("is_template").(bool)),
		AllowMergeCommit:    github.Bool(d.Get("allow_merge_commit").(bool)),
		AllowSquashMerge:    github.Bool(d.Get("allow_squash_merge").(bool)),
		AllowRebaseMerge:    github.Bool(d.Get("allow_rebase_merge").(bool)),
		DeleteBranchOnMerge: github.Bool(d.Get("delete_branch_on_merge").(bool)),
		AutoInit:            github.Bool(d.Get("auto_init").(bool)),
		LicenseTemplate:     github.String(d.Get("license_template").(string)),
		GitignoreTemplate:   github.String(d.Get("gitignore_template").(string)),
		Archived:            github.Bool(d.Get("archived").(bool)),
		Topics:              expandStringList(d.Get("topics").(*schema.Set).List()),
	}
}

func resourceGithubRepositoryCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	if branchName, hasDefaultBranch := d.GetOk("default_branch"); hasDefaultBranch && (branchName != "main") {
		return fmt.Errorf("Cannot set the default branch on a new repository to something other than 'main'.")
	}

	repoReq := resourceGithubRepositoryObject(d)
	owner := meta.(*Owner).name

	repoName := repoReq.GetName()
	ctx := context.Background()

	log.Printf("[DEBUG] Creating repository: %s/%s", owner, repoName)

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

			templateRepoReq := github.TemplateRepoRequest{
				Name:        &repoName,
				Owner:       &owner,
				Description: github.String(d.Get("description").(string)),
				Private:     github.Bool(isPrivate),
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

	log.Printf("[DEBUG] Reading repository: %s/%s", owner, repoName)

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
				log.Printf("[WARN] Removing repository %s/%s from state because it no longer exists in GitHub",
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
	d.Set("homepage_url", repo.GetHomepage())
	d.Set("private", repo.GetPrivate())
	d.Set("visibility", repo.GetVisibility())
	d.Set("has_issues", repo.GetHasIssues())
	d.Set("has_projects", repo.GetHasProjects())
	d.Set("has_wiki", repo.GetHasWiki())
	d.Set("is_template", repo.GetIsTemplate())
	d.Set("allow_merge_commit", repo.GetAllowMergeCommit())
	d.Set("allow_squash_merge", repo.GetAllowSquashMerge())
	d.Set("allow_rebase_merge", repo.GetAllowRebaseMerge())
	d.Set("delete_branch_on_merge", repo.GetDeleteBranchOnMerge())
	d.Set("has_downloads", repo.GetHasDownloads())
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

	branches, _, err := client.Repositories.ListBranches(ctx, owner, repoName, nil)
	if err != nil {
		return err
	}
	d.Set("branches", flattenBranches(branches))

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
		d.Set("template", []interface{}{
			map[string]interface{}{
				"owner":      repo.TemplateRepository.Owner.Login,
				"repository": repo.TemplateRepository.Name,
			},
		})
	} else {
		d.Set("template", []interface{}{})
	}

	vulnerabilityAlerts, _, err := client.Repositories.GetVulnerabilityAlerts(ctx, owner, repoName)
	if err != nil {
		return fmt.Errorf("Error reading repository vulnerability alerts: %v", err)
	}
	d.Set("vulnerability_alerts", vulnerabilityAlerts)

	return nil
}

func resourceGithubRepositoryUpdate(d *schema.ResourceData, meta interface{}) error {
	// Can only update a repository if it is not archived or the update is to
	// archive the repository (unarchiving is not supported by the Github API)
	if d.Get("archived").(bool) && !d.HasChange("archived") {
		log.Printf("[DEBUG] Skipping update of archived repository")
		return nil
	}

	client := meta.(*Owner).v3client

	repoReq := resourceGithubRepositoryObject(d)

	// handle visibility updates separately from other fields
	repoReq.Visibility = nil

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

	log.Printf("[DEBUG] Updating repository: %s/%s", owner, repoName)
	repo, _, err := client.Repositories.Edit(ctx, owner, repoName, repoReq)
	if err != nil {
		return err
	}
	d.SetId(*repo.Name)

	if d.HasChange("pages") && !d.IsNewResource() {
		opts := expandPagesUpdate(d.Get("pages").([]interface{}))
		if opts != nil {
			_, err := client.Repositories.UpdatePages(ctx, owner, repoName, opts)
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
		_, _, err = client.Repositories.Edit(ctx, owner, repoName, repoReq)
		if err != nil {
			if !strings.Contains(err.Error(), fmt.Sprintf("422 Visibility is already %s", n.(string))) {
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
		log.Printf("[DEBUG] <<<<<<<<<< no privacy update required. private: %v", d.Get("private"))
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
			d.Set("archived", true)
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
	pagesSource := pages["source"].([]interface{})[0].(map[string]interface{})
	source := &github.PagesSource{
		Branch: github.String(pagesSource["branch"].(string)),
	}
	if v, ok := pagesSource["path"].(string); ok {
		// To set to the root directory "/", leave source.Path unset
		if v != "" && v != "/" {
			source.Path = github.String(v)
		}
	}
	return &github.Pages{Source: source}
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

	// To update the Github Pages source, the github.PagesUpdate Source field
	// must include the branch name and optionally the subdirectory /docs.
	// e.g. "master" or "master /docs"
	pagesSource := pages["source"].([]interface{})[0].(map[string]interface{})
	source := pagesSource["branch"].(string)
	if v, ok := pagesSource["path"].(string); ok {
		if v != "" && v != "/" {
			source += fmt.Sprintf(" %s", v)
		}
	}
	update.Source = github.String(source)

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
	pagesMap["url"] = pages.GetURL()
	pagesMap["status"] = pages.GetStatus()
	pagesMap["cname"] = pages.GetCNAME()
	pagesMap["custom_404"] = pages.GetCustom404()
	pagesMap["html_url"] = pages.GetHTMLURL()

	return []interface{}{pagesMap}
}

func flattenBranches(branches []*github.Branch) []interface{} {
	if branches == nil {
		return []interface{}{}
	}

	branchList := make([]interface{}, 0, len(branches))

	for _, branch := range branches {
		branchMap := make(map[string]interface{})
		branchMap["name"] = branch.Name
		branchMap["protected"] = branch.Protected
		branchList = append(branchList, branchMap)
	}

	return branchList
}
