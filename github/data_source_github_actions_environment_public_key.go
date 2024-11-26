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
			"repository_id": {
				Type:        schema.TypeInt,
				Description: "The repository in which the Environment is defined.",
				Required:    true,
			},
			"environment": {
				Type:        schema.TypeString,
				Description: "The name of the Environment.",
				Required:    true,
			},
			"key_id": {
				Type:        schema.TypeString,
				Description: "The ID of the public key.",
				Computed:    true,
			},
			"key": {
				Type:        schema.TypeString,
				Description: "The public key value.",
				Computed:    true,
			},
		},
	}
}

func dataSourceGithubActionsEnvironmentPublicKeyRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	ctx := context.Background()

	envName := d.Get("environment").(string)
	escapedEnvName := url.PathEscape(envName)

	repoId := d.Get("repository_id").(int)

	publicKey, _, err := client.Actions.GetEnvPublicKey(ctx, repoId, escapedEnvName)
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
