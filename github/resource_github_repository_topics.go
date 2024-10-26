package github

import (
	"context"
	"log"
	"net/http"
	"regexp"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceGithubRepositoryTopics() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubRepositoryTopicsCreateOrUpdate,
		Read:   resourceGithubRepositoryTopicsRead,
		Update: resourceGithubRepositoryTopicsCreateOrUpdate,
		Delete: resourceGithubRepositoryTopicsDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("repository", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			"repository": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[-a-zA-Z0-9_.]{1,100}$`), "must include only alphanumeric characters, underscores or hyphens and consist of 100 characters or less"),
				Description:  "The name of the repository. The name is not case sensitive.",
			},
			"topics": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "An array of topics to add to the repository. Pass one or more topics to replace the set of existing topics. Send an empty array ([]) to clear all topics from the repository. Note: Topic names cannot contain uppercase letters.",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,49}$`), "must include only lowercase alphanumeric characters or hyphens and cannot start with a hyphen and consist of 50 characters or less"),
				},
			}},
	}

}

func resourceGithubRepositoryTopicsCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	ctx := context.Background()

	owner := meta.(*Owner).name
	repoName := d.Get("repository").(string)
	topics := expandStringList(d.Get("topics").(*schema.Set).List())

	if len(topics) > 0 {
		_, _, err := client.Repositories.ReplaceAllTopics(ctx, owner, repoName, topics)
		if err != nil {
			return err
		}
	}

	d.SetId(repoName)
	return resourceGithubRepositoryTopicsRead(d, meta)
}

func resourceGithubRepositoryTopicsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	owner := meta.(*Owner).name
	repoName := d.Get("repository").(string)

	topics, _, err := client.Repositories.ListAllTopics(ctx, owner, repoName)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotModified {
				return nil
			}
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[INFO] Removing topics from repository %s/%s from state because it no longer exists in GitHub",
					owner, repoName)
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.Set("topics", flattenStringList(topics))
	return nil
}

func resourceGithubRepositoryTopicsDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	owner := meta.(*Owner).name
	repoName := d.Get("repository").(string)

	_, _, err := client.Repositories.ReplaceAllTopics(ctx, owner, repoName, []string{})
	if err != nil {
		return err
	}

	return nil
}
