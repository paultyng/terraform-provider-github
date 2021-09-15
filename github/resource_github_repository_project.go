package github

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/go-github/v39/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceGithubRepositoryProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubRepositoryProjectCreate,
		Read:   resourceGithubRepositoryProjectRead,
		Update: resourceGithubRepositoryProjectUpdate,
		Delete: resourceGithubRepositoryProjectDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				parts := strings.Split(d.Id(), "/")
				if len(parts) != 2 {
					return nil, fmt.Errorf("Invalid ID specified. Supplied ID must be written as <repository>/<project_id>")
				}
				d.Set("repository", parts[0])
				d.SetId(parts[1])
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"repository": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"body": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceGithubRepositoryProjectCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	owner := meta.(*Owner).name
	repoName := d.Get("repository").(string)
	name := d.Get("name").(string)
	body := d.Get("body").(string)

	options := github.ProjectOptions{
		Name: &name,
		Body: &body,
	}
	ctx := context.Background()

	log.Printf("[DEBUG] Creating repository project: %s (%s/%s)", name, owner, repoName)
	project, _, err := client.Repositories.CreateProject(ctx,
		owner, repoName, &options)
	if err != nil {
		return err
	}
	d.SetId(strconv.FormatInt(project.GetID(), 10))

	return resourceGithubRepositoryProjectRead(d, meta)
}

func resourceGithubRepositoryProjectRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name

	projectID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return unconvertibleIdErr(d.Id(), err)
	}
	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	if !d.IsNewResource() {
		ctx = context.WithValue(ctx, ctxEtag, d.Get("etag").(string))
	}

	log.Printf("[DEBUG] Reading repository project: %s", d.Id())
	project, resp, err := client.Projects.GetProject(ctx, projectID)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotModified {
				return nil
			}
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[WARN] Removing repository project %s from state because it no longer exists in GitHub",
					d.Id())
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.Set("etag", resp.Header.Get("ETag"))
	d.Set("name", project.GetName())
	d.Set("body", project.GetBody())
	d.Set("url", fmt.Sprintf("https://github.com/%s/%s/projects/%d",
		owner, d.Get("repository"), project.GetNumber()))

	return nil
}

func resourceGithubRepositoryProjectUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	name := d.Get("name").(string)
	body := d.Get("body").(string)

	options := github.ProjectOptions{
		Name: &name,
		Body: &body,
	}

	projectID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return unconvertibleIdErr(d.Id(), err)
	}
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	log.Printf("[DEBUG] Updating repository project: %s", d.Id())
	_, _, err = client.Projects.UpdateProject(ctx, projectID, &options)
	if err != nil {
		return err
	}

	return resourceGithubRepositoryProjectRead(d, meta)
}

func resourceGithubRepositoryProjectDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	projectID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return unconvertibleIdErr(d.Id(), err)
	}
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	log.Printf("[DEBUG] Deleting repository project: %s", d.Id())
	_, err = client.Projects.DeleteProject(ctx, projectID)
	return err
}
