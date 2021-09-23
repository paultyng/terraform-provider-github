package github

import (
	"context"
	"fmt"
	"log"

	"github.com/google/go-github/v38/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceGithubRepositoryFile() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGithubRepositoryFileRead,
		Schema: map[string]*schema.Schema{
			"repository": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The repository name",
			},
			"file": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The file path to manage",
			},
			"branch": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The branch name, defaults to \"main\"",
				Default:     "main",
			},
			"content": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The file's content",
			},
			"commit_sha": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The SHA of the commit that modified the file",
			},
			"commit_message": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The commit message when creating or updating the file",
			},
			"commit_author": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The commit author name, defaults to the authenticated user's name",
			},
			"commit_email": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The commit author email address, defaults to the authenticated user's email address",
			},
			"sha": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The blob SHA of the file",
			},
		},
	}
}

func dataSourceGithubRepositoryFileRead(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	repo := d.Get("repository").(string)
	file := d.Get("file").(string)
	branch := d.Get("branch").(string)
	if err := checkRepositoryBranchExists(client, owner, repo, branch); err != nil {
		return err
	}

	log.Printf("[DEBUG] Data Source reading repository file: %s/%s/%s, branch: %s", owner, repo, file, branch)
	opts := &github.RepositoryContentGetOptions{Ref: branch}
	fc, _, _, err := client.Repositories.GetContents(ctx, owner, repo, file, opts)
	if err != nil {
		return err
	}

	content, err := fc.GetContent()
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%s/%s", repo, file))
	d.Set("content", content)
	d.Set("repository", repo)
	d.Set("file", file)
	d.Set("sha", fc.GetSHA())

	log.Printf("[DEBUG] Data Source fetching commit info for repository file: %s/%s/%s", owner, repo, file)
	var commit *github.RepositoryCommit

	// Use the SHA to lookup the commit info if we know it, otherwise loop through commits
	if sha, ok := d.GetOk("commit_sha"); ok {
		log.Printf("[DEBUG] Using known commit SHA: %s", sha.(string))
		commit, _, err = client.Repositories.GetCommit(ctx, owner, repo, sha.(string), nil)
	} else {
		log.Printf("[DEBUG] Commit SHA unknown for file: %s/%s/%s, looking for commit...", owner, repo, file)
		commit, err = getFileCommit(client, owner, repo, file, branch)
		log.Printf("[DEBUG] Found file: %s/%s/%s, in commit SHA: %s ", owner, repo, file, commit.GetSHA())
	}
	if err != nil {
		return err
	}

	d.Set("commit_sha", commit.GetSHA())
	d.Set("commit_author", commit.Commit.GetCommitter().GetName())
	d.Set("commit_email", commit.Commit.GetCommitter().GetEmail())
	d.Set("commit_message", commit.GetCommit().GetMessage())

	return nil
}
