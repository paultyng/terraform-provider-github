package github

import (
	"context"
	"fmt"
	"net/url"

	"github.com/google/go-github/v65/github"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGithubActionsEnvironmentVariables() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGithubActionsEnvironmentVariablesRead,

		Schema: map[string]*schema.Schema{
			"full_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"name"},
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"full_name"},
			},
			"environment": {
				Type:     schema.TypeString,
				Required: true,
			},
			"variables": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"value": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"created_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"updated_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceGithubActionsEnvironmentVariablesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	var repoName string

	envName := d.Get("environment").(string)
	escapedEnvName := url.PathEscape(envName)

	if fullName, ok := d.GetOk("full_name"); ok {
		var err error
		owner, repoName, err = splitRepoFullName(fullName.(string))
		if err != nil {
			return err
		}
	}

	if name, ok := d.GetOk("name"); ok {
		repoName = name.(string)
	}

	if repoName == "" {
		return fmt.Errorf("one of %q or %q has to be provided", "full_name", "name")
	}

	options := github.ListOptions{
		PerPage: 100,
	}

	var all_variables []map[string]string
	for {
		variables, resp, err := client.Actions.ListEnvVariables(context.TODO(), owner, repoName, escapedEnvName, &options)
		if err != nil {
			return err
		}
		for _, variable := range variables.Variables {
			new_variable := map[string]string{
				"name":       variable.Name,
				"value":      variable.Value,
				"created_at": variable.CreatedAt.String(),
				"updated_at": variable.UpdatedAt.String(),
			}
			all_variables = append(all_variables, new_variable)
		}
		if resp.NextPage == 0 {
			break
		}
		options.Page = resp.NextPage
	}

	d.SetId(buildTwoPartID(repoName, envName))
	d.Set("variables", all_variables)

	return nil
}
