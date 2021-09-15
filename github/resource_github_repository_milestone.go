package github

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v39/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceGithubRepositoryMilestone() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubRepositoryMilestoneCreate,
		Read:   resourceGithubRepositoryMilestoneRead,
		Update: resourceGithubRepositoryMilestoneUpdate,
		Delete: resourceGithubRepositoryMilestoneDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				parts := strings.Split(d.Id(), "/")
				if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
					return nil, fmt.Errorf("Invalid ID format, must be provided as OWNER/REPOSITORY/NUMBER")
				}
				d.Set("owner", parts[0])
				d.Set("repository", parts[1])
				number, err := strconv.Atoi(parts[2])
				if err != nil {
					return nil, err
				}
				d.Set("number", number)
				d.SetId(fmt.Sprintf("%s/%s/%d", parts[0], parts[1], number))

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"title": {
				Type:     schema.TypeString,
				Required: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Required: true,
			},
			"repository": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"due_date": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "in yyyy-mm-dd format",
			},
			"state": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"open", "closed",
				}, true),
				Default: "open",
			},
			"number": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

const (
	layoutISO = "2006-01-02"
)

func resourceGithubRepositoryMilestoneCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*Owner).v3client
	ctx := context.Background()
	owner := d.Get("owner").(string)
	repoName := d.Get("repository").(string)

	milestone := &github.Milestone{
		Title: github.String(d.Get("title").(string)),
	}

	if v, ok := d.GetOk("description"); ok && len(v.(string)) > 0 {
		milestone.Description = github.String(v.(string))
	}
	if v, ok := d.GetOk("due_date"); ok && len(v.(string)) > 0 {
		dueDate, err := time.Parse(layoutISO, v.(string))
		if err != nil {
			return err
		}
		date := time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 23, 39, 0, 0, time.UTC)
		milestone.DueOn = &date
	}
	if v, ok := d.GetOk("state"); ok && len(v.(string)) > 0 {
		milestone.State = github.String(v.(string))
	}

	log.Printf("[DEBUG] Creating milestone for repository: %s/%s", owner, repoName)
	milestone, _, err := conn.Issues.CreateMilestone(ctx, owner, repoName, milestone)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%s/%s/%d", owner, repoName, milestone.GetNumber()))

	return resourceGithubRepositoryMilestoneRead(d, meta)
}

func resourceGithubRepositoryMilestoneRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*Owner).v3client
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	owner := d.Get("owner").(string)
	repoName := d.Get("repository").(string)
	number, err := parseMilestoneNumber(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Reading milestone for repository: %s/%s", owner, repoName)
	milestone, _, err := conn.Issues.GetMilestone(ctx, owner, repoName, number)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotModified {
				return nil
			}
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[WARN] Removing milestone for %s/%s from state because it no longer exists in GitHub",
					owner, repoName)
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.Set("title", milestone.GetTitle())
	d.Set("description", milestone.GetDescription())
	d.Set("number", milestone.GetNumber())
	d.Set("state", milestone.GetState())
	if dueOn := milestone.GetDueOn(); !dueOn.IsZero() {
		d.Set("due_date", milestone.GetDueOn().Format(layoutISO))
	}

	return nil
}

func resourceGithubRepositoryMilestoneUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*Owner).v3client
	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	owner := d.Get("owner").(string)
	repoName := d.Get("repository").(string)
	number, err := parseMilestoneNumber(d.Id())
	if err != nil {
		return err
	}

	milestone := &github.Milestone{}
	if d.HasChanges("title") {
		_, n := d.GetChange("title")
		milestone.Title = github.String(n.(string))
	}

	if d.HasChanges("description") {
		_, n := d.GetChange("description")
		milestone.Description = github.String(n.(string))
	}

	if d.HasChanges("due_date") {
		_, n := d.GetChange("due_date")
		dueDate, err := time.Parse(layoutISO, n.(string))
		if err != nil {
			return err
		}
		date := time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 7, 0, 0, 0, time.UTC)
		milestone.DueOn = &date
	}

	if d.HasChanges("state") {
		_, n := d.GetChange("state")
		milestone.State = github.String(n.(string))
	}

	log.Printf("[DEBUG] Updating milestone for repository: %s/%s", owner, repoName)
	_, _, err = conn.Issues.EditMilestone(ctx, owner, repoName, number, milestone)
	if err != nil {
		return err
	}

	return resourceGithubRepositoryMilestoneRead(d, meta)
}

func resourceGithubRepositoryMilestoneDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*Owner).v3client
	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	owner := d.Get("owner").(string)
	repoName := d.Get("repository").(string)
	number, err := parseMilestoneNumber(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Deleting milestone for repository: %s/%s", owner, repoName)
	_, err = conn.Issues.DeleteMilestone(ctx, owner, repoName, number)
	if err != nil {
		return err
	}

	return nil
}

func parseMilestoneNumber(id string) (int, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 3 {
		return -1, fmt.Errorf("ID not properly formatted: %s", id)
	}
	number, err := strconv.Atoi(parts[2])
	if err != nil {
		return -1, err
	}
	return number, nil
}
