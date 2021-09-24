package github

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/google/go-github/v39/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceGithubBranchProtectionV3() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubBranchProtectionV3Create,
		Read:   resourceGithubBranchProtectionV3Read,
		Update: resourceGithubBranchProtectionV3Update,
		Delete: resourceGithubBranchProtectionV3Delete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"branch": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"required_status_checks": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"include_admins": {
							Type:       schema.TypeBool,
							Optional:   true,
							Default:    false,
							Deprecated: "Use enforce_admins instead",
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								return true
							},
						},
						"strict": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"contexts": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"required_pull_request_reviews": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// FIXME: Remove this deprecated field
						"include_admins": {
							Type:       schema.TypeBool,
							Optional:   true,
							Default:    false,
							Deprecated: "Use enforce_admins instead",
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								return true
							},
						},
						"dismiss_stale_reviews": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"dismissal_users": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"dismissal_teams": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"require_code_owner_reviews": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"required_approving_review_count": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      1,
							ValidateFunc: validation.IntBetween(1, 6),
						},
					},
				},
			},
			"restrictions": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"users": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"teams": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"apps": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"enforce_admins": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"require_signed_commits": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceGithubBranchProtectionV3Create(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client

	orgName := meta.(*Owner).name
	repoName := d.Get("repository").(string)
	branch := d.Get("branch").(string)

	protectionRequest, err := buildProtectionRequest(d)
	if err != nil {
		return err
	}
	ctx := context.Background()

	log.Printf("[DEBUG] Creating branch protection: %s/%s (%s)",
		orgName, repoName, branch)
	protection, _, err := client.Repositories.UpdateBranchProtection(ctx,
		orgName,
		repoName,
		branch,
		protectionRequest,
	)
	if err != nil {
		return err
	}

	if err := checkBranchRestrictionsUsers(protection.GetRestrictions(), protectionRequest.GetRestrictions()); err != nil {
		return err
	}

	d.SetId(buildTwoPartID(repoName, branch))

	if err = requireSignedCommitsUpdate(d, meta); err != nil {
		return err
	}

	return resourceGithubBranchProtectionV3Read(d, meta)
}

func resourceGithubBranchProtectionV3Read(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client

	repoName, branch, err := parseTwoPartID(d.Id(), "repository", "branch")
	if err != nil {
		return err
	}
	orgName := meta.(*Owner).name

	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	if !d.IsNewResource() {
		ctx = context.WithValue(ctx, ctxEtag, d.Get("etag").(string))
	}

	log.Printf("[DEBUG] Reading branch protection: %s/%s (%s)",
		orgName, repoName, branch)
	githubProtection, resp, err := client.Repositories.GetBranchProtection(ctx,
		orgName, repoName, branch)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotModified {
				if err := requireSignedCommitsRead(d, meta); err != nil {
					return fmt.Errorf("Error setting signed commit restriction: %v", err)
				}
				return nil
			}
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[WARN] Removing branch protection %s/%s (%s) from state because it no longer exists in GitHub",
					orgName, repoName, branch)
				d.SetId("")
				return nil
			}
		}

		return err
	}

	d.Set("etag", resp.Header.Get("ETag"))
	d.Set("repository", repoName)
	d.Set("branch", branch)
	d.Set("enforce_admins", githubProtection.GetEnforceAdmins().Enabled)

	if err := flattenAndSetRequiredStatusChecks(d, githubProtection); err != nil {
		return fmt.Errorf("Error setting required_status_checks: %v", err)
	}

	if err := flattenAndSetRequiredPullRequestReviews(d, githubProtection); err != nil {
		return fmt.Errorf("Error setting required_pull_request_reviews: %v", err)
	}

	if err := flattenAndSetRestrictions(d, githubProtection); err != nil {
		return fmt.Errorf("Error setting restrictions: %v", err)
	}

	if err := requireSignedCommitsRead(d, meta); err != nil {
		return fmt.Errorf("Error setting signed commit restriction: %v", err)
	}

	return nil
}

func resourceGithubBranchProtectionV3Update(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client
	repoName, branch, err := parseTwoPartID(d.Id(), "repository", "branch")
	if err != nil {
		return err
	}

	protectionRequest, err := buildProtectionRequest(d)
	if err != nil {
		return err
	}

	orgName := meta.(*Owner).name
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	log.Printf("[DEBUG] Updating branch protection: %s/%s (%s)",
		orgName, repoName, branch)
	protection, _, err := client.Repositories.UpdateBranchProtection(ctx,
		orgName,
		repoName,
		branch,
		protectionRequest,
	)
	if err != nil {
		return err
	}

	if err := checkBranchRestrictionsUsers(protection.GetRestrictions(), protectionRequest.GetRestrictions()); err != nil {
		return err
	}

	if protectionRequest.RequiredPullRequestReviews == nil {
		_, err = client.Repositories.RemovePullRequestReviewEnforcement(ctx,
			orgName,
			repoName,
			branch,
		)
		if err != nil {
			return err
		}
	}

	d.SetId(buildTwoPartID(repoName, branch))

	if err = requireSignedCommitsUpdate(d, meta); err != nil {
		return err
	}

	return resourceGithubBranchProtectionV3Read(d, meta)
}

func resourceGithubBranchProtectionV3Delete(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client
	repoName, branch, err := parseTwoPartID(d.Id(), "repository", "branch")
	if err != nil {
		return err
	}

	orgName := meta.(*Owner).name
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	log.Printf("[DEBUG] Deleting branch protection: %s/%s (%s)", orgName, repoName, branch)
	_, err = client.Repositories.RemoveBranchProtection(ctx,
		orgName, repoName, branch)
	return err
}
