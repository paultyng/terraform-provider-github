package github

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceGithubRepositoryRuleset() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubRepositoryRulesetCreate,
		Read:   resourceGithubRepositoryRulesetRead,
		Update: resourceGithubRepositoryRulesetUpdate,
		Delete: resourceGithubRepositoryRulesetDelete,
		Importer: &schema.ResourceImporter{
			State: resourceGithubRepositoryRulesetImport,
		},

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
				Description:  "The name of the ruleset.",
			},
			"target": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"branch", "tag"}, false),
				Description:  "Possible values are `branch` and `tag`.",
			},
			"repository": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the repository to apply rulset to.",
			},
			"enforcement": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"disabled", "active", "evaluate"}, false),
				Description:  "Possible values for Enforcement are `disabled`, `active`, `evaluate`. Note: `evaluate` is currently only supported for owners of type `organization`.",
			},
			"bypass_actors": {
				Type:             schema.TypeList,
				Optional:         true,
				DiffSuppressFunc: bypassActorsDiffSuppressFunc,
				Description:      "The actors that can bypass the rules in this ruleset.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"actor_id": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The ID of the actor that can bypass a ruleset. When `actor_type` is `OrganizationAdmin`, this should be set to `1`.",
						},
						"actor_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"RepositoryRole", "Team", "Integration", "OrganizationAdmin"}, false),
							Description:  "The type of actor that can bypass a ruleset. Can be one of: `RepositoryRole`, `Team`, `Integration`, `OrganizationAdmin`.",
						},
						"bypass_mode": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"always", "pull_request"}, false),
							Description:  "When the specified actor can bypass the ruleset. pull_request means that an actor can only bypass rules on pull requests. Can be one of: `always`, `pull_request`.",
						},
					}},
			},
			"node_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "GraphQL global node id for use with v4 API.",
			},
			"ruleset_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "GitHub ID for the ruleset.",
			},
			"conditions": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Parameters for a repository ruleset ref name condition.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ref_name": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"include": {
										Type:        schema.TypeList,
										Required:    true,
										Description: "Array of ref names or patterns to include. One of these patterns must match for the condition to pass. Also accepts `~DEFAULT_BRANCH` to include the default branch or `~ALL` to include all branches.",
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"exclude": {
										Type:        schema.TypeList,
										Required:    true,
										Description: "Array of ref names or patterns to exclude. The condition will not pass if any of these patterns match.",
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
						},
					},
				},
			},
			"rules": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "Rules within the ruleset.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"creation": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Only allow users with bypass permission to create matching refs.",
						},
						"update": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Only allow users with bypass permission to update matching refs.",
						},
						"update_allows_fetch_and_merge": {
							Type:         schema.TypeBool,
							Optional:     true,
							Default:      false,
							RequiredWith: []string{"rules.0.update"},
							Description:  "Branch can pull changes from its upstream repository. This is only applicable to forked repositories. Requires `update` to be set to `true`.",
						},
						"deletion": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Only allow users with bypass permissions to delete matching refs.",
						},
						"required_linear_history": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Prevent merge commits from being pushed to matching branches.",
						},
						"required_deployments": {
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Description: "Choose which environments must be successfully deployed to before branches can be merged into a branch that matches this rule.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"required_deployment_environments": {
										Type:        schema.TypeList,
										Required:    true,
										Description: "The environments that must be successfully deployed to before branches can be merged.",
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
						},
						"required_signatures": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Commits pushed to matching branches must have verified signatures.",
						},
						"pull_request": {
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Description: "Require all commits be made to a non-target branch and submitted via a pull request before they can be merged.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"dismiss_stale_reviews_on_push": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										Description: "New, reviewable commits pushed will dismiss previous pull request review approvals. Defaults to `false`.",
									},
									"require_code_owner_review": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										Description: "Require an approving review in pull requests that modify files that have a designated code owner. Defaults to `false`.",
									},
									"require_last_push_approval": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										Description: "Whether the most recent reviewable push must be approved by someone other than the person who pushed it. Defaults to `false`.",
									},
									"required_approving_review_count": {
										Type:        schema.TypeInt,
										Optional:    true,
										Default:     0,
										Description: "The number of approving reviews that are required before a pull request can be merged. Defaults to `0`.",
									},
									"required_review_thread_resolution": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										Description: "All conversations on code must be resolved before a pull request can be merged. Defaults to `false`.",
									},
								},
							},
						},
						"required_status_checks": {
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Description: "Choose which status checks must pass before branches can be merged into a branch that matches this rule. When enabled, commits must first be pushed to another branch, then merged or pushed directly to a branch that matches this rule after status checks have passed.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"required_check": {
										Type:        schema.TypeSet,
										MinItems:    1,
										Required:    true,
										Description: "Status checks that are required. Several can be defined.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"context": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "The status check context name that must be present on the commit.",
												},
												"integration_id": {
													Type:        schema.TypeInt,
													Optional:    true,
													Default:     0,
													Description: "The optional integration ID that this status check must originate from.",
												},
											},
										},
									},
									"strict_required_status_checks_policy": {
										Type:        schema.TypeBool,
										Optional:    true,
										Description: "Whether pull requests targeting a matching branch must be tested with the latest code. This setting will not take effect unless at least one status check is enabled. Defaults to `false`.",
									},
								},
							},
						},
						"non_fast_forward": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Prevent users with push access from force pushing to branches.",
						},
						"commit_message_pattern": {
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Description: "Parameters to be used for the commit_message_pattern rule. This rule only applies to repositories within an enterprise, it cannot be applied to repositories owned by individuals or regular organizations.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "How this rule will appear to users.",
									},
									"negate": {
										Type:        schema.TypeBool,
										Optional:    true,
										Description: "If true, the rule will fail if the pattern matches.",
									},
									"operator": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The operator to use for matching. Can be one of: `starts_with`, `ends_with`, `contains`, `regex`.",
									},
									"pattern": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The pattern to match with.",
									},
								},
							},
						},
						"commit_author_email_pattern": {
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Description: "Parameters to be used for the commit_author_email_pattern rule. This rule only applies to repositories within an enterprise, it cannot be applied to repositories owned by individuals or regular organizations.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "How this rule will appear to users.",
									},
									"negate": {
										Type:        schema.TypeBool,
										Optional:    true,
										Description: "If true, the rule will fail if the pattern matches.",
									},
									"operator": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The operator to use for matching. Can be one of: `starts_with`, `ends_with`, `contains`, `regex`.",
									},
									"pattern": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The pattern to match with.",
									},
								},
							},
						},
						"committer_email_pattern": {
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Description: "Parameters to be used for the committer_email_pattern rule. This rule only applies to repositories within an enterprise, it cannot be applied to repositories owned by individuals or regular organizations.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "How this rule will appear to users.",
									},
									"negate": {
										Type:        schema.TypeBool,
										Optional:    true,
										Description: "If true, the rule will fail if the pattern matches.",
									},
									"operator": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The operator to use for matching. Can be one of: `starts_with`, `ends_with`, `contains`, `regex`.",
									},
									"pattern": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The pattern to match with.",
									},
								},
							},
						},
						"branch_name_pattern": {
							Type:          schema.TypeList,
							MaxItems:      1,
							Optional:      true,
							ConflictsWith: []string{"rules.0.tag_name_pattern"},
							Description:   "Parameters to be used for the branch_name_pattern rule. This rule only applies to repositories within an enterprise, it cannot be applied to repositories owned by individuals or regular organizations. Conflicts with `tag_name_pattern` as it only applies to rulesets with target `branch`.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "How this rule will appear to users.",
									},
									"negate": {
										Type:        schema.TypeBool,
										Optional:    true,
										Description: "If true, the rule will fail if the pattern matches.",
									},
									"operator": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The operator to use for matching. Can be one of: `starts_with`, `ends_with`, `contains`, `regex`.",
									},
									"pattern": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The pattern to match with.",
									},
								},
							},
						},
						"tag_name_pattern": {
							Type:          schema.TypeList,
							MaxItems:      1,
							Optional:      true,
							ConflictsWith: []string{"rules.0.branch_name_pattern"},
							Description:   "Parameters to be used for the tag_name_pattern rule. This rule only applies to repositories within an enterprise, it cannot be applied to repositories owned by individuals or regular organizations. Conflicts with `branch_name_pattern` as it only applies to rulesets with target `tag`.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "How this rule will appear to users.",
									},
									"negate": {
										Type:        schema.TypeBool,
										Optional:    true,
										Description: "If true, the rule will fail if the pattern matches.",
									},
									"operator": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The operator to use for matching. Can be one of: `starts_with`, `ends_with`, `contains`, `regex`.",
									},
									"pattern": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The pattern to match with.",
									},
								},
							},
						},
					},
				},
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceGithubRepositoryRulesetCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	rulesetReq := resourceGithubRulesetObject(d, "")

	owner := meta.(*Owner).name

	repoName := d.Get("repository").(string)
	ctx := context.Background()

	var ruleset *github.Ruleset
	var err error

	ruleset, _, err = client.Repositories.CreateRuleset(ctx, owner, repoName, rulesetReq)
	if err != nil {
		return err
	}
	d.SetId(strconv.FormatInt(*ruleset.ID, 10))

	return resourceGithubRepositoryRulesetRead(d, meta)
}

func resourceGithubRepositoryRulesetRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	owner := meta.(*Owner).name

	repoName := d.Get("repository").(string)
	rulesetID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return unconvertibleIdErr(d.Id(), err)
	}

	ctx := context.WithValue(context.Background(), ctxId, rulesetID)
	if !d.IsNewResource() {
		ctx = context.WithValue(ctx, ctxEtag, d.Get("etag").(string))
	}

	var ruleset *github.Ruleset
	var resp *github.Response

	ruleset, resp, err = client.Repositories.GetRuleset(ctx, owner, repoName, rulesetID, false)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotModified {
				return nil
			}
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[INFO] Removing ruleset %s/%s: %d from state because it no longer exists in GitHub",
					owner, repoName, rulesetID)
				d.SetId("")
				return nil
			}
		}
	}

	d.Set("etag", resp.Header.Get("ETag"))
	d.Set("name", ruleset.Name)
	d.Set("target", ruleset.GetTarget())
	d.Set("enforcement", ruleset.Enforcement)
	d.Set("bypass_actors", flattenBypassActors(ruleset.BypassActors))
	d.Set("conditions", flattenConditions(ruleset.GetConditions(), false))
	d.Set("rules", flattenRules(ruleset.Rules, false))
	d.Set("node_id", ruleset.GetNodeID())
	d.Set("ruleset_id", ruleset.ID)

	return nil
}

func resourceGithubRepositoryRulesetUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	rulesetReq := resourceGithubRulesetObject(d, "")

	owner := meta.(*Owner).name

	repoName := d.Get("repository").(string)
	rulesetID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return unconvertibleIdErr(d.Id(), err)
	}

	ctx := context.WithValue(context.Background(), ctxId, rulesetID)

	ruleset, _, err := client.Repositories.UpdateRuleset(ctx, owner, repoName, rulesetID, rulesetReq)
	if err != nil {
		return err
	}
	d.SetId(strconv.FormatInt(*ruleset.ID, 10))

	return resourceGithubRepositoryRulesetRead(d, meta)
}

func resourceGithubRepositoryRulesetDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name

	repoName := d.Get("repository").(string)
	rulesetID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return unconvertibleIdErr(d.Id(), err)
	}
	ctx := context.WithValue(context.Background(), ctxId, rulesetID)

	log.Printf("[DEBUG] Deleting repository ruleset: %s/%s: %d", owner, repoName, rulesetID)
	_, err = client.Repositories.DeleteRuleset(ctx, owner, repoName, rulesetID)
	return err
}

func resourceGithubRepositoryRulesetImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	repoName, rulesetIDStr, err := parseTwoPartID(d.Id(), "repository", "ruleset")
	if err != nil {
		return []*schema.ResourceData{d}, err
	}

	rulesetID, err := strconv.ParseInt(rulesetIDStr, 10, 64)
	if err != nil {
		return []*schema.ResourceData{d}, unconvertibleIdErr(rulesetIDStr, err)
	}
	if rulesetID == 0 {
		return []*schema.ResourceData{d}, fmt.Errorf("`ruleset_id` must be present")
	}
	log.Printf("[DEBUG] Importing repository ruleset with ID: %d, for repository: %s", rulesetID, repoName)

	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.Background()
	repository, _, err := client.Repositories.Get(ctx, owner, repoName)
	if repository == nil || err != nil {
		return []*schema.ResourceData{d}, err
	}
	d.Set("repository", *repository.Name)

	ruleset, _, err := client.Repositories.GetRuleset(ctx, owner, *repository.Name, rulesetID, false)
	if ruleset == nil || err != nil {
		return []*schema.ResourceData{d}, err
	}
	d.SetId(strconv.FormatInt(ruleset.GetID(), 10))

	return []*schema.ResourceData{d}, nil
}
