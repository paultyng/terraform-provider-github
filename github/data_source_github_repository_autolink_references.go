package github

import (
	"context"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGithubRepositoryAutolinkReferences() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGithubRepositoryAutolinkReferencesRead,

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:     schema.TypeString,
				Required: true,
			},
			"autolink_references": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key_prefix": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"target_url_template": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"is_alphanumeric": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceGithubRepositoryAutolinkReferencesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name
	repoName := d.Get("repository").(string)

	results := make([]map[string]interface{}, 0)

	var listOptions *github.ListOptions
	for {
		autoLinks, resp, err := client.Repositories.ListAutolinks(context.Background(), orgName, repoName, listOptions)
		if err != nil {
			return err
		}

		results = append(results, flattenAutolinkReferences(autoLinks)...)

		if resp.NextPage == 0 {
			break
		}

		listOptions.Page = resp.NextPage
	}

	d.SetId(repoName)
	err := d.Set("autolink_references", results)
	if err != nil {
		return err
	}

	return nil
}

func flattenAutolinkReferences(autoLinks []*github.Autolink) []map[string]interface{} {
	results := make([]map[string]interface{}, 0)
	if autoLinks == nil {
		return results
	}

	for _, autolink := range autoLinks {
		linkMap := make(map[string]interface{})
		linkMap["key_prefix"] = autolink.GetKeyPrefix()
		linkMap["target_url_template"] = autolink.GetURLTemplate()
		linkMap["is_alphanumeric"] = autolink.GetIsAlphanumeric()
		results = append(results, linkMap)
	}

	return results
}
