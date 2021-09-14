package github

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/shurcooL/githubv4"
)

type Actor struct {
	ID   githubv4.ID
	Name githubv4.String
}

type DismissalActorTypes struct {
	Actor struct {
		Team Actor `graphql:"... on Team"`
		User Actor `graphql:"... on User"`
	}
}

type PushActorTypes struct {
	Actor struct {
		App  Actor `graphql:"... on App"`
		Team Actor `graphql:"... on Team"`
		User Actor `graphql:"... on User"`
	}
}

type BranchProtectionRule struct {
	Repository struct {
		ID   githubv4.String
		Name githubv4.String
	}
	PushAllowances struct {
		Nodes []PushActorTypes
	} `graphql:"pushAllowances(first: 100)"`
	ReviewDismissalAllowances struct {
		Nodes []DismissalActorTypes
	} `graphql:"reviewDismissalAllowances(first: 100)"`
	AllowsDeletions                githubv4.Boolean
	AllowsForcePushes              githubv4.Boolean
	DismissesStaleReviews          githubv4.Boolean
	ID                             githubv4.ID
	IsAdminEnforced                githubv4.Boolean
	Pattern                        githubv4.String
	RequiredApprovingReviewCount   githubv4.Int
	RequiredStatusCheckContexts    []githubv4.String
	RequiresApprovingReviews       githubv4.Boolean
	RequiresCodeOwnerReviews       githubv4.Boolean
	RequiresCommitSignatures       githubv4.Boolean
	RequiresConversationResolution githubv4.Boolean
	RequiresStatusChecks           githubv4.Boolean
	RequiresStrictStatusChecks     githubv4.Boolean
	RestrictsPushes                githubv4.Boolean
	RestrictsReviewDismissals      githubv4.Boolean
}

type BranchProtectionResourceData struct {
	AllowsDeletions                bool
	AllowsForcePushes              bool
	BranchProtectionRuleID         string
	DismissesStaleReviews          bool
	IsAdminEnforced                bool
	Pattern                        string
	PushActorIDs                   []string
	RepositoryID                   string
	RequiredApprovingReviewCount   int
	RequiredStatusCheckContexts    []string
	RequiresApprovingReviews       bool
	RequiresCodeOwnerReviews       bool
	RequiresCommitSignatures       bool
	RequiresConversationResolution bool
	RequiresStatusChecks           bool
	RequiresStrictStatusChecks     bool
	RestrictsPushes                bool
	RestrictsReviewDismissals      bool
	ReviewDismissalActorIDs        []string
}

func branchProtectionResourceData(d *schema.ResourceData, meta interface{}) (BranchProtectionResourceData, error) {
	data := BranchProtectionResourceData{}

	if v, ok := d.GetOk(REPOSITORY_ID); ok {
		repoID, err := getRepositoryID(v.(string), meta)
		if err != nil {
			return data, err
		}
		data.RepositoryID = repoID.(string)
	}

	if v, ok := d.GetOk(PROTECTION_PATTERN); ok {
		data.Pattern = v.(string)
	}

	if v, ok := d.GetOk(PROTECTION_ALLOWS_DELETIONS); ok {
		data.AllowsDeletions = v.(bool)
	}

	if v, ok := d.GetOk(PROTECTION_ALLOWS_FORCE_PUSHES); ok {
		data.AllowsForcePushes = v.(bool)
	}

	if v, ok := d.GetOk(PROTECTION_IS_ADMIN_ENFORCED); ok {
		data.IsAdminEnforced = v.(bool)
	}

	if v, ok := d.GetOk(PROTECTION_REQUIRES_COMMIT_SIGNATURES); ok {
		data.RequiresCommitSignatures = v.(bool)
	}

	if v, ok := d.GetOk(PROTECTION_REQUIRES_CONVERSATION_RESOLUTION); ok {
		data.RequiresConversationResolution = v.(bool)
	}

	if v, ok := d.GetOk(PROTECTION_REQUIRES_APPROVING_REVIEWS); ok {
		vL := v.([]interface{})
		if len(vL) > 1 {
			return BranchProtectionResourceData{},
				fmt.Errorf("error multiple %s declarations", PROTECTION_REQUIRES_APPROVING_REVIEWS)
		}
		for _, v := range vL {
			if v == nil {
				break
			}

			data.RequiresApprovingReviews = true

			m := v.(map[string]interface{})
			if v, ok := m[PROTECTION_REQUIRED_APPROVING_REVIEW_COUNT]; ok {
				data.RequiredApprovingReviewCount = v.(int)
			}
			if v, ok := m[PROTECTION_DISMISSES_STALE_REVIEWS]; ok {
				data.DismissesStaleReviews = v.(bool)
			}
			if v, ok := m[PROTECTION_REQUIRES_CODE_OWNER_REVIEWS]; ok {
				data.RequiresCodeOwnerReviews = v.(bool)
			}
			if v, ok := m[PROTECTION_RESTRICTS_REVIEW_DISMISSALS]; ok {
				data.RestrictsReviewDismissals = v.(bool)
			}
			if v, ok := m[PROTECTION_RESTRICTS_REVIEW_DISMISSERS]; ok {
				reviewDismissalActorIDs := make([]string, 0)
				vL := v.(*schema.Set).List()
				for _, v := range vL {
					reviewDismissalActorIDs = append(reviewDismissalActorIDs, v.(string))
				}
				if len(reviewDismissalActorIDs) > 0 {
					data.ReviewDismissalActorIDs = reviewDismissalActorIDs
					data.RestrictsReviewDismissals = true
				}
			}
		}
	}

	if v, ok := d.GetOk(PROTECTION_REQUIRES_STATUS_CHECKS); ok {
		vL := v.([]interface{})
		if len(vL) > 1 {
			return BranchProtectionResourceData{},
				fmt.Errorf("error multiple %s declarations", PROTECTION_REQUIRES_STATUS_CHECKS)
		}
		for _, v := range vL {
			if v == nil {
				break
			}

			m := v.(map[string]interface{})
			if v, ok := m[PROTECTION_REQUIRES_STRICT_STATUS_CHECKS]; ok {
				data.RequiresStrictStatusChecks = v.(bool)
			}

			data.RequiredStatusCheckContexts = expandNestedSet(m, PROTECTION_REQUIRED_STATUS_CHECK_CONTEXTS)
			if len(data.RequiredStatusCheckContexts) > 0 {
				data.RequiresStatusChecks = true
			}
		}
	}

	if v, ok := d.GetOk(PROTECTION_RESTRICTS_PUSHES); ok {
		pushActorIDs := make([]string, 0)
		vL := v.(*schema.Set).List()
		for _, v := range vL {
			pushActorIDs = append(pushActorIDs, v.(string))
		}
		if len(pushActorIDs) > 0 {
			data.PushActorIDs = pushActorIDs
			data.RestrictsPushes = true
		}
	}

	return data, nil
}

func setDismissalActorIDs(actors []DismissalActorTypes) []string {
	pushActors := make([]string, 0, len(actors))
	for _, a := range actors {
		if a.Actor.Team != (Actor{}) {
			pushActors = append(pushActors, a.Actor.Team.ID.(string))
		}
		if a.Actor.User != (Actor{}) {
			pushActors = append(pushActors, a.Actor.User.ID.(string))
		}
	}

	return pushActors
}

func setPushActorIDs(actors []PushActorTypes) []string {
	pushActors := make([]string, 0, len(actors))
	for _, a := range actors {
		if a.Actor.Team != (Actor{}) {
			pushActors = append(pushActors, a.Actor.Team.ID.(string))
		}
		if a.Actor.User != (Actor{}) {
			pushActors = append(pushActors, a.Actor.User.ID.(string))
		}
		if a.Actor.App != (Actor{}) {
			pushActors = append(pushActors, a.Actor.App.ID.(string))
		}
	}

	return pushActors
}

func setApprovingReviews(protection BranchProtectionRule) interface{} {
	if !protection.RequiresApprovingReviews {
		return nil
	}

	dismissalAllowances := protection.ReviewDismissalAllowances.Nodes
	dismissalActors := setDismissalActorIDs(dismissalAllowances)
	approvalReviews := []interface{}{
		map[string]interface{}{
			PROTECTION_REQUIRED_APPROVING_REVIEW_COUNT: protection.RequiredApprovingReviewCount,
			PROTECTION_REQUIRES_CODE_OWNER_REVIEWS:     protection.RequiresCodeOwnerReviews,
			PROTECTION_DISMISSES_STALE_REVIEWS:         protection.DismissesStaleReviews,
			PROTECTION_RESTRICTS_REVIEW_DISMISSALS:     protection.RestrictsReviewDismissals,
			PROTECTION_RESTRICTS_REVIEW_DISMISSERS:     dismissalActors,
		},
	}

	return approvalReviews
}

func setStatusChecks(protection BranchProtectionRule) interface{} {
	if !protection.RequiresStatusChecks {
		return nil
	}

	statusChecks := []interface{}{
		map[string]interface{}{
			PROTECTION_REQUIRES_STRICT_STATUS_CHECKS:  protection.RequiresStrictStatusChecks,
			PROTECTION_REQUIRED_STATUS_CHECK_CONTEXTS: protection.RequiredStatusCheckContexts,
		},
	}

	return statusChecks
}

func setPushes(protection BranchProtectionRule) []string {
	if !protection.RestrictsPushes {
		return nil
	}
	pushAllowances := protection.PushAllowances.Nodes
	pushActors := setPushActorIDs(pushAllowances)

	return pushActors
}

func getBranchProtectionID(repoID githubv4.ID, pattern string, meta interface{}) (githubv4.ID, error) {
	var query struct {
		Node struct {
			Repository struct {
				BranchProtectionRules struct {
					Nodes []struct {
						ID      string
						Pattern string
					}
					PageInfo PageInfo
				} `graphql:"branchProtectionRules(first: $first, after: $cursor)"`
				ID string
			} `graphql:"... on Repository"`
		} `graphql:"node(id: $id)"`
	}
	variables := map[string]interface{}{
		"id":     repoID,
		"first":  githubv4.Int(100),
		"cursor": (*githubv4.String)(nil),
	}

	ctx := context.Background()
	client := meta.(*Owner).v4client

	var allRules []struct {
		ID      string
		Pattern string
	}
	for {
		err := client.Query(ctx, &query, variables)
		if err != nil {
			return nil, err
		}

		allRules = append(allRules, query.Node.Repository.BranchProtectionRules.Nodes...)

		if !query.Node.Repository.BranchProtectionRules.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(query.Node.Repository.BranchProtectionRules.PageInfo.EndCursor)
	}

	for i := range allRules {
		if allRules[i].Pattern == pattern {
			return allRules[i].ID, nil
		}
	}

	return nil, fmt.Errorf("Could not find a branch protection rule with the pattern '%s'.", pattern)
}

func statusChecksDiffSuppression(k, old, new string, d *schema.ResourceData) bool {
	data := BranchProtectionResourceData{}
	checks := false

	if v, ok := d.GetOk(PROTECTION_REQUIRES_STATUS_CHECKS); ok {
		vL := v.([]interface{})
		for _, v := range vL {
			if v == nil {
				break
			}

			m := v.(map[string]interface{})
			data.RequiredStatusCheckContexts = expandNestedSet(m, PROTECTION_REQUIRED_STATUS_CHECK_CONTEXTS)
			if len(data.RequiredStatusCheckContexts) > 0 {
				checks = true
			}
		}
	}

	if old == "0" && new == "1" && !checks {
		return true
	}
	return false
}
