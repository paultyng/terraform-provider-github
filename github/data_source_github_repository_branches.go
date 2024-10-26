package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGithubRepositoryBranches() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGithubRepositoryBranchesRead,
		Schema: map[string]*schema.Schema{
			"repository": {
				Type:     schema.TypeString,
				Required: true,
			},
			"only_protected_branches": {
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{"only_non_protected_branches"},
			},
			"only_non_protected_branches": {
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{"only_protected_branches"},
			},
			"branches": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"protected": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func flattenBranches(branches []*github.Branch) []map[string]interface{} {
	results := make([]map[string]interface{}, 0)
	if branches == nil {
		return results
	}

	for _, branch := range branches {
		branchMap := make(map[string]interface{})
		branchMap["name"] = branch.GetName()
		branchMap["protected"] = branch.GetProtected()
		results = append(results, branchMap)
	}

	return results
}

func dataSourceGithubRepositoryBranchesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name
	repoName := d.Get("repository").(string)

	onlyProtectedBranches := d.Get("only_protected_branches").(bool)
	onlyNonProtectedBranches := d.Get("only_non_protected_branches").(bool)
	var listBranchOptions *github.BranchListOptions
	if onlyProtectedBranches {
		listBranchOptions = &github.BranchListOptions{
			Protected: &onlyProtectedBranches,
		}
	} else if onlyNonProtectedBranches {
		listBranchOptions = &github.BranchListOptions{
			Protected: &onlyProtectedBranches,
		}
	} else {
		listBranchOptions = &github.BranchListOptions{}
	}

	results := make([]map[string]interface{}, 0)
	for {
		branches, resp, err := client.Repositories.ListBranches(context.TODO(), orgName, repoName, listBranchOptions)
		if err != nil {
			return err
		}
		results = append(results, flattenBranches(branches)...)

		if resp.NextPage == 0 {
			break
		}

		listBranchOptions.Page = resp.NextPage
	}

	d.SetId(fmt.Sprintf("%s/%s", orgName, repoName))
	err := d.Set("repository", repoName)
	if err != nil {
		return err
	}
	err = d.Set("branches", results)
	if err != nil {
		return err
	}

	return nil
}
