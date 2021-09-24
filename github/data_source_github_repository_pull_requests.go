package github

import (
	"context"
	"strings"

	"github.com/google/go-github/v39/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

// Docs: https://docs.github.com/en/rest/reference/pulls#list-pull-requests
func dataSourceGithubRepositoryPullRequests() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGithubRepositoryPullRequestsRead,
		Schema: map[string]*schema.Schema{
			"owner": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"base_repository": {
				Type:     schema.TypeString,
				Required: true,
			},
			"base_ref": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"head_ref": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"sort_by": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "created",
				ValidateFunc: validation.StringInSlice([]string{"created", "updated", "popularity", "long-running"}, false),
			},
			"sort_direction": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "asc",
				ValidateFunc: validation.StringInSlice([]string{"asc", "desc"}, false),
			},
			"state": {
				Type:         schema.TypeString,
				Default:      "open",
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"open", "closed", "all"}, false),
			},
			"results": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"number": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Per-repository, monotonically increasing ID of this PR",
						},
						"base_ref": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"base_sha": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"body": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"draft": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"head_owner": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"head_ref": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"head_repository": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"head_sha": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"labels": {
							Type:        schema.TypeList,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
							Description: "List of names of labels on the PR",
						},
						"maintainer_can_modify": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"opened_at": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"opened_by": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Username of the PR creator",
						},
						"state": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"title": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"updated_at": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceGithubRepositoryPullRequestsRead(d *schema.ResourceData, meta interface{}) error {
	ctx := context.TODO()
	client := meta.(*Owner).v3client

	owner := meta.(*Owner).name
	if explicitOwner, ok := d.GetOk("owner"); ok {
		owner = explicitOwner.(string)
	}

	baseRepository := d.Get("base_repository").(string)
	state := d.Get("state").(string)
	head := d.Get("head_ref").(string)
	base := d.Get("base_ref").(string)
	sort := d.Get("sort_by").(string)
	direction := d.Get("sort_direction").(string)

	options := &github.PullRequestListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		State:       state,
		Head:        head,
		Base:        base,
		Sort:        sort,
		Direction:   direction,
	}

	results := make([]map[string]interface{}, 0)

	for {
		pullRequests, resp, err := client.PullRequests.List(ctx, owner, baseRepository, options)
		if err != nil {
			return err
		}

		for _, pullRequest := range pullRequests {
			result := map[string]interface{}{
				"number":                pullRequest.GetNumber(),
				"body":                  pullRequest.GetBody(),
				"draft":                 pullRequest.GetDraft(),
				"maintainer_can_modify": pullRequest.GetMaintainerCanModify(),
				"opened_at":             pullRequest.GetCreatedAt().Unix(),
				"state":                 pullRequest.GetState(),
				"title":                 pullRequest.GetTitle(),
				"updated_at":            pullRequest.GetUpdatedAt().Unix(),
			}

			if head := pullRequest.GetHead(); head != nil {
				result["head_ref"] = head.GetRef()
				result["head_sha"] = head.GetSHA()

				if headRepo := head.GetRepo(); headRepo != nil {
					result["head_repository"] = headRepo.GetName()

					if headOwner := headRepo.GetOwner(); headOwner != nil {
						result["head_owner"] = headOwner.GetLogin()
					}
				}
			}

			if base := pullRequest.GetBase(); base != nil {
				result["base_ref"] = base.GetRef()
				result["base_sha"] = base.GetSHA()
			}

			labels := []string{}
			for _, label := range pullRequest.Labels {
				labels = append(labels, label.GetName())
			}
			result["labels"] = labels

			if user := pullRequest.GetUser(); user != nil {
				result["opened_by"] = user.GetLogin()
			}

			results = append(results, result)
		}

		if resp.NextPage == 0 {
			break
		}

		options.Page = resp.NextPage
	}

	d.SetId(strings.Join([]string{
		owner,
		baseRepository,
		state,
		head,
		base,
		sort,
		direction,
	}, "/"))

	d.Set("results", results)

	return nil
}
