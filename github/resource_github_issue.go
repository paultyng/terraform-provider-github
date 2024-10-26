package github

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGithubIssue() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubIssueCreateOrUpdate,
		Read:   resourceGithubIssueRead,
		Update: resourceGithubIssueCreateOrUpdate,
		Delete: resourceGithubIssueDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"repository": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The GitHub repository name.",
			},
			"number": {
				Type:        schema.TypeInt,
				Required:    false,
				Computed:    true,
				Description: "The issue number.",
			},
			"title": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Title of the issue.",
			},
			"body": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Body of the issue.",
			},
			"labels": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
				Description: "List of labels to attach to the issue.",
			},
			"assignees": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
				Description: "List of Logins to assign to the issue.",
			},
			"milestone_number": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Milestone number to assign to the issue.",
			},
			"issue_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The issue id.",
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceGithubIssueCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	ctx := context.Background()
	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name
	repoName := d.Get("repository").(string)
	title := d.Get("title").(string)
	milestone := d.Get("milestone_number").(int)

	req := &github.IssueRequest{
		Title: github.String(title),
	}

	if v, ok := d.GetOk("body"); ok {
		req.Body = github.String(v.(string))
	}

	labels := expandStringList(d.Get("labels").(*schema.Set).List())
	req.Labels = &labels

	assignees := expandStringList(d.Get("assignees").(*schema.Set).List())
	req.Assignees = &assignees

	if milestone > 0 {
		req.Milestone = intPtr(milestone)
	}

	var issue *github.Issue
	var resp *github.Response
	var err error
	if d.IsNewResource() {
		log.Printf("[DEBUG] Creating issue: %s (%s/%s)",
			title, orgName, repoName)
		issue, resp, err = client.Issues.Create(ctx, orgName, repoName, req)
		if resp != nil {
			log.Printf("[DEBUG] Response from creating issue: %#v", *resp)
		}
	} else {
		number := d.Get("number").(int)
		log.Printf("[DEBUG] Updating issue: %d:%s (%s/%s)",
			number, title, orgName, repoName)
		issue, resp, err = client.Issues.Edit(ctx, orgName, repoName, number, req)
		if resp != nil {
			log.Printf("[DEBUG] Response from updating issue: %#v", *resp)
		}
	}
	if err != nil {
		return err
	}

	d.SetId(buildTwoPartID(repoName, strconv.Itoa(issue.GetNumber())))
	if err = d.Set("issue_id", issue.GetID()); err != nil {
		return err
	}
	return resourceGithubIssueRead(d, meta)
}

func resourceGithubIssueRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	repoName, idNumber, err := parseTwoPartID(d.Id(), "repository", "issue_number")
	if err != nil {
		return err
	}

	number, err := strconv.Atoi(idNumber)
	if err != nil {
		return err
	}

	orgName := meta.(*Owner).name
	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	if !d.IsNewResource() {
		ctx = context.WithValue(ctx, ctxEtag, d.Get("etag").(string))
	}

	log.Printf("[DEBUG] Reading issue: %d (%s/%s)", number, orgName, repoName)
	issue, resp, err := client.Issues.Get(ctx,
		orgName, repoName, number)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotModified {
				return nil
			}
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[WARN] Removing issue %d (%s/%s) from state because it no longer exists in GitHub",
					number, orgName, repoName)
				d.SetId("")
				return nil
			}
		}
		return err
	}

	if err = d.Set("etag", resp.Header.Get("ETag")); err != nil {
		return err
	}
	if err = d.Set("repository", repoName); err != nil {
		return err
	}
	if err = d.Set("number", number); err != nil {
		return err
	}
	if err = d.Set("title", issue.GetTitle()); err != nil {
		return err
	}
	if err = d.Set("body", issue.GetBody()); err != nil {
		return err
	}
	if err = d.Set("milestone_number", issue.GetMilestone().GetNumber()); err != nil {
		return err
	}

	var labels []string
	for _, v := range issue.Labels {
		labels = append(labels, v.GetName())
	}
	if err = d.Set("labels", flattenStringList(labels)); err != nil {
		return err
	}

	var assignees []string
	for _, v := range issue.Assignees {
		assignees = append(assignees, v.GetLogin())
	}
	if err = d.Set("assignees", flattenStringList(assignees)); err != nil {
		return err
	}

	if err = d.Set("issue_id", issue.GetID()); err != nil {
		return err
	}
	return nil
}

func resourceGithubIssueDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	orgName := meta.(*Owner).name
	repoName := d.Get("repository").(string)
	number := d.Get("number").(int)
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	log.Printf("[DEBUG] Deleting issue by closing: %d (%s/%s)", number, orgName, repoName)

	request := &github.IssueRequest{State: github.String("closed")}

	_, _, err := client.Issues.Edit(ctx, orgName, repoName, number, request)

	return err
}

func intPtr(i int) *int {
	return &i
}
