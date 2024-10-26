package github

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The GitHub repository name.",
			},
			"branch": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The repository branch to create.",
			},
			"source_branch": {
				Type:        schema.TypeString,
				Default:     "main",
				Optional:    true,
				ForceNew:    true,
				Description: "The branch name to start from. Defaults to 'main'.",
			},
			"source_sha": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
				Description: "The commit hash to start from. Defaults to the tip of 'source_branch'. If provided, 'source_branch' is ignored.",
			},
			"etag": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "An etag representing the Branch object.",
			},
			"ref": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "A string representing a branch reference, in the form of 'refs/heads/<branch>'.",
			},
			"sha": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "A string storing the reference's HEAD commit's SHA1.",
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
		ref, _, err := client.Git.GetRef(ctx, orgName, repoName, sourceBranchRefName)
		if err != nil {
			return fmt.Errorf("error querying GitHub branch reference %s/%s (%s): %s",
				orgName, repoName, sourceBranchRefName, err)
		}
		if err = d.Set("source_sha", *ref.Object.SHA); err != nil {
			return err
		}
	}
	sourceBranchSHA := d.Get("source_sha").(string)

	_, _, err := client.Git.CreateRef(ctx, orgName, repoName, &github.Reference{
		Ref:    &branchRefName,
		Object: &github.GitObject{SHA: &sourceBranchSHA},
	})
	// If the branch already exists, rather than erroring out just continue on to importing the branch
	//   This avoids the case where a repo with gitignore_template and branch are being created at the same time crashing terraform
	if err != nil && !strings.HasSuffix(err.Error(), "422 Reference already exists []") {
		return fmt.Errorf("error creating GitHub branch reference %s/%s (%s): %s",
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

	ref, resp, err := client.Git.GetRef(ctx, orgName, repoName, branchRefName)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotModified {
				return nil
			}
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[INFO] Removing branch %s/%s (%s) from state because it no longer exists in GitHub",
					orgName, repoName, branchName)
				d.SetId("")
				return nil
			}
		}
		return fmt.Errorf("error querying GitHub branch reference %s/%s (%s): %s",
			orgName, repoName, branchRefName, err)
	}

	d.SetId(buildTwoPartID(repoName, branchName))
	if err = d.Set("etag", resp.Header.Get("ETag")); err != nil {
		return err
	}
	if err = d.Set("repository", repoName); err != nil {
		return err
	}
	if err = d.Set("branch", branchName); err != nil {
		return err
	}
	if err = d.Set("ref", *ref.Ref); err != nil {
		return err
	}
	if err = d.Set("sha", *ref.Object.SHA); err != nil {
		return err
	}

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

	_, err = client.Git.DeleteRef(ctx, orgName, repoName, branchRefName)
	if err != nil {
		return fmt.Errorf("error deleting GitHub branch reference %s/%s (%s): %s",
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

	if err = d.Set("source_branch", sourceBranch); err != nil {
		return nil, err
	}

	err = resourceGithubBranchRead(d, meta)
	if err != nil {
		return nil, err
	}

	// resourceGithubBranchRead calls d.SetId("") if the branch does not exist
	if d.Id() == "" {
		return nil, fmt.Errorf("repository %s does not have a branch named %s", repoName, branchName)
	}

	return []*schema.ResourceData{d}, nil
}
