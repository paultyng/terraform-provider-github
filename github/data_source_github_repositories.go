package github

import (
	"context"
	"log"

	"github.com/google/go-github/v39/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
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
				Type:         schema.TypeString,
				Default:      "updated",
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"stars", "fork", "updated"}, false),
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
		},
	}
}

func dataSourceGithubRepositoriesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	query := d.Get("query").(string)
	opt := &github.SearchOptions{
		Sort: d.Get("sort").(string),
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	log.Printf("[DEBUG] Searching for GitHub repositories: %q", query)
	fullNames, names, err := searchGithubRepositories(client, query, opt)
	if err != nil {
		return err
	}

	d.SetId(query)
	d.Set("full_names", fullNames)
	d.Set("names", names)

	return nil
}

func searchGithubRepositories(client *github.Client, query string, opt *github.SearchOptions) ([]string, []string, error) {
	fullNames := make([]string, 0)

	names := make([]string, 0)

	for {
		results, resp, err := client.Search.Repositories(context.TODO(), query, opt)
		if err != nil {
			return fullNames, names, err
		}

		for _, repo := range results.Repositories {
			fullNames = append(fullNames, repo.GetFullName())
			names = append(names, repo.GetName())
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return fullNames, names, nil
}
