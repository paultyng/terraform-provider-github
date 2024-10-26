package github

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGithubRepositoryWebhook() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubRepositoryWebhookCreate,
		Read:   resourceGithubRepositoryWebhookRead,
		Update: resourceGithubRepositoryWebhookUpdate,
		Delete: resourceGithubRepositoryWebhookDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				parts := strings.Split(d.Id(), "/")
				if len(parts) != 2 {
					return nil, fmt.Errorf("invalid ID specified: supplied ID must be written as <repository>/<webhook_id>")
				}
				if err := d.Set("repository", parts[0]); err != nil {
					return nil, err
				}
				d.SetId(parts[1])
				return []*schema.ResourceData{d}, nil
			},
		},

		SchemaVersion: 1,
		MigrateState:  resourceGithubWebhookMigrateState,

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The repository of the webhook.",
			},
			"events": {
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Description: "A list of events which should trigger the webhook",
			},
			"configuration": webhookConfigurationSchema(),
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Configuration block for the webhook",
			},
			"active": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Indicate if the webhook should receive events. Defaults to 'true'.",
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceGithubRepositoryWebhookObject(d *schema.ResourceData) *github.Hook {
	url := d.Get("url").(string)
	active := d.Get("active").(bool)
	events := []string{}
	eventSet := d.Get("events").(*schema.Set)
	for _, v := range eventSet.List() {
		events = append(events, v.(string))
	}

	hook := &github.Hook{
		URL:    &url,
		Events: events,
		Active: &active,
	}

	config := d.Get("configuration").([]interface{})[0].(map[string]interface{})
	if len(config) > 0 {
		hook.Config = webhookConfigFromInterface(config)
	}

	return hook
}

func resourceGithubRepositoryWebhookCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	owner := meta.(*Owner).name
	repoName := d.Get("repository").(string)
	hk := resourceGithubRepositoryWebhookObject(d)
	ctx := context.Background()

	hook, _, err := client.Repositories.CreateHook(ctx, owner, repoName, hk)
	if err != nil {
		return err
	}
	d.SetId(strconv.FormatInt(hook.GetID(), 10))

	// GitHub returns the secret as a string of 8 astrisks "********"
	// We would prefer to store the real secret in state, so we'll
	// write the configuration secret in state from our request to GitHub
	if hook.Config.Secret != nil {
		hook.Config.Secret = hk.Config.Secret
	}

	if err = d.Set("configuration", interfaceFromWebhookConfig(hook.Config)); err != nil {
		return err
	}

	return resourceGithubRepositoryWebhookRead(d, meta)
}

func resourceGithubRepositoryWebhookRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	owner := meta.(*Owner).name
	repoName := d.Get("repository").(string)
	hookID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return unconvertibleIdErr(d.Id(), err)
	}
	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	if !d.IsNewResource() {
		ctx = context.WithValue(ctx, ctxEtag, d.Get("etag").(string))
	}

	hook, _, err := client.Repositories.GetHook(ctx, owner, repoName, hookID)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotModified {
				return nil
			}
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[INFO] Removing repository webhook %s from state because it no longer exists in GitHub",
					d.Id())
				d.SetId("")
				return nil
			}
		}
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
		currentSecret := d.Get("configuration").([]interface{})[0].(map[string]interface{})["secret"]

		if hook.Config.Secret != nil {
			hook.Config.Secret = github.String(currentSecret.(string))
		}
	}

	if err = d.Set("configuration", interfaceFromWebhookConfig(hook.Config)); err != nil {
		return err
	}

	return nil
}

func resourceGithubRepositoryWebhookUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	owner := meta.(*Owner).name
	repoName := d.Get("repository").(string)
	hk := resourceGithubRepositoryWebhookObject(d)
	hookID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return unconvertibleIdErr(d.Id(), err)
	}
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	_, _, err = client.Repositories.EditHook(ctx, owner, repoName, hookID, hk)
	if err != nil {
		return err
	}

	return resourceGithubRepositoryWebhookRead(d, meta)
}

func resourceGithubRepositoryWebhookDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	owner := meta.(*Owner).name
	repoName := d.Get("repository").(string)
	hookID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return unconvertibleIdErr(d.Id(), err)
	}
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	_, err = client.Repositories.DeleteHook(ctx, owner, repoName, hookID)
	return err
}
