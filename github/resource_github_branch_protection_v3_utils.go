package github

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func buildProtectionRequest(d *schema.ResourceData) (*github.ProtectionRequest, error) {
	req := &github.ProtectionRequest{
		EnforceAdmins:                  d.Get("enforce_admins").(bool),
		RequiredConversationResolution: github.Bool(d.Get("require_conversation_resolution").(bool)),
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

		// Contexts and Checks arrays to flatten into
		var contexts []interface{}
		var checks []interface{}

		// TODO: Remove once contexts is fully deprecated.
		// Flatten contexts
		for _, c := range *rsc.Contexts {
			// Parse into contexts
			contexts = append(contexts, c)
		}

		// Flatten checks
		for _, chk := range *rsc.Checks {
			// Parse into checks
			if chk.AppID != nil {
				checks = append(checks, fmt.Sprintf("%s:%d", chk.Context, *chk.AppID))
			} else {
				checks = append(checks, chk.Context)
			}
		}

		return d.Set("required_status_checks", []interface{}{
			map[string]interface{}{
				"strict": rsc.Strict,
				// TODO: Remove once contexts is fully deprecated.
				"contexts": schema.NewSet(schema.HashString, contexts),
				"checks":   schema.NewSet(schema.HashString, checks),
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

	signedCommitStatus, _, err := client.Repositories.GetSignaturesProtectedBranch(ctx,
		orgName, repoName, branch)
	if err != nil {
		log.Printf("[INFO] Not able to read signature protection: %s/%s (%s)", orgName, repoName, branch)
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
		_, _, err = client.Repositories.RequireSignaturesOnProtectedBranch(ctx, orgName, repoName, branch)
	} else {
		_, err = client.Repositories.OptionalSignaturesOnProtectedBranch(ctx, orgName, repoName, branch)
	}
	return err
}

func flattenBypassPullRequestAllowances(bpra *github.BypassPullRequestAllowances) []interface{} {
	if bpra == nil {
		return nil
	}
	users := make([]interface{}, 0, len(bpra.Users))
	for _, u := range bpra.Users {
		if u.Login != nil {
			users = append(users, *u.Login)
		}
	}

	teams := make([]interface{}, 0, len(bpra.Teams))
	for _, t := range bpra.Teams {
		if t.Slug != nil {
			teams = append(teams, *t.Slug)
		}
	}

	apps := make([]interface{}, 0, len(bpra.Apps))
	for _, t := range bpra.Apps {
		if t.Slug != nil {
			apps = append(apps, *t.Slug)
		}
	}

	return []interface{}{
		map[string]interface{}{
			"users": schema.NewSet(schema.HashString, users),
			"teams": schema.NewSet(schema.HashString, teams),
			"apps":  schema.NewSet(schema.HashString, apps),
		},
	}
}

func flattenAndSetRequiredPullRequestReviews(d *schema.ResourceData, protection *github.Protection) error {
	rprr := protection.GetRequiredPullRequestReviews()
	if rprr != nil {
		var users, teams, apps []interface{}
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
			apps = make([]interface{}, 0, len(restrictions.Apps))
			for _, t := range restrictions.Apps {
				if t.Slug != nil {
					apps = append(apps, *t.Slug)
				}
			}
		}

		bpra := flattenBypassPullRequestAllowances(rprr.GetBypassPullRequestAllowances())

		return d.Set("required_pull_request_reviews", []interface{}{
			map[string]interface{}{
				"dismiss_stale_reviews":           rprr.DismissStaleReviews,
				"dismissal_users":                 schema.NewSet(schema.HashString, users),
				"dismissal_teams":                 schema.NewSet(schema.HashString, teams),
				"dismissal_apps":                  schema.NewSet(schema.HashString, apps),
				"require_code_owner_reviews":      rprr.RequireCodeOwnerReviews,
				"require_last_push_approval":      rprr.RequireLastPushApproval,
				"required_approving_review_count": rprr.RequiredApprovingReviewCount,
				"bypass_pull_request_allowances":  bpra,
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

			// Initialise empty literal to ensure an empty array is passed mitigating schema errors like so:
			// For 'anyOf/1', {"strict"=>true, "checks"=>nil} is not a null. []
			rscChecks := []*github.RequiredStatusCheck{}

			// TODO: Remove once contexts is deprecated
			// Iterate and parse contexts into checks using -1 as default to allow checks from all apps.
			contexts := expandNestedSet(m, "contexts")
			for _, c := range contexts {
				appID := int64(-1) // Default
				rscChecks = append(rscChecks, &github.RequiredStatusCheck{
					Context: c,
					AppID:   &appID,
				})
			}

			// Iterate and parse checks
			checks := expandNestedSet(m, "checks")
			for _, c := range checks {

				// Expect a string of "context:app_id", allowing for the absence of "app_id"
				index := strings.LastIndex(c, ":")
				var cContext, cAppId string
				if index <= 0 {
					// If there is no ":" or it's in the first position, there is no app_id.
					cContext, cAppId = c, ""
				} else {
					cContext, cAppId = c[:index], c[index+1:]
				}

				var rscCheck *github.RequiredStatusCheck
				if cAppId != "" {
					// If we have a valid app_id, include it in the RSC
					rscAppId, err := strconv.Atoi(cAppId)
					if err != nil {
						return nil, fmt.Errorf("could not parse %v as valid app_id", cAppId)
					}
					rscAppId64 := int64(rscAppId)
					rscCheck = &github.RequiredStatusCheck{Context: cContext, AppID: &rscAppId64}
				} else {
					// Else simply provide the context
					rscCheck = &github.RequiredStatusCheck{Context: cContext}
				}

				// Append
				rscChecks = append(rscChecks, rscCheck)
			}
			// Assign after looping both checks and contexts
			rsc.Checks = &rscChecks
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

			apps := expandNestedSet(m, "dismissal_apps")
			if len(apps) > 0 {
				drr.Apps = &apps
			}

			bpra, err := expandBypassPullRequestAllowances(m)
			if err != nil {
				return nil, err
			}

			rprr.DismissalRestrictionsRequest = drr
			rprr.DismissStaleReviews = m["dismiss_stale_reviews"].(bool)
			rprr.RequireCodeOwnerReviews = m["require_code_owner_reviews"].(bool)
			rprr.RequiredApprovingReviewCount = m["required_approving_review_count"].(int)
			requireLastPushApproval := m["require_last_push_approval"].(bool)
			rprr.RequireLastPushApproval = &requireLastPushApproval
			rprr.BypassPullRequestAllowancesRequest = bpra
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

func expandBypassPullRequestAllowances(m map[string]interface{}) (*github.BypassPullRequestAllowancesRequest, error) {
	if m["bypass_pull_request_allowances"] == nil {
		return nil, nil
	}

	vL := m["bypass_pull_request_allowances"].([]interface{})
	if len(vL) > 1 {
		return nil, errors.New("cannot specify bypass_pull_request_allowances more than one time")
	}

	var bpra *github.BypassPullRequestAllowancesRequest

	for _, v := range vL {
		if v == nil {
			return nil, errors.New("invalid bypass_pull_request_allowances")
		}
		bpra = new(github.BypassPullRequestAllowancesRequest)
		m := v.(map[string]interface{})

		users := expandNestedSet(m, "users")
		bpra.Users = users
		teams := expandNestedSet(m, "teams")
		bpra.Teams = teams
		apps := expandNestedSet(m, "apps")
		bpra.Apps = apps
	}

	return bpra, nil
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
