package github

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/google/go-github/v39/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func buildProtectionRequest(d *schema.ResourceData) (*github.ProtectionRequest, error) {
	req := &github.ProtectionRequest{
		EnforceAdmins: d.Get("enforce_admins").(bool),
	}

	rsc, err := expandRequiredStatusChecks(d)
	if err != nil {
		return nil, err
	}
	req.RequiredStatusChecks = rsc

	rprr, err := expandRequiredPullRequestReviews(d)
	if err != nil {
		return nil, err
	}
	req.RequiredPullRequestReviews = rprr

	res, err := expandRestrictions(d)
	if err != nil {
		return nil, err
	}
	req.Restrictions = res

	return req, nil
}

func flattenAndSetRequiredStatusChecks(d *schema.ResourceData, protection *github.Protection) error {
	rsc := protection.GetRequiredStatusChecks()
	if rsc != nil {
		contexts := make([]interface{}, 0, len(rsc.Contexts))
		for _, c := range rsc.Contexts {
			contexts = append(contexts, c)
		}

		return d.Set("required_status_checks", []interface{}{
			map[string]interface{}{
				"strict":   rsc.Strict,
				"contexts": schema.NewSet(schema.HashString, contexts),
			},
		})
	}

	return d.Set("required_status_checks", []interface{}{})
}

func requireSignedCommitsRead(d *schema.ResourceData, meta interface{}) error {
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

	log.Printf("[DEBUG] Reading branch protection signed commit status: %s/%s (%s)", orgName, repoName, branch)
	signedCommitStatus, _, err := client.Repositories.GetSignaturesProtectedBranch(ctx,
		orgName, repoName, branch)
	if err != nil {
		log.Printf("[WARN] Not able to read signature protection: %s/%s (%s)", orgName, repoName, branch)
		return nil
	}

	return d.Set("require_signed_commits", signedCommitStatus.Enabled)
}

func requireSignedCommitsUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	requiredSignedCommit := d.Get("require_signed_commits").(bool)
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

	if requiredSignedCommit {
		log.Printf("[DEBUG] Enabling branch protection signed commit: %s/%s (%s) - $s", orgName, repoName, branch)
		_, _, err = client.Repositories.RequireSignaturesOnProtectedBranch(ctx, orgName, repoName, branch)
	} else {
		log.Printf("[DEBUG] Removing branch protection signed commit: %s/%s (%s) - $s", orgName, repoName, branch)
		_, err = client.Repositories.OptionalSignaturesOnProtectedBranch(ctx, orgName, repoName, branch)
	}
	return err
}

func flattenAndSetRequiredPullRequestReviews(d *schema.ResourceData, protection *github.Protection) error {
	rprr := protection.GetRequiredPullRequestReviews()
	if rprr != nil {
		var users, teams []interface{}
		restrictions := rprr.GetDismissalRestrictions()

		if restrictions != nil {
			users = make([]interface{}, 0, len(restrictions.Users))
			for _, u := range restrictions.Users {
				if u.Login != nil {
					users = append(users, *u.Login)
				}
			}
			teams = make([]interface{}, 0, len(restrictions.Teams))
			for _, t := range restrictions.Teams {
				if t.Slug != nil {
					teams = append(teams, *t.Slug)
				}
			}
		}

		return d.Set("required_pull_request_reviews", []interface{}{
			map[string]interface{}{
				"dismiss_stale_reviews":           rprr.DismissStaleReviews,
				"dismissal_users":                 schema.NewSet(schema.HashString, users),
				"dismissal_teams":                 schema.NewSet(schema.HashString, teams),
				"require_code_owner_reviews":      rprr.RequireCodeOwnerReviews,
				"required_approving_review_count": rprr.RequiredApprovingReviewCount,
			},
		})
	}

	return d.Set("required_pull_request_reviews", []interface{}{})
}

func flattenAndSetRestrictions(d *schema.ResourceData, protection *github.Protection) error {
	restrictions := protection.GetRestrictions()
	if restrictions != nil {
		users := make([]interface{}, 0, len(restrictions.Users))
		for _, u := range restrictions.Users {
			if u.Login != nil {
				users = append(users, *u.Login)
			}
		}

		teams := make([]interface{}, 0, len(restrictions.Teams))
		for _, t := range restrictions.Teams {
			if t.Slug != nil {
				teams = append(teams, *t.Slug)
			}
		}

		apps := make([]interface{}, 0, len(restrictions.Apps))
		for _, t := range restrictions.Apps {
			if t.Slug != nil {
				apps = append(apps, *t.Slug)
			}
		}

		return d.Set("restrictions", []interface{}{
			map[string]interface{}{
				"users": schema.NewSet(schema.HashString, users),
				"teams": schema.NewSet(schema.HashString, teams),
				"apps":  schema.NewSet(schema.HashString, apps),
			},
		})
	}

	return d.Set("restrictions", []interface{}{})
}

func expandRequiredStatusChecks(d *schema.ResourceData) (*github.RequiredStatusChecks, error) {
	if v, ok := d.GetOk("required_status_checks"); ok {
		vL := v.([]interface{})
		if len(vL) > 1 {
			return nil, errors.New("cannot specify required_status_checks more than one time")
		}
		rsc := new(github.RequiredStatusChecks)

		for _, v := range vL {
			// List can only have one item, safe to early return here
			if v == nil {
				return nil, nil
			}
			m := v.(map[string]interface{})
			rsc.Strict = m["strict"].(bool)

			contexts := expandNestedSet(m, "contexts")
			rsc.Contexts = contexts
		}
		return rsc, nil
	}

	return nil, nil
}

func expandRequiredPullRequestReviews(d *schema.ResourceData) (*github.PullRequestReviewsEnforcementRequest, error) {
	if v, ok := d.GetOk("required_pull_request_reviews"); ok {
		vL := v.([]interface{})
		if len(vL) > 1 {
			return nil, errors.New("cannot specify required_pull_request_reviews more than one time")
		}

		rprr := new(github.PullRequestReviewsEnforcementRequest)
		drr := new(github.DismissalRestrictionsRequest)

		for _, v := range vL {
			// List can only have one item, safe to early return here
			if v == nil {
				return nil, nil
			}
			m := v.(map[string]interface{})

			users := expandNestedSet(m, "dismissal_users")
			if len(users) > 0 {
				drr.Users = &users
			}
			teams := expandNestedSet(m, "dismissal_teams")
			if len(teams) > 0 {
				drr.Teams = &teams
			}

			rprr.DismissalRestrictionsRequest = drr
			rprr.DismissStaleReviews = m["dismiss_stale_reviews"].(bool)
			rprr.RequireCodeOwnerReviews = m["require_code_owner_reviews"].(bool)
			rprr.RequiredApprovingReviewCount = m["required_approving_review_count"].(int)
		}

		return rprr, nil
	}

	return nil, nil
}

func expandRestrictions(d *schema.ResourceData) (*github.BranchRestrictionsRequest, error) {
	if v, ok := d.GetOk("restrictions"); ok {
		vL := v.([]interface{})
		if len(vL) > 1 {
			return nil, errors.New("cannot specify restrictions more than one time")
		}
		restrictions := new(github.BranchRestrictionsRequest)

		for _, v := range vL {
			// Restrictions only have set attributes nested, need to return nil values for these.
			// The API won't initialize these as nil
			if v == nil {
				restrictions.Users = []string{}
				restrictions.Teams = []string{}
				restrictions.Apps = []string{}
				return restrictions, nil
			}
			m := v.(map[string]interface{})

			users := expandNestedSet(m, "users")
			restrictions.Users = users
			teams := expandNestedSet(m, "teams")
			restrictions.Teams = teams
			apps := expandNestedSet(m, "apps")
			restrictions.Apps = apps
		}
		return restrictions, nil
	}

	return nil, nil
}

func checkBranchRestrictionsUsers(actual *github.BranchRestrictions, expected *github.BranchRestrictionsRequest) error {
	if expected == nil {
		return nil
	}

	expectedUsers := expected.Users

	if actual == nil {
		return fmt.Errorf("unable to add users in restrictions: %s", strings.Join(expectedUsers, ", "))
	}

	actualLoopUp := make(map[string]struct{}, len(actual.Users))
	for _, a := range actual.Users {
		actualLoopUp[a.GetLogin()] = struct{}{}
	}

	notFounds := make([]string, 0, len(actual.Users))

	for _, e := range expectedUsers {
		if _, ok := actualLoopUp[e]; !ok {
			notFounds = append(notFounds, e)
		}
	}

	if len(notFounds) == 0 {
		return nil
	}

	return fmt.Errorf("unable to add users in restrictions: %s", strings.Join(notFounds, ", "))
}
