package github

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceGithubRepositoryAutolinkReference() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubRepositoryAutolinkReferenceCreate,
		Read:   resourceGithubRepositoryAutolinkReferenceRead,
		Delete: resourceGithubRepositoryAutolinkReferenceDelete,

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				parts := strings.Split(d.Id(), "/")
				if len(parts) != 2 {
					return nil, fmt.Errorf("invalid ID specified: supplied ID must be written as <repository>/<autolink_reference_id>")
				}

				repository := parts[0]
				id := parts[1]

				// If the second part of the provided ID isn't an integer, assume that the
				// caller provided the key prefix for the autolink reference, and look up
				// the autolink by the key prefix.

				_, err := strconv.Atoi(id)
				if err != nil {
					client := meta.(*Owner).v3client
					owner := meta.(*Owner).name

					autolink, err := getAutolinkByKeyPrefix(client, owner, repository, id)
					if err != nil {
						return nil, err
					}

					id = strconv.FormatInt(*autolink.ID, 10)
				}

				if err = d.Set("repository", repository); err != nil {
					return nil, err
				}
				d.SetId(id)
				return []*schema.ResourceData{d}, nil
			},
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"repository": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The repository name",
			},
			"key_prefix": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "This prefix appended by a number will generate a link any time it is found in an issue, pull request, or commit",
			},
			"target_url_template": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				Description:      "The template of the target URL used for the links; must be a valid URL and contain `<num>` for the reference number",
				ValidateDiagFunc: toDiagFunc(validation.StringMatch(regexp.MustCompile(`^http[s]?:\/\/[a-z0-9-.]*(:[0-9]+)?\/.*?<num>.*?$`), "must be a valid URL and contain <num> token"), "target_url_template"),
			},
			"is_alphanumeric": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Default:     true,
				Description: "Whether this autolink reference matches alphanumeric characters. If false, this autolink reference only matches numeric characters.",
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceGithubRepositoryAutolinkReferenceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	owner := meta.(*Owner).name
	repoName := d.Get("repository").(string)
	keyPrefix := d.Get("key_prefix").(string)
	targetURLTemplate := d.Get("target_url_template").(string)
	isAlphanumeric := d.Get("is_alphanumeric").(bool)
	ctx := context.Background()

	opts := &github.AutolinkOptions{
		KeyPrefix:      &keyPrefix,
		URLTemplate:    &targetURLTemplate,
		IsAlphanumeric: &isAlphanumeric,
	}

	autolinkRef, _, err := client.Repositories.AddAutolink(ctx, owner, repoName, opts)
	if err != nil {
		return err
	}
	d.SetId(strconv.FormatInt(autolinkRef.GetID(), 10))

	return resourceGithubRepositoryAutolinkReferenceRead(d, meta)
}

func resourceGithubRepositoryAutolinkReferenceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	owner := meta.(*Owner).name
	repoName := d.Get("repository").(string)
	autolinkRefID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return unconvertibleIdErr(d.Id(), err)
	}
	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	if !d.IsNewResource() {
		ctx = context.WithValue(ctx, ctxEtag, d.Get("etag").(string))
	}

	autolinkRef, _, err := client.Repositories.GetAutolink(ctx, owner, repoName, autolinkRefID)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[INFO] Removing autolink reference for repository %s/%s from state because it no longer exists in GitHub",
					owner, repoName)
				d.SetId("")
				return nil
			}
		}
		return err
	}

	// Set resource fields
	d.SetId(strconv.FormatInt(autolinkRef.GetID(), 10))
	if err = d.Set("repository", repoName); err != nil {
		return err
	}
	if err = d.Set("key_prefix", autolinkRef.KeyPrefix); err != nil {
		return err
	}
	if err = d.Set("target_url_template", autolinkRef.URLTemplate); err != nil {
		return err
	}
	if err = d.Set("is_alphanumeric", autolinkRef.IsAlphanumeric); err != nil {
		return err
	}

	return nil
}

func resourceGithubRepositoryAutolinkReferenceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	owner := meta.(*Owner).name
	repoName := d.Get("repository").(string)
	autolinkRefID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return unconvertibleIdErr(d.Id(), err)
	}
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	_, err = client.Repositories.DeleteAutolink(ctx, owner, repoName, autolinkRefID)
	return err
}
