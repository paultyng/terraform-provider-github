package github

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/google/go-github/v39/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceGithubOrganizationProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubOrganizationProjectCreate,
		Read:   resourceGithubOrganizationProjectRead,
		Update: resourceGithubOrganizationProjectUpdate,
		Delete: resourceGithubOrganizationProjectDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
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

	log.Printf("[DEBUG] Creating organization project: %s (%s)", name, orgName)
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

	log.Printf("[DEBUG] Reading organization project: %s (%s)", d.Id(), orgName)
	project, resp, err := client.Projects.GetProject(ctx, projectID)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotModified {
				return nil
			}
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[WARN] Removing organization project %s/%s from state because it no longer exists in GitHub",
					orgName, d.Id())
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.Set("etag", resp.Header.Get("ETag"))
	d.Set("name", project.GetName())
	d.Set("body", project.GetBody())
	d.Set("url", fmt.Sprintf("https://github.com/orgs/%s/projects/%d",
		orgName, project.GetNumber()))

	return nil
}

func resourceGithubOrganizationProjectUpdate(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name

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

	log.Printf("[DEBUG] Updating organization project: %s (%s)", d.Id(), orgName)
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
	orgName := meta.(*Owner).name
	projectID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	log.Printf("[DEBUG] Deleting organization project: %s (%s)", d.Id(), orgName)
	_, err = client.Projects.DeleteProject(ctx, projectID)
	return err
}
