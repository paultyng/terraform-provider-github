package github

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/google/go-github/v39/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceGithubRepositoryEnvironment() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubRepositoryEnvironmentCreate,
		Read:   resourceGithubRepositoryEnvironmentRead,
		Update: resourceGithubRepositoryEnvironmentUpdate,
		Delete: resourceGithubRepositoryEnvironmentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"repository": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"environment": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"wait_timer": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 43200),
			},
			"reviewers": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 6,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"teams": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeInt},
						},
						"users": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeInt},
						},
					},
				},
			},
			"deployment_branch_policy": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"protected_branches": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"custom_branch_policies": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceGithubRepositoryEnvironmentCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	owner := meta.(*Owner).name
	repoName := d.Get("repository").(string)
	envName := d.Get("environment").(string)
	updateData := createUpdateEnvironmentData(d, meta)

	ctx := context.Background()

	log.Printf("[DEBUG] Creating repository environment: %s/%s/%s", owner, repoName, envName)
	_, _, err := client.Repositories.CreateUpdateEnvironment(ctx, owner, repoName, envName, &updateData)

	if err != nil {
		return err
	}

	d.SetId(buildTwoPartID(repoName, envName))

	return resourceGithubRepositoryEnvironmentRead(d, meta)
}

func resourceGithubRepositoryEnvironmentRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	owner := meta.(*Owner).name
	repoName, envName, err := parseTwoPartID(d.Id(), "repository", "environment")
	if err != nil {
		return err
	}

	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	log.Printf("[DEBUG] Reading repository environment: %s (%s/%s/%s)", d.Id(), owner, repoName, envName)
	env, _, err := client.Repositories.GetEnvironment(ctx, owner, repoName, envName)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[WARN] Removing repository environment %s from state because it no longer exists in GitHub",
					d.Id())
				d.SetId("")
				return nil
			}
		}
	}

	d.Set("repository", repoName)
	d.Set("environment", envName)

	for _, pr := range env.ProtectionRules {
		switch *pr.Type {
		case "wait_timer":
			d.Set("wait_timer", pr.WaitTimer)

		case "required_reviewers":
			teams := make([]int64, 0)
			users := make([]int64, 0)

			for _, r := range pr.Reviewers {
				switch *r.Type {
				case "Team":
					teams = append(teams, *r.Reviewer.(*github.Team).ID)
				case "User":
					users = append(users, *r.Reviewer.(*github.User).ID)
				}
			}
			d.Set("reviewers", []interface{}{
				map[string]interface{}{
					"teams": teams,
					"users": users,
				},
			})
		}
	}

	if env.DeploymentBranchPolicy != nil {
		d.Set("deployment_branch_policy", []interface{}{
			map[string]interface{}{
				"protected_branches":     env.DeploymentBranchPolicy.ProtectedBranches,
				"custom_branch_policies": env.DeploymentBranchPolicy.CustomBranchPolicies,
			},
		})
	}

	return nil
}

func resourceGithubRepositoryEnvironmentUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	owner := meta.(*Owner).name
	repoName := d.Get("repository").(string)
	envName := d.Get("environment").(string)
	updateData := createUpdateEnvironmentData(d, meta)

	ctx := context.Background()

	log.Printf("[DEBUG] Updating repository environment: %s/%s/%s", owner, repoName, envName)
	resultKey, _, err := client.Repositories.CreateUpdateEnvironment(ctx, owner, repoName, envName, &updateData)

	if err != nil {
		return err
	}

	d.SetId(buildTwoPartID(repoName, strconv.FormatInt(resultKey.GetID(), 10)))

	return resourceGithubRepositoryEnvironmentRead(d, meta)
}

func resourceGithubRepositoryEnvironmentDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	owner := meta.(*Owner).name
	repoName, envName, err := parseTwoPartID(d.Id(), "repository", "environment")
	if err != nil {
		return err
	}

	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	log.Printf("[DEBUG] Deleting repository environment: %s/%s/%s", owner, repoName, envName)
	_, err = client.Repositories.DeleteEnvironment(ctx, owner, repoName, envName)
	return err
}

func createUpdateEnvironmentData(d *schema.ResourceData, meta interface{}) github.CreateUpdateEnvironment {
	data := github.CreateUpdateEnvironment{}

	if v, ok := d.GetOk("wait_timer"); ok {
		data.WaitTimer = github.Int(v.(int))
	}

	if v, ok := d.GetOk("reviewers"); ok {
		envReviewers := make([]*github.EnvReviewers, 0)

		for _, team := range expandReviewers(v, "teams") {
			envReviewers = append(envReviewers, &github.EnvReviewers{
				Type: github.String("Team"),
				ID:   github.Int64(team),
			})
		}

		for _, user := range expandReviewers(v, "users") {
			envReviewers = append(envReviewers, &github.EnvReviewers{
				Type: github.String("User"),
				ID:   github.Int64(user),
			})
		}

		data.Reviewers = envReviewers
	}

	if v, ok := d.GetOk("deployment_branch_policy"); ok {
		policy := v.([]interface{})[0].(map[string]interface{})
		data.DeploymentBranchPolicy = &github.BranchPolicy{
			ProtectedBranches:    github.Bool(policy["protected_branches"].(bool)),
			CustomBranchPolicies: github.Bool(policy["custom_branch_policies"].(bool)),
		}
	}

	return data
}

func expandReviewers(v interface{}, target string) []int64 {
	res := make([]int64, 0)
	m := v.([]interface{})[0]
	if m != nil {
		if v, ok := m.(map[string]interface{})[target]; ok {
			vL := v.(*schema.Set).List()
			for _, v := range vL {
				res = append(res, int64(v.(int)))
			}
		}
	}
	return res
}
