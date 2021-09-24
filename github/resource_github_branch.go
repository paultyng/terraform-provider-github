package github

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/go-github/v39/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceGithubBranch() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubBranchCreate,
		Read:   resourceGithubBranchRead,
		Delete: resourceGithubBranchDelete,
		Importer: &schema.ResourceImporter{
			State: resourceGithubBranchImport,
		},

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"branch": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"source_branch": {
				Type:     schema.TypeString,
				Default:  "main",
				Optional: true,
				ForceNew: true,
			},
			"source_sha": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ref": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sha": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceGithubBranchCreate(d *schema.ResourceData, meta interface{}) error {
	ctx := context.Background()
	if !d.IsNewResource() {
		ctx = context.WithValue(ctx, ctxId, d.Id())
	}

	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name
	repoName := d.Get("repository").(string)
	branchName := d.Get("branch").(string)
	branchRefName := "refs/heads/" + branchName
	sourceBranchName := d.Get("source_branch").(string)
	sourceBranchRefName := "refs/heads/" + sourceBranchName

	if _, hasSourceSHA := d.GetOk("source_sha"); !hasSourceSHA {
		log.Printf("[DEBUG] Querying GitHub branch reference %s/%s (%s) to derive source_sha",
			orgName, repoName, sourceBranchRefName)
		ref, _, err := client.Git.GetRef(ctx, orgName, repoName, sourceBranchRefName)
		if err != nil {
			return fmt.Errorf("Error querying GitHub branch reference %s/%s (%s): %s",
				orgName, repoName, sourceBranchRefName, err)
		}
		d.Set("source_sha", *ref.Object.SHA)
	}
	sourceBranchSHA := d.Get("source_sha").(string)

	log.Printf("[DEBUG] Creating GitHub branch reference %s/%s (%s)",
		orgName, repoName, branchRefName)
	_, _, err := client.Git.CreateRef(ctx, orgName, repoName, &github.Reference{
		Ref:    &branchRefName,
		Object: &github.GitObject{SHA: &sourceBranchSHA},
	})
	if err != nil {
		return fmt.Errorf("Error creating GitHub branch reference %s/%s (%s): %s",
			orgName, repoName, branchRefName, err)
	}

	d.SetId(buildTwoPartID(repoName, branchName))

	return resourceGithubBranchRead(d, meta)
}

func resourceGithubBranchRead(d *schema.ResourceData, meta interface{}) error {
	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	if !d.IsNewResource() {
		ctx = context.WithValue(ctx, ctxEtag, d.Get("etag").(string))
	}

	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name
	repoName, branchName, err := parseTwoPartID(d.Id(), "repository", "branch")
	if err != nil {
		return err
	}
	branchRefName := "refs/heads/" + branchName

	log.Printf("[DEBUG] Querying GitHub branch reference %s/%s (%s)",
		orgName, repoName, branchRefName)
	ref, resp, err := client.Git.GetRef(ctx, orgName, repoName, branchRefName)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotModified {
				return nil
			}
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[WARN] Removing branch %s/%s (%s) from state because it no longer exists in Github",
					orgName, repoName, branchName)
				d.SetId("")
				return nil
			}
		}
		return fmt.Errorf("Error querying GitHub branch reference %s/%s (%s): %s",
			orgName, repoName, branchRefName, err)
	}

	d.SetId(buildTwoPartID(repoName, branchName))
	d.Set("etag", resp.Header.Get("ETag"))
	d.Set("repository", repoName)
	d.Set("branch", branchName)
	d.Set("ref", *ref.Ref)
	d.Set("sha", *ref.Object.SHA)

	return nil
}

func resourceGithubBranchDelete(d *schema.ResourceData, meta interface{}) error {
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name
	repoName, branchName, err := parseTwoPartID(d.Id(), "repository", "branch")
	if err != nil {
		return err
	}
	branchRefName := "refs/heads/" + branchName

	log.Printf("[DEBUG] Deleting GitHub branch reference %s/%s (%s)",
		orgName, repoName, branchRefName)
	_, err = client.Git.DeleteRef(ctx, orgName, repoName, branchRefName)
	if err != nil {
		return fmt.Errorf("Error deleting GitHub branch reference %s/%s (%s): %s",
			orgName, repoName, branchRefName, err)
	}

	return nil
}

func resourceGithubBranchImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	repoName, branchName, err := parseTwoPartID(d.Id(), "repository", "branch")
	if err != nil {
		return nil, err
	}

	sourceBranch := "main"
	if strings.Contains(branchName, ":") {
		branchName, sourceBranch, err = parseTwoPartID(branchName, "branch", "source_branch")
		if err != nil {
			return nil, err
		}
		d.SetId(buildTwoPartID(repoName, branchName))
	}

	d.Set("source_branch", sourceBranch)

	err = resourceGithubBranchRead(d, meta)
	if err != nil {
		return nil, err
	}

	// resourceGithubBranchRead calls d.SetId("") if the branch does not exist
	if d.Id() == "" {
		return nil, fmt.Errorf("Repository %s does not have a branch named %s.", repoName, branchName)
	}

	return []*schema.ResourceData{d}, nil
}
