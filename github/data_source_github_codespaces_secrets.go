package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v65/github"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGithubCodespacesSecrets() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGithubCodespacesSecretsRead,

		Schema: map[string]*schema.Schema{
			"full_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "Full name of the repository (in `org/name` format).",
				ConflictsWith: []string{"name"},
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"full_name"},
				Description:   "The name of the repository.",
			},
			"secrets": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
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

func dataSourceGithubCodespacesSecretsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.Background()

	var repoName string

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

	var all_secrets []map[string]string
	for {
		secrets, resp, err := client.Codespaces.ListRepoSecrets(ctx, owner, repoName, &options)
		if err != nil {
			return err
		}
		for _, secret := range secrets.Secrets {
			new_secret := map[string]string{
				"name":       secret.Name,
				"created_at": secret.CreatedAt.String(),
				"updated_at": secret.UpdatedAt.String(),
			}
			all_secrets = append(all_secrets, new_secret)
		}
		if resp.NextPage == 0 {
			break
		}
		options.Page = resp.NextPage
	}

	d.SetId(repoName)
	err := d.Set("secrets", all_secrets)
	if err != nil {
		return err
	}

	return nil
}
