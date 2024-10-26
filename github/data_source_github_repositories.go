package github

import (
	"context"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceGithubRepositories() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGithubRepositoriesRead,

		Schema: map[string]*schema.Schema{
			"query": {
				Type:     schema.TypeString,
				Required: true,
			},
			"sort": {
				Type:             schema.TypeString,
				Default:          "updated",
				Optional:         true,
				ValidateDiagFunc: toDiagFunc(validation.StringInSlice([]string{"stars", "fork", "updated"}, false), "sort"),
			},
			"include_repo_id": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			"results_per_page": {
				Type:             schema.TypeInt,
				Optional:         true,
				Default:          100,
				ValidateDiagFunc: toDiagFunc(validation.IntBetween(0, 1000), "results_per_page"),
			},
			"full_names": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"names": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"repo_ids": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Computed: true,
			},
		},
	}
}

func dataSourceGithubRepositoriesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	includeRepoId := d.Get("include_repo_id").(bool)
	resultsPerPage := d.Get("results_per_page").(int)

	query := d.Get("query").(string)
	opt := &github.SearchOptions{
		Sort: d.Get("sort").(string),
		ListOptions: github.ListOptions{
			PerPage: resultsPerPage,
		},
	}

	fullNames, names, repoIDs, err := searchGithubRepositories(client, query, opt)
	if err != nil {
		return err
	}

	d.SetId(query)
	err = d.Set("full_names", fullNames)
	if err != nil {
		return err
	}
	err = d.Set("names", names)
	if err != nil {
		return err
	}
	if includeRepoId {
		err = d.Set("repo_ids", repoIDs)
		if err != nil {
			return err
		}
	}

	return nil
}

func searchGithubRepositories(client *github.Client, query string, opt *github.SearchOptions) ([]string, []string, []int64, error) {
	fullNames := make([]string, 0)

	names := make([]string, 0)

	repoIDs := make([]int64, 0)

	for {
		results, resp, err := client.Search.Repositories(context.TODO(), query, opt)
		if err != nil {
			return fullNames, names, repoIDs, err
		}

		for _, repo := range results.Repositories {
			fullNames = append(fullNames, repo.GetFullName())
			names = append(names, repo.GetName())
			repoIDs = append(repoIDs, repo.GetID())
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return fullNames, names, repoIDs, nil
}
