package github

import (
	"context"
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGithubActionsEnvironmentPublicKey() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGithubActionsEnvironmentPublicKeyRead,

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:     schema.TypeString,
				Required: true,
			},
			"environment": {
				Type:     schema.TypeString,
				Required: true,
			},
			"key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceGithubActionsEnvironmentPublicKeyRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	repository := d.Get("repository").(string)

	envName := d.Get("environment").(string)
	escapedEnvName := url.PathEscape(envName)

	repo, _, err := client.Repositories.Get(context.TODO(), owner, repository)
	if err != nil {
		return err
	}

	publicKey, _, err := client.Actions.GetEnvPublicKey(context.TODO(), int(repo.GetID()), escapedEnvName)
	if err != nil {
		return err
	}

	d.SetId(publicKey.GetKeyID())
	err = d.Set("key_id", publicKey.GetKeyID())
	if err != nil {
		return err
	}
	err = d.Set("key", publicKey.GetKey())
	if err != nil {
		return err
	}

	return nil
}
