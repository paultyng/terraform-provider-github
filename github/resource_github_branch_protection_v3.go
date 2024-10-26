package github

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceGithubBranchProtectionV3() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubBranchProtectionV3Create,
		Read:   resourceGithubBranchProtectionV3Read,
		Update: resourceGithubBranchProtectionV3Update,
		Delete: resourceGithubBranchProtectionV3Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The GitHub repository name.",
			},
			"branch": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The Git branch to protect.",
			},
			"required_status_checks": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Enforce restrictions for required status checks.",
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
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Require branches to be up to date before merging.",
						},
						"contexts": {
							Type:       schema.TypeSet,
							Optional:   true,
							Deprecated: "GitHub is deprecating the use of `contexts`. Use a `checks` array instead.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"checks": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "The list of status checks to require in order to merge into this branch. No status checks are required by default. Checks should be strings containing the 'context' and 'app_id' like so 'context:app_id'",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"required_pull_request_reviews": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Enforce restrictions for pull request reviews.",
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
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Dismiss approved reviews automatically when a new commit is pushed.",
						},
						"dismissal_users": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "The list of user logins with dismissal access.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"dismissal_teams": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "The list of team slugs with dismissal access. Always use slug of the team, not its name. Each team already has to have access to the repository.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"dismissal_apps": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "The list of apps slugs with dismissal access. Always use slug of the app, not its name. Each app already has to have access to the repository.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"require_code_owner_reviews": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Require an approved review in pull requests including files with a designated code owner.",
						},
						"required_approving_review_count": {
							Type:             schema.TypeInt,
							Optional:         true,
							Default:          1,
							Description:      "Require 'x' number of approvals to satisfy branch protection requirements. If this is specified it must be a number between 0-6.",
							ValidateDiagFunc: toDiagFunc(validation.IntBetween(0, 6), "required_approving_review_count"),
						},
						"require_last_push_approval": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Require that the most recent push must be approved by someone other than the last pusher.",
						},
						"bypass_pull_request_allowances": {
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
					},
				},
			},
			"restrictions": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Enforce restrictions for the users and teams that may push to the branch.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"users": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "The list of user logins with push access.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"teams": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "The list of team slugs with push access. Always use slug of the team, not its name. Each team already has to have access to the repository.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"apps": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "The list of app slugs with push access.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"enforce_admins": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Setting this to 'true' enforces status checks for repository administrators.",
			},
			"require_signed_commits": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Setting this to 'true' requires all commits to be signed with GPG.",
			},
			"require_conversation_resolution": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Setting this to 'true' requires all conversations on code must be resolved before a pull request can be merged.",
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

	githubProtection, resp, err := client.Repositories.GetBranchProtection(ctx,
		orgName, repoName, branch)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotModified {
				if err := requireSignedCommitsRead(d, meta); err != nil {
					return fmt.Errorf("error setting signed commit restriction: %v", err)
				}
				return nil
			}
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[INFO] Removing branch protection %s/%s (%s) from state because it no longer exists in GitHub",
					orgName, repoName, branch)
				d.SetId("")
				return nil
			}
		}

		return err
	}

	if err = d.Set("etag", resp.Header.Get("ETag")); err != nil {
		return err
	}
	if err = d.Set("repository", repoName); err != nil {
		return err
	}
	if err = d.Set("branch", branch); err != nil {
		return err
	}
	if err = d.Set("enforce_admins", githubProtection.GetEnforceAdmins().Enabled); err != nil {
		return err
	}
	if rcr := githubProtection.GetRequiredConversationResolution(); rcr != nil {
		if err = d.Set("require_conversation_resolution", rcr.Enabled); err != nil {
			return err
		}
	}

	if err := flattenAndSetRequiredStatusChecks(d, githubProtection); err != nil {
		return fmt.Errorf("error setting required_status_checks: %v", err)
	}

	if err := flattenAndSetRequiredPullRequestReviews(d, githubProtection); err != nil {
		return fmt.Errorf("error setting required_pull_request_reviews: %v", err)
	}

	if err := flattenAndSetRestrictions(d, githubProtection); err != nil {
		return fmt.Errorf("error setting restrictions: %v", err)
	}

	if err := requireSignedCommitsRead(d, meta); err != nil {
		return fmt.Errorf("error setting signed commit restriction: %v", err)
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

	_, err = client.Repositories.RemoveBranchProtection(ctx,
		orgName, repoName, branch)
	return err
}
