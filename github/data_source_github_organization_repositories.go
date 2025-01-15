package github

import (
	"github.com/google/go-github/v66/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGithubOrganizationRepositories() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGithubOrganizationRepositoriesRead,
		Schema: map[string]*schema.Schema{
			"repositories": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"repo_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"node_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"archived": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"visibility": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceGithubOrganizationRepositoriesRead(d *schema.ResourceData, meta interface{}) error {
	org := meta.(*Owner).name
	client3 := meta.(*Owner).v3client
	ctx := meta.(*Owner).StopContext

	options := github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	var allRepositories []map[string]interface{}
	for {
		repositories, resp, err := client3.Repositories.ListByOrg(ctx, org, &options)
		if err != nil {
			return err
		}
		for _, repository := range repositories {
			repo := make(map[string]interface{})
			repo["repo_id"] = repository.GetID()
			repo["node_id"] = repository.GetNodeID()
			repo["name"] = repository.GetName()
			repo["archived"] = repository.GetArchived()
			repo["visibility"] = repository.GetVisibility()
			allRepositories = append(allRepositories, repo)
		}
		if resp.NextPage == 0 {
			break
		}
		options.Page = resp.NextPage
	}

	d.SetId(org)
	d.Set("repositories", allRepositories)

	return nil
}
