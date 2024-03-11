package github

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/google/go-github/v60/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGithubOrganizationWebhook() *schema.Resource {

	return &schema.Resource{
		Create: resourceGithubOrganizationWebhookCreate,
		Read:   resourceGithubOrganizationWebhookRead,
		Update: resourceGithubOrganizationWebhookUpdate,
		Delete: resourceGithubOrganizationWebhookDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaVersion: 1,
		MigrateState:  resourceGithubWebhookMigrateState,

		Schema: map[string]*schema.Schema{
			"events": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "A list of events which should trigger the webhook.",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},
			"configuration": webhookConfigurationSchema(),
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL of the webhook.",
			},
			"active": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Indicate if the webhook should receive events.",
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceGithubOrganizationWebhookObject(d *schema.ResourceData) *github.Hook {
	events := []string{}
	eventSet := d.Get("events").(*schema.Set)
	for _, v := range eventSet.List() {
		events = append(events, v.(string))
	}

	hook := &github.Hook{
		URL:    github.String(d.Get("url").(string)),
		Events: events,
		Active: github.Bool(d.Get("active").(bool)),
	}

	config := d.Get("configuration").([]interface{})
	if len(config) > 0 {
		hook.Config = flattenWebhookConfig(config[0].(map[string]interface{}))
	}

	return hook
}

func resourceGithubOrganizationWebhookCreate(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client

	orgName := meta.(*Owner).name
	webhookObj := resourceGithubOrganizationWebhookObject(d)
	ctx := context.Background()

	hook, _, err := client.Organizations.CreateHook(ctx, orgName, webhookObj)

	if err != nil {
		return err
	}
	d.SetId(strconv.FormatInt(hook.GetID(), 10))

	// GitHub returns the secret as a string of 8 astrisks "********"
	// We would prefer to store the real secret in state, so we'll
	// write the configuration secret in state from our request to GitHub
	if hook.Config.Secret != nil {
		hook.Config.Secret = webhookObj.Config.Secret
	}

	hook.Config = insecureSslStringToBool(hook.Config)

	if err = d.Set("configuration", []interface{}{hook.Config}); err != nil {
		return err
	}

	return resourceGithubOrganizationWebhookRead(d, meta)
}

func resourceGithubOrganizationWebhookRead(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client

	orgName := meta.(*Owner).name
	hookID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return unconvertibleIdErr(d.Id(), err)
	}
	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	if !d.IsNewResource() {
		ctx = context.WithValue(ctx, ctxEtag, d.Get("etag").(string))
	}

	hook, resp, err := client.Organizations.GetHook(ctx, orgName, hookID)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotModified {
				return nil
			}
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[INFO] Removing organization webhook %s/%s from state because it no longer exists in GitHub",
					orgName, d.Id())
				d.SetId("")
				return nil
			}
		}
		return err
	}

	if err = d.Set("etag", resp.Header.Get("ETag")); err != nil {
		return err
	}
	if err = d.Set("url", hook.GetURL()); err != nil {
		return err
	}
	if err = d.Set("active", hook.GetActive()); err != nil {
		return err
	}
	if err = d.Set("events", hook.Events); err != nil {
		return err
	}

	// GitHub returns the secret as a string of 8 astrisks "********"
	// We would prefer to store the real secret in state, so we'll
	// write the configuration secret in state from what we get from
	// ResourceData
	if len(d.Get("configuration").([]interface{})) > 0 {
		currentSecret := d.Get("configuration").([]interface{})[0].(map[string]interface{})["secret"].(string)

		if hook.Config.Secret != nil {
			hook.Config.Secret = &currentSecret
		}
	}

	if err = d.Set("configuration", expandWebhookConfig(hook.Config)); err != nil {
		return err
	}

	return nil
}

func resourceGithubOrganizationWebhookUpdate(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client

	orgName := meta.(*Owner).name
	webhookObj := resourceGithubOrganizationWebhookObject(d)
	hookID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return unconvertibleIdErr(d.Id(), err)
	}
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	_, _, err = client.Organizations.EditHook(ctx,
		orgName, hookID, webhookObj)
	if err != nil {
		return err
	}

	return resourceGithubOrganizationWebhookRead(d, meta)
}

func resourceGithubOrganizationWebhookDelete(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client

	orgName := meta.(*Owner).name
	hookID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return unconvertibleIdErr(d.Id(), err)
	}
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	_, err = client.Organizations.DeleteHook(ctx, orgName, hookID)
	return err
}
