package github

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/shurcooL/githubv4"
)

func resourceGithubBranchProtection() *schema.Resource {
	return &schema.Resource{
		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			// Input
			REPOSITORY_ID: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Node ID or name of repository",
			},
			PROTECTION_PATTERN: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "",
			},
			PROTECTION_ALLOWS_DELETIONS: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			PROTECTION_ALLOWS_FORCE_PUSHES: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			PROTECTION_IS_ADMIN_ENFORCED: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			PROTECTION_REQUIRES_COMMIT_SIGNATURES: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			PROTECTION_REQUIRES_CONVERSATION_RESOLUTION: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			PROTECTION_REQUIRES_APPROVING_REVIEWS: {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						PROTECTION_REQUIRED_APPROVING_REVIEW_COUNT: {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      1,
							ValidateFunc: validation.IntBetween(1, 6),
						},
						PROTECTION_REQUIRES_CODE_OWNER_REVIEWS: {
							Type:     schema.TypeBool,
							Optional: true,
						},
						PROTECTION_DISMISSES_STALE_REVIEWS: {
							Type:     schema.TypeBool,
							Optional: true,
						},
						PROTECTION_RESTRICTS_REVIEW_DISMISSALS: {
							Type:     schema.TypeBool,
							Optional: true,
						},
						PROTECTION_RESTRICTS_REVIEW_DISMISSERS: {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			PROTECTION_REQUIRES_STATUS_CHECKS: {
				Type:             schema.TypeList,
				Optional:         true,
				DiffSuppressFunc: statusChecksDiffSuppression,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						PROTECTION_REQUIRES_STRICT_STATUS_CHECKS: {
							Type:     schema.TypeBool,
							Optional: true,
						},
						PROTECTION_REQUIRED_STATUS_CHECK_CONTEXTS: {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			PROTECTION_RESTRICTS_PUSHES: {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},

		Create: resourceGithubBranchProtectionCreate,
		Read:   resourceGithubBranchProtectionRead,
		Update: resourceGithubBranchProtectionUpdate,
		Delete: resourceGithubBranchProtectionDelete,

		Importer: &schema.ResourceImporter{
			State: resourceGithubBranchProtectionImport,
		},

		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceGithubBranchProtectionV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceGithubBranchProtectionUpgradeV0,
				Version: 0,
			},
		},
	}
}

func resourceGithubBranchProtectionCreate(d *schema.ResourceData, meta interface{}) error {
	var mutate struct {
		CreateBranchProtectionRule struct {
			BranchProtectionRule struct {
				ID githubv4.ID
			}
		} `graphql:"createBranchProtectionRule(input: $input)"`
	}
	data, err := branchProtectionResourceData(d, meta)
	if err != nil {
		return err
	}
	input := githubv4.CreateBranchProtectionRuleInput{
		AllowsDeletions:                githubv4.NewBoolean(githubv4.Boolean(data.AllowsDeletions)),
		AllowsForcePushes:              githubv4.NewBoolean(githubv4.Boolean(data.AllowsForcePushes)),
		DismissesStaleReviews:          githubv4.NewBoolean(githubv4.Boolean(data.DismissesStaleReviews)),
		IsAdminEnforced:                githubv4.NewBoolean(githubv4.Boolean(data.IsAdminEnforced)),
		Pattern:                        githubv4.String(data.Pattern),
		PushActorIDs:                   githubv4NewIDSlice(githubv4IDSlice(data.PushActorIDs)),
		RepositoryID:                   githubv4.NewID(githubv4.ID(data.RepositoryID)),
		RequiredApprovingReviewCount:   githubv4.NewInt(githubv4.Int(data.RequiredApprovingReviewCount)),
		RequiredStatusCheckContexts:    githubv4NewStringSlice(githubv4StringSlice(data.RequiredStatusCheckContexts)),
		RequiresApprovingReviews:       githubv4.NewBoolean(githubv4.Boolean(data.RequiresApprovingReviews)),
		RequiresCodeOwnerReviews:       githubv4.NewBoolean(githubv4.Boolean(data.RequiresCodeOwnerReviews)),
		RequiresCommitSignatures:       githubv4.NewBoolean(githubv4.Boolean(data.RequiresCommitSignatures)),
		RequiresConversationResolution: githubv4.NewBoolean(githubv4.Boolean(data.RequiresConversationResolution)),
		RequiresStatusChecks:           githubv4.NewBoolean(githubv4.Boolean(data.RequiresStatusChecks)),
		RequiresStrictStatusChecks:     githubv4.NewBoolean(githubv4.Boolean(data.RequiresStrictStatusChecks)),
		RestrictsPushes:                githubv4.NewBoolean(githubv4.Boolean(data.RestrictsPushes)),
		RestrictsReviewDismissals:      githubv4.NewBoolean(githubv4.Boolean(data.RestrictsReviewDismissals)),
		ReviewDismissalActorIDs:        githubv4NewIDSlice(githubv4IDSlice(data.ReviewDismissalActorIDs)),
	}

	ctx := context.Background()
	client := meta.(*Owner).v4client
	err = client.Mutate(ctx, &mutate, input, nil)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%s", mutate.CreateBranchProtectionRule.BranchProtectionRule.ID))

	return resourceGithubBranchProtectionRead(d, meta)
}

func resourceGithubBranchProtectionRead(d *schema.ResourceData, meta interface{}) error {
	var query struct {
		Node struct {
			Node BranchProtectionRule `graphql:"... on BranchProtectionRule"`
		} `graphql:"node(id: $id)"`
	}
	variables := map[string]interface{}{
		"id": d.Id(),
	}

	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	client := meta.(*Owner).v4client
	err := client.Query(ctx, &query, variables)
	if err != nil {
		if strings.Contains(err.Error(), "Could not resolve to a node with the global id") {
			log.Printf("[WARN] Removing branch protection (%s) from state because it no longer exists in GitHub", d.Id())
			d.SetId("")
			return nil
		}

		return err
	}

	protection := query.Node.Node

	err = d.Set(PROTECTION_PATTERN, protection.Pattern)
	if err != nil {
		log.Printf("[WARN] Problem setting '%s' in %s %s branch protection (%s)", PROTECTION_PATTERN, protection.Repository.Name, protection.Pattern, d.Id())
	}

	err = d.Set(PROTECTION_ALLOWS_DELETIONS, protection.AllowsDeletions)
	if err != nil {
		log.Printf("[WARN] Problem setting '%s' in %s %s branch protection (%s)", PROTECTION_ALLOWS_DELETIONS, protection.Repository.Name, protection.Pattern, d.Id())
	}

	err = d.Set(PROTECTION_ALLOWS_FORCE_PUSHES, protection.AllowsForcePushes)
	if err != nil {
		log.Printf("[WARN] Problem setting '%s' in %s %s branch protection (%s)", PROTECTION_ALLOWS_FORCE_PUSHES, protection.Repository.Name, protection.Pattern, d.Id())
	}

	err = d.Set(PROTECTION_IS_ADMIN_ENFORCED, protection.IsAdminEnforced)
	if err != nil {
		log.Printf("[WARN] Problem setting '%s' in %s %s branch protection (%s)", PROTECTION_IS_ADMIN_ENFORCED, protection.Repository.Name, protection.Pattern, d.Id())
	}

	err = d.Set(PROTECTION_REQUIRES_COMMIT_SIGNATURES, protection.RequiresCommitSignatures)
	if err != nil {
		log.Printf("[WARN] Problem setting '%s' in %s %s branch protection (%s)", PROTECTION_REQUIRES_COMMIT_SIGNATURES, protection.Repository.Name, protection.Pattern, d.Id())
	}

	err = d.Set(PROTECTION_REQUIRES_CONVERSATION_RESOLUTION, protection.RequiresConversationResolution)
	if err != nil {
		log.Printf("[WARN] Problem setting '%s' in %s %s branch protection (%s)", PROTECTION_REQUIRES_CONVERSATION_RESOLUTION, protection.Repository.Name, protection.Pattern, d.Id())
	}

	approvingReviews := setApprovingReviews(protection)
	err = d.Set(PROTECTION_REQUIRES_APPROVING_REVIEWS, approvingReviews)
	if err != nil {
		log.Printf("[WARN] Problem setting '%s' in %s %s branch protection (%s)", PROTECTION_REQUIRES_APPROVING_REVIEWS, protection.Repository.Name, protection.Pattern, d.Id())
	}

	statusChecks := setStatusChecks(protection)
	err = d.Set(PROTECTION_REQUIRES_STATUS_CHECKS, statusChecks)
	if err != nil {
		log.Printf("[WARN] Problem setting '%s' in %s %s branch protection (%s)", PROTECTION_REQUIRES_STATUS_CHECKS, protection.Repository.Name, protection.Pattern, d.Id())
	}

	restrictsPushes := setPushes(protection)
	err = d.Set(PROTECTION_RESTRICTS_PUSHES, restrictsPushes)
	if err != nil {
		log.Printf("[WARN] Problem setting '%s' in %s %s branch protection (%s)", PROTECTION_RESTRICTS_PUSHES, protection.Repository.Name, protection.Pattern, d.Id())
	}

	return nil
}

func resourceGithubBranchProtectionUpdate(d *schema.ResourceData, meta interface{}) error {
	var mutate struct {
		UpdateBranchProtectionRule struct {
			BranchProtectionRule struct {
				ID githubv4.ID
			}
		} `graphql:"updateBranchProtectionRule(input: $input)"`
	}
	data, err := branchProtectionResourceData(d, meta)
	if err != nil {
		return err
	}
	input := githubv4.UpdateBranchProtectionRuleInput{
		BranchProtectionRuleID:         d.Id(),
		AllowsDeletions:                githubv4.NewBoolean(githubv4.Boolean(data.AllowsDeletions)),
		AllowsForcePushes:              githubv4.NewBoolean(githubv4.Boolean(data.AllowsForcePushes)),
		DismissesStaleReviews:          githubv4.NewBoolean(githubv4.Boolean(data.DismissesStaleReviews)),
		IsAdminEnforced:                githubv4.NewBoolean(githubv4.Boolean(data.IsAdminEnforced)),
		Pattern:                        githubv4.NewString(githubv4.String(data.Pattern)),
		PushActorIDs:                   githubv4NewIDSlice(githubv4IDSlice(data.PushActorIDs)),
		RequiredApprovingReviewCount:   githubv4.NewInt(githubv4.Int(data.RequiredApprovingReviewCount)),
		RequiredStatusCheckContexts:    githubv4NewStringSlice(githubv4StringSlice(data.RequiredStatusCheckContexts)),
		RequiresApprovingReviews:       githubv4.NewBoolean(githubv4.Boolean(data.RequiresApprovingReviews)),
		RequiresCodeOwnerReviews:       githubv4.NewBoolean(githubv4.Boolean(data.RequiresCodeOwnerReviews)),
		RequiresCommitSignatures:       githubv4.NewBoolean(githubv4.Boolean(data.RequiresCommitSignatures)),
		RequiresConversationResolution: githubv4.NewBoolean(githubv4.Boolean(data.RequiresConversationResolution)),
		RequiresStatusChecks:           githubv4.NewBoolean(githubv4.Boolean(data.RequiresStatusChecks)),
		RequiresStrictStatusChecks:     githubv4.NewBoolean(githubv4.Boolean(data.RequiresStrictStatusChecks)),
		RestrictsPushes:                githubv4.NewBoolean(githubv4.Boolean(data.RestrictsPushes)),
		RestrictsReviewDismissals:      githubv4.NewBoolean(githubv4.Boolean(data.RestrictsReviewDismissals)),
		ReviewDismissalActorIDs:        githubv4NewIDSlice(githubv4IDSlice(data.ReviewDismissalActorIDs)),
	}

	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	client := meta.(*Owner).v4client
	err = client.Mutate(ctx, &mutate, input, nil)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%s", mutate.UpdateBranchProtectionRule.BranchProtectionRule.ID))

	return resourceGithubBranchProtectionRead(d, meta)
}

func resourceGithubBranchProtectionDelete(d *schema.ResourceData, meta interface{}) error {
	var mutate struct {
		DeleteBranchProtectionRule struct { // Empty struct does not work
			ClientMutationId githubv4.ID
		} `graphql:"deleteBranchProtectionRule(input: $input)"`
	}
	input := githubv4.DeleteBranchProtectionRuleInput{
		BranchProtectionRuleID: d.Id(),
	}

	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	client := meta.(*Owner).v4client
	err := client.Mutate(ctx, &mutate, input, nil)

	return err
}

func resourceGithubBranchProtectionImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	repoName, pattern, err := parseTwoPartID(d.Id(), "repository", "pattern")
	if err != nil {
		return nil, err
	}

	repoID, err := getRepositoryID(repoName, meta)
	if err != nil {
		return nil, err
	}
	d.Set("repository_id", repoID)

	id, err := getBranchProtectionID(repoID, pattern, meta)
	if err != nil {
		return nil, err
	}
	d.SetId(fmt.Sprintf("%s", id))

	return []*schema.ResourceData{d}, resourceGithubBranchProtectionRead(d, meta)
}
