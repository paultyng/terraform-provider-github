package github

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGithubOrganizationProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubOrganizationProjectCreate,
		Read:   resourceGithubOrganizationProjectRead,
		Update: resourceGithubOrganizationProjectUpdate,
		Delete: resourceGithubOrganizationProjectDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the project.",
			},
			"body": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The body of the project.",
			},
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL of the project.",
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceGithubOrganizationProjectCreate(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name
	name := d.Get("name").(string)
	body := d.Get("body").(string)
	ctx := context.Background()

	project, _, err := client.Organizations.CreateProject(ctx,
		orgName,
		&github.ProjectOptions{
			Name: &name,
			Body: &body,
		},
	)
	if err != nil {
		return err
	}
	d.SetId(strconv.FormatInt(project.GetID(), 10))

	return resourceGithubOrganizationProjectRead(d, meta)
}

func resourceGithubOrganizationProjectRead(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name

	projectID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}
	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	if !d.IsNewResource() {
		ctx = context.WithValue(ctx, ctxEtag, d.Get("etag").(string))
	}

	project, resp, err := client.Projects.GetProject(ctx, projectID)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotModified {
				return nil
			}
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[INFO] Removing organization project %s/%s from state because it no longer exists in GitHub",
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
	if err = d.Set("name", project.GetName()); err != nil {
		return err
	}
	if err = d.Set("body", project.GetBody()); err != nil {
		return err
	}
	if err = d.Set("url", fmt.Sprintf("https://github.com/orgs/%s/projects/%d",
		orgName, project.GetNumber())); err != nil {
		return err
	}

	return nil
}

func resourceGithubOrganizationProjectUpdate(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client

	name := d.Get("name").(string)
	body := d.Get("body").(string)

	options := github.ProjectOptions{
		Name: &name,
		Body: &body,
	}

	projectID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	if _, _, err := client.Projects.UpdateProject(ctx, projectID, &options); err != nil {
		return err
	}

	return resourceGithubOrganizationProjectRead(d, meta)
}

func resourceGithubOrganizationProjectDelete(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client
	projectID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	_, err = client.Projects.DeleteProject(ctx, projectID)
	return err
}
