package github

import (
	"context"
	"log"
	"strings"

	"fmt"

	"github.com/google/go-github/v38/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceGithubRepositoryFile() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubRepositoryFileCreate,
		Read:   resourceGithubRepositoryFileRead,
		Update: resourceGithubRepositoryFileUpdate,
		Delete: resourceGithubRepositoryFileDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				parts := strings.Split(d.Id(), ":")
				branch := "main"

				if len(parts) > 2 {
					return nil, fmt.Errorf("Invalid ID specified. Supplied ID must be written as <repository>/<file path> (when branch is \"main\") or <repository>/<file path>:<branch>")
				}

				if len(parts) == 2 {
					branch = parts[1]
				}

				client := meta.(*Owner).v3client
				owner := meta.(*Owner).name
				repo, file := splitRepoFilePath(parts[0])
				if err := checkRepositoryFileExists(client, owner, repo, file, branch); err != nil {
					return nil, err
				}

				d.SetId(fmt.Sprintf("%s/%s", repo, file))
				d.Set("branch", branch)
				d.Set("overwrite_on_create", false)

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The repository name",
			},
			"file": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The file path to manage",
			},
			"content": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The file's content",
			},
			"branch": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The branch name, defaults to \"main\"",
				Default:     "main",
			},
			"commit_sha": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The SHA of the commit that modified the file",
			},
			"commit_message": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The commit message when creating or updating the file",
			},
			"commit_author": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The commit author name, defaults to the authenticated user's name",
			},
			"commit_email": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The commit author email address, defaults to the authenticated user's email address",
			},
			"sha": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The blob SHA of the file",
			},
			"overwrite_on_create": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Enable overwriting existing files, defaults to \"false\"",
				Default:     false,
			},
		},
	}
}

func resourceGithubRepositoryFileOptions(d *schema.ResourceData) (*github.RepositoryContentFileOptions, error) {
	opts := &github.RepositoryContentFileOptions{
		Content: []byte(*github.String(d.Get("content").(string))),
		Branch:  github.String(d.Get("branch").(string)),
	}

	if commitMessage, hasCommitMessage := d.GetOk("commit_message"); hasCommitMessage {
		opts.Message = new(string)
		*opts.Message = commitMessage.(string)
	}

	if SHA, hasSHA := d.GetOk("sha"); hasSHA {
		opts.SHA = new(string)
		*opts.SHA = SHA.(string)
	}

	commitAuthor, hasCommitAuthor := d.GetOk("commit_author")
	commitEmail, hasCommitEmail := d.GetOk("commit_email")

	if hasCommitAuthor && !hasCommitEmail {
		return nil, fmt.Errorf("Cannot set commit_author without setting commit_email")
	}

	if hasCommitEmail && !hasCommitAuthor {
		return nil, fmt.Errorf("Cannot set commit_email without setting commit_author")
	}

	if hasCommitAuthor && hasCommitEmail {
		name := commitAuthor.(string)
		mail := commitEmail.(string)
		opts.Author = &github.CommitAuthor{Name: &name, Email: &mail}
		opts.Committer = &github.CommitAuthor{Name: &name, Email: &mail}
	}

	return opts, nil
}

func resourceGithubRepositoryFileCreate(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.Background()

	repo := d.Get("repository").(string)
	file := d.Get("file").(string)
	branch := d.Get("branch").(string)

	if err := checkRepositoryBranchExists(client, owner, repo, branch); err != nil {
		return err
	}

	opts, err := resourceGithubRepositoryFileOptions(d)
	if err != nil {
		return err
	}

	if opts.Message == nil {
		m := fmt.Sprintf("Add %s", file)
		opts.Message = &m
	}

	log.Printf("[DEBUG] Checking if overwriting a repository file: %s/%s/%s in branch: %s", owner, repo, file, branch)
	checkOpt := github.RepositoryContentGetOptions{Ref: branch}
	fileContent, _, resp, err := client.Repositories.GetContents(ctx, owner, repo, file, &checkOpt)
	if err != nil {
		if resp != nil {
			if resp.StatusCode != 404 {
				// 404 is a valid response if the file does not exist
				return err
			}
		} else {
			// Response should be non-nil
			return err
		}
	}

	if fileContent != nil {
		if d.Get("overwrite_on_create").(bool) {
			// Overwrite existing file if requested by configuring the options for
			// `client.Repositories.CreateFile` to match the existing file's SHA
			opts.SHA = fileContent.SHA
		} else {
			// Error if overwriting a file is not requested
			return fmt.Errorf("[ERROR] Refusing to overwrite existing file. Configure `overwrite_on_create` to `true` to override.")
		}
	}

	// Create a new or overwritten file
	log.Printf("[DEBUG] Creating repository file: %s/%s/%s in branch: %s", owner, repo, file, branch)
	create, _, err := client.Repositories.CreateFile(ctx, owner, repo, file, opts)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%s/%s", repo, file))
	d.Set("commit_sha", create.Commit.GetSHA())

	return resourceGithubRepositoryFileRead(d, meta)
}

func resourceGithubRepositoryFileRead(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	repo, file := splitRepoFilePath(d.Id())
	branch := d.Get("branch").(string)

	if err := checkRepositoryBranchExists(client, owner, repo, branch); err != nil {
		return err
	}

	log.Printf("[DEBUG] Reading repository file: %s/%s/%s, branch: %s", owner, repo, file, branch)
	opts := &github.RepositoryContentGetOptions{Ref: branch}
	fc, _, _, _ := client.Repositories.GetContents(ctx, owner, repo, file, opts)
	if fc == nil {
		log.Printf("[WARN] Removing repository path %s/%s/%s from state because it no longer exists in GitHub",
			owner, repo, file)
		d.SetId("")
		return nil
	}

	content, err := fc.GetContent()
	if err != nil {
		return err
	}

	d.Set("content", content)
	d.Set("repository", repo)
	d.Set("file", file)
	d.Set("sha", fc.GetSHA())

	log.Printf("[DEBUG] Fetching commit info for repository file: %s/%s/%s", owner, repo, file)
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

func resourceGithubRepositoryFileUpdate(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.Background()

	repo := d.Get("repository").(string)
	file := d.Get("file").(string)
	branch := d.Get("branch").(string)

	if err := checkRepositoryBranchExists(client, owner, repo, branch); err != nil {
		return err
	}

	opts, err := resourceGithubRepositoryFileOptions(d)
	if err != nil {
		return err
	}

	if *opts.Message == fmt.Sprintf("Add %s", file) {
		m := fmt.Sprintf("Update %s", file)
		opts.Message = &m
	}

	log.Printf("[DEBUG] Updating content in repository file: %s/%s/%s", owner, repo, file)
	create, _, err := client.Repositories.CreateFile(ctx, owner, repo, file, opts)
	if err != nil {
		return err
	}

	d.Set("commit_sha", create.GetSHA())

	return resourceGithubRepositoryFileRead(d, meta)
}

func resourceGithubRepositoryFileDelete(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.Background()

	repo := d.Get("repository").(string)
	file := d.Get("file").(string)
	branch := d.Get("branch").(string)

	message := fmt.Sprintf("Delete %s", file)
	sha := d.Get("sha").(string)
	opts := &github.RepositoryContentFileOptions{
		Message: &message,
		SHA:     &sha,
		Branch:  &branch,
	}

	log.Printf("[DEBUG] Deleting repository file: %s/%s/%s", owner, repo, file)
	_, _, err := client.Repositories.DeleteFile(ctx, owner, repo, file, opts)
	if err != nil {
		return nil
	}

	return nil
}
