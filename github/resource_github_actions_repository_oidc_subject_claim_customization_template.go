package github

import (
	"context"
	"errors"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceGithubActionsRepositoryOIDCSubjectClaimCustomizationTemplate() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubActionsRepositoryOIDCSubjectClaimCustomizationTemplateCreateOrUpdate,
		Read:   resourceGithubActionsRepositoryOIDCSubjectClaimCustomizationTemplateRead,
		Update: resourceGithubActionsRepositoryOIDCSubjectClaimCustomizationTemplateCreateOrUpdate,
		Delete: resourceGithubActionsRepositoryOIDCSubjectClaimCustomizationTemplateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"repository": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "The name of the repository.",
				ValidateDiagFunc: toDiagFunc(validation.StringLenBetween(1, 100), "repository"),
			},
			"use_default": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Whether to use the default template or not. If 'true', 'include_claim_keys' must not be set.",
			},
			"include_claim_keys": {
				Type:        schema.TypeList,
				Optional:    true,
				MinItems:    1,
				Description: "A list of OpenID Connect claims.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceGithubActionsRepositoryOIDCSubjectClaimCustomizationTemplateCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Owner).v3client

	repository := d.Get("repository").(string)
	owner := meta.(*Owner).name

	useDefault := d.Get("use_default").(bool)
	includeClaimKeys, hasClaimKeys := d.GetOk("include_claim_keys")

	if useDefault && hasClaimKeys {
		return errors.New("include_claim_keys cannot be set when use_default is true")
	}

	customOIDCSubjectClaimTemplate := &github.OIDCSubjectClaimCustomTemplate{
		UseDefault: &useDefault,
	}

	if includeClaimKeys != nil {

		includeClaimKeysVal := includeClaimKeys.([]interface{})

		claimsStr := make([]string, len(includeClaimKeysVal))

		for i, v := range includeClaimKeysVal {
			claimsStr[i] = v.(string)
		}

		customOIDCSubjectClaimTemplate.IncludeClaimKeys = claimsStr
	}

	ctx := context.Background()
	_, err := client.Actions.SetRepoOIDCSubjectClaimCustomTemplate(ctx, owner, repository, customOIDCSubjectClaimTemplate)

	if err != nil {
		return err
	}

	d.SetId(repository)
	return resourceGithubActionsRepositoryOIDCSubjectClaimCustomizationTemplateRead(d, meta)
}

func resourceGithubActionsRepositoryOIDCSubjectClaimCustomizationTemplateRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	repository := d.Id()
	owner := meta.(*Owner).name

	ctx := context.Background()
	template, _, err := client.Actions.GetRepoOIDCSubjectClaimCustomTemplate(ctx, owner, repository)

	if err != nil {
		return err
	}

	if err = d.Set("repository", repository); err != nil {
		return err
	}
	if err = d.Set("use_default", template.UseDefault); err != nil {
		return err
	}
	if err = d.Set("include_claim_keys", template.IncludeClaimKeys); err != nil {
		return err
	}

	return nil
}

func resourceGithubActionsRepositoryOIDCSubjectClaimCustomizationTemplateDelete(d *schema.ResourceData, meta interface{}) error {
	// Reset the repository to use the default claims
	// https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/about-security-hardening-with-openid-connect#using-the-default-subject-claims
	client := meta.(*Owner).v3client

	repository := d.Get("repository").(string)
	owner := meta.(*Owner).name

	customOIDCSubjectClaimTemplate := &github.OIDCSubjectClaimCustomTemplate{
		UseDefault: github.Bool(true),
	}

	ctx := context.Background()
	_, err := client.Actions.SetRepoOIDCSubjectClaimCustomTemplate(ctx, owner, repository, customOIDCSubjectClaimTemplate)

	if err != nil {
		return err
	}

	return nil
}
