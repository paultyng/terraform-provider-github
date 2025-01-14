package github

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/google/go-github/v66/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceGithubOrganizationRuleset() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubOrganizationRulesetCreate,
		Read:   resourceGithubOrganizationRulesetRead,
		Update: resourceGithubOrganizationRulesetUpdate,
		Delete: resourceGithubOrganizationRulesetDelete,
		Importer: &schema.ResourceImporter{
			State: resourceGithubOrganizationRulesetImport,
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
				ValidateFunc: validation.StringInSlice([]string{"branch", "tag", "push"}, false),
				Description:  "Possible values are `branch`, `tag` and `push`. Note: The `push` target is in beta and is subject to change.",
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
					},
				},
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
				Description: "Parameters for an organization ruleset condition. `ref_name` is required alongside one of `repository_name` or `repository_id`.",
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
						"repository_name": {
							Type:         schema.TypeList,
							Optional:     true,
							MaxItems:     1,
							ExactlyOneOf: []string{"conditions.0.repository_id"},
							AtLeastOneOf: []string{"conditions.0.repository_id"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"include": {
										Type:        schema.TypeList,
										Required:    true,
										Description: "Array of repository names or patterns to include. One of these patterns must match for the condition to pass. Also accepts `~ALL` to include all repositories.",
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"exclude": {
										Type:        schema.TypeList,
										Required:    true,
										Description: "Array of repository names or patterns to exclude. The condition will not pass if any of these patterns match.",
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"protected": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										Description: "Whether renaming of target repositories is prevented.",
									},
								},
							},
						},
						"repository_id": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "The repository IDs that the ruleset applies to. One of these IDs must match for the condition to pass.",
							Elem: &schema.Schema{
								Type: schema.TypeInt,
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
							Description: "Parameters to be used for the commit_message_pattern rule.",
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
							Description: "Parameters to be used for the commit_author_email_pattern rule.",
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
							Description: "Parameters to be used for the committer_email_pattern rule.",
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
						"required_workflows": {
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Description: "Choose which Actions workflows must pass before branches can be merged into a branch that matches this rule.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"do_not_enforce_on_create": {
										Type:        schema.TypeBool,
										Optional:    true,
										Description: "Allow repositories and branches to be created if a check would otherwise prohibit it.",
									},
									"required_workflow": {
										Type:        schema.TypeSet,
										MinItems:    1,
										Required:    true,
										Description: "Actions workflows that are required. Several can be defined.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"repository_id": {
													Type:        schema.TypeInt,
													Required:    true,
													Description: "The repository in which the workflow is defined.",
												},
												"path": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "The path to the workflow YAML definition file.",
												},
												"ref": {
													Type:        schema.TypeString,
													Optional:    true,
													Default:     "master",
													Description: "The ref (branch or tag) of the workflow file to use.",
												},
											},
										},
									},
								},
							},
						},
						"required_code_scanning": {
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Description: "Choose which tools must provide code scanning results before the reference is updated. When configured, code scanning must be enabled and have results for both the commit and the reference being updated.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"required_code_scanning_tool": {
										Type:        schema.TypeSet,
										MinItems:    1,
										Required:    true,
										Description: "Tools that must provide code scanning results for this rule to pass.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"alerts_threshold": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "The severity level at which code scanning results that raise alerts block a reference update. Can be one of: `none`, `errors`, `errors_and_warnings`, `all`.",
												},
												"security_alerts_threshold": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "The severity level at which code scanning results that raise security alerts block a reference update. Can be one of: `none`, `critical`, `high_or_higher`, `medium_or_higher`, `all`.",
												},
												"tool": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "The name of a code scanning tool.",
												},
											},
										},
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

func resourceGithubOrganizationRulesetCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	owner := meta.(*Owner).name

	rulesetReq := resourceGithubRulesetObject(d, owner)

	ctx := context.Background()

	var ruleset *github.Ruleset
	var err error

	ruleset, _, err = client.Organizations.CreateOrganizationRuleset(ctx, owner, rulesetReq)
	if err != nil {
		return err
	}
	d.SetId(strconv.FormatInt(*ruleset.ID, 10))
	return resourceGithubOrganizationRulesetRead(d, meta)
}

func resourceGithubOrganizationRulesetRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	owner := meta.(*Owner).name

	rulesetID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return unconvertibleIdErr(d.Id(), err)
	}

	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	if !d.IsNewResource() {
		ctx = context.WithValue(ctx, ctxEtag, d.Get("etag").(string))
	}

	var ruleset *github.Ruleset
	var resp *github.Response

	ruleset, resp, err = client.Organizations.GetOrganizationRuleset(ctx, owner, rulesetID)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotModified {
				return nil
			}
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[INFO] Removing ruleset %s: %d from state because it no longer exists in GitHub",
					owner, rulesetID)
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
	d.Set("conditions", flattenConditions(ruleset.GetConditions(), true))
	d.Set("rules", flattenRules(ruleset.Rules, true))
	d.Set("node_id", ruleset.GetNodeID())
	d.Set("ruleset_id", ruleset.ID)

	return nil
}

func resourceGithubOrganizationRulesetUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	owner := meta.(*Owner).name

	rulesetReq := resourceGithubRulesetObject(d, owner)

	rulesetID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return unconvertibleIdErr(d.Id(), err)
	}

	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	ruleset, _, err := client.Organizations.UpdateOrganizationRuleset(ctx, owner, rulesetID, rulesetReq)
	if err != nil {
		return err
	}
	d.SetId(strconv.FormatInt(*ruleset.ID, 10))

	return resourceGithubOrganizationRulesetRead(d, meta)
}

func resourceGithubOrganizationRulesetDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name

	rulesetID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return unconvertibleIdErr(d.Id(), err)
	}
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	log.Printf("[DEBUG] Deleting organization ruleset: %s: %d", owner, rulesetID)
	_, err = client.Organizations.DeleteOrganizationRuleset(ctx, owner, rulesetID)
	return err
}

func resourceGithubOrganizationRulesetImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	rulesetID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return []*schema.ResourceData{d}, unconvertibleIdErr(d.Id(), err)
	}
	if rulesetID == 0 {
		return []*schema.ResourceData{d}, fmt.Errorf("`ruleset_id` must be present")
	}
	log.Printf("[DEBUG] Importing organization ruleset with ID: %d", rulesetID)

	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.Background()

	ruleset, _, err := client.Organizations.GetOrganizationRuleset(ctx, owner, rulesetID)
	if ruleset == nil || err != nil {
		return []*schema.ResourceData{d}, err
	}
	d.SetId(strconv.FormatInt(ruleset.GetID(), 10))

	return []*schema.ResourceData{d}, nil
}
