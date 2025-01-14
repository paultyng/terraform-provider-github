package github

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"

	"fmt"

	"github.com/google/go-github/v66/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

				if len(parts) > 2 {
					return nil, fmt.Errorf("invalid ID specified. Supplied ID must be written as <repository>/<file path> (when branch is \"main\") or <repository>/<file path>:<branch>")
				}

				client := meta.(*Owner).v3client
				owner := meta.(*Owner).name
				repo, file := splitRepoFilePath(parts[0])
				// test if a file exists in a repository.
				ctx := context.WithValue(context.Background(), ctxId, fmt.Sprintf("%s/%s", repo, file))
				opts := &github.RepositoryContentGetOptions{}
				if len(parts) == 2 {
					opts.Ref = parts[1]
					if err := d.Set("branch", parts[1]); err != nil {
						return nil, err
					}
				}
				fc, _, _, err := client.Repositories.GetContents(ctx, owner, repo, file, opts)
				if err != nil {
					return nil, err
				}
				if fc == nil {
					return nil, fmt.Errorf("file %s is not a file in repository %s/%s or repository is not readable", file, owner, repo)
				}

				d.SetId(fmt.Sprintf("%s/%s", repo, file))
				if err = d.Set("overwrite_on_create", false); err != nil {
					return nil, err
				}

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
				Description: "The branch name, defaults to the repository's default branch",
			},
			"ref": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the commit/branch/tag",
				ForceNew:    true,
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
				Description: "The commit message when creating, updating or deleting the file",
			},
			"commit_author": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    false,
				Description: "The commit author name, defaults to the authenticated user's name. GitHub app users may omit author and email information so GitHub can verify commits as the GitHub App. ",
			},
			"commit_email": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    false,
				Description: "The commit author email address, defaults to the authenticated user's email address. GitHub app users may omit author and email information so GitHub can verify commits as the GitHub App.",
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
			"autocreate_branch": {
				Type:             schema.TypeBool,
				Optional:         true,
				Description:      "Automatically create the branch if it could not be found. Subsequent reads if the branch is deleted will occur from 'autocreate_branch_source_branch'",
				Default:          false,
				DiffSuppressFunc: autoBranchDiffSuppressFunc,
			},
			"autocreate_branch_source_branch": {
				Type:             schema.TypeString,
				Default:          "main",
				Optional:         true,
				Description:      "The branch name to start from, if 'autocreate_branch' is set. Defaults to 'main'.",
				RequiredWith:     []string{"autocreate_branch"},
				DiffSuppressFunc: autoBranchDiffSuppressFunc,
			},
			"autocreate_branch_source_sha": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Description:      "The commit hash to start from, if 'autocreate_branch' is set. Defaults to the tip of 'autocreate_branch_source_branch'. If provided, 'autocreate_branch_source_branch' is ignored.",
				RequiredWith:     []string{"autocreate_branch"},
				DiffSuppressFunc: autoBranchDiffSuppressFunc,
			},
		},
	}
}

func resourceGithubRepositoryFileOptions(d *schema.ResourceData) (*github.RepositoryContentFileOptions, error) {
	opts := &github.RepositoryContentFileOptions{
		Content: []byte(*github.String(d.Get("content").(string))),
	}

	if branch, ok := d.GetOk("branch"); ok {
		opts.Branch = github.String(branch.(string))
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
		return nil, fmt.Errorf("cannot set commit_author without setting commit_email")
	}

	if hasCommitEmail && !hasCommitAuthor {
		return nil, fmt.Errorf("cannot set commit_email without setting commit_author")
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

	checkOpt := github.RepositoryContentGetOptions{}

	if branch, ok := d.GetOk("branch"); ok {
		log.Printf("[DEBUG] Using explicitly set branch: %s", branch.(string))
		if err := checkRepositoryBranchExists(client, owner, repo, branch.(string)); err != nil {
			if d.Get("autocreate_branch").(bool) {
				branchRefName := "refs/heads/" + branch.(string)
				sourceBranchName := d.Get("autocreate_branch_source_branch").(string)
				sourceBranchRefName := "refs/heads/" + sourceBranchName

				if _, hasSourceSHA := d.GetOk("autocreate_branch_source_sha"); !hasSourceSHA {
					ref, _, err := client.Git.GetRef(ctx, owner, repo, sourceBranchRefName)
					if err != nil {
						return fmt.Errorf("error querying GitHub branch reference %s/%s (%s): %s",
							owner, repo, sourceBranchRefName, err)
					}
					d.Set("autocreate_branch_source_sha", *ref.Object.SHA)
				}
				sourceBranchSHA := d.Get("autocreate_branch_source_sha").(string)
				if _, _, err := client.Git.CreateRef(ctx, owner, repo, &github.Reference{
					Ref:    &branchRefName,
					Object: &github.GitObject{SHA: &sourceBranchSHA},
				}); err != nil {
					return err
				}
			} else {
				return err
			}
		}
		checkOpt.Ref = branch.(string)
	}

	opts, err := resourceGithubRepositoryFileOptions(d)
	if err != nil {
		return err
	}

	if opts.Message == nil {
		m := fmt.Sprintf("Add %s", file)
		opts.Message = &m
	}

	log.Printf("[DEBUG] Checking if overwriting a repository file: %s/%s/%s", owner, repo, file)
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
			return fmt.Errorf("refusing to overwrite existing file: configure `overwrite_on_create` to `true` to override")
		}
	}

	// Create a new or overwritten file
	create, _, err := client.Repositories.CreateFile(ctx, owner, repo, file, opts)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%s/%s", repo, file))
	if err = d.Set("commit_sha", create.Commit.GetSHA()); err != nil {
		return err
	}

	return resourceGithubRepositoryFileRead(d, meta)
}

func resourceGithubRepositoryFileRead(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	repo, file := splitRepoFilePath(d.Id())

	opts := &github.RepositoryContentGetOptions{}

	if branch, ok := d.GetOk("branch"); ok {
		log.Printf("[DEBUG] Using explicitly set branch: %s", branch.(string))
		if err := checkRepositoryBranchExists(client, owner, repo, branch.(string)); err != nil {
			if d.Get("autocreate_branch").(bool) {
				branch = d.Get("autocreate_branch_source_branch").(string)
			} else {
				log.Printf("[INFO] Removing repository path %s/%s/%s from state because the branch no longer exists in GitHub",
					owner, repo, file)
				d.SetId("")
				return nil
			}
		}
		opts.Ref = branch.(string)
	}

	fc, _, _, err := client.Repositories.GetContents(ctx, owner, repo, file, opts)
	if err != nil {
		var errorResponse *github.ErrorResponse
		if errors.As(err, &errorResponse) && errorResponse.Response.StatusCode == http.StatusTooManyRequests {
			return err
		}
	}
	if fc == nil {
		log.Printf("[INFO] Removing repository path %s/%s/%s from state because it no longer exists in GitHub",
			owner, repo, file)
		d.SetId("")
		return nil
	}

	content, err := fc.GetContent()
	if err != nil {
		return err
	}

	if err = d.Set("content", content); err != nil {
		return err
	}
	if err = d.Set("repository", repo); err != nil {
		return err
	}
	if err = d.Set("file", file); err != nil {
		return err
	}
	if err = d.Set("sha", fc.GetSHA()); err != nil {
		return err
	}

	var commit *github.RepositoryCommit

	parsedUrl, err := url.Parse(fc.GetURL())
	if err != nil {
		return err
	}
	parsedQuery, err := url.ParseQuery(parsedUrl.RawQuery)
	if err != nil {
		return err
	}
	ref := parsedQuery["ref"][0]
	if err = d.Set("ref", ref); err != nil {
		return err
	}

	// Use the SHA to lookup the commit info if we know it, otherwise loop through commits
	if sha, ok := d.GetOk("commit_sha"); ok {
		log.Printf("[DEBUG] Using known commit SHA: %s", sha.(string))
		commit, _, err = client.Repositories.GetCommit(ctx, owner, repo, sha.(string), nil)
	} else {
		log.Printf("[DEBUG] Commit SHA unknown for file: %s/%s/%s, looking for commit...", owner, repo, file)
		commit, err = getFileCommit(client, owner, repo, file, ref)
		log.Printf("[DEBUG] Found file: %s/%s/%s, in commit SHA: %s ", owner, repo, file, commit.GetSHA())
	}
	if err != nil {
		return err
	}

	if err = d.Set("commit_sha", commit.GetSHA()); err != nil {
		return err
	}

	commit_author := commit.Commit.GetCommitter().GetName()
	commit_email := commit.Commit.GetCommitter().GetEmail()

	_, hasCommitAuthor := d.GetOk("commit_author")
	_, hasCommitEmail := d.GetOk("commit_email")

	//read from state if author+email is set explicitly, and if it was not github signing it for you previously
	if commit_author != "GitHub" && commit_email != "noreply@github.com" && hasCommitAuthor && hasCommitEmail {
		if err = d.Set("commit_author", commit_author); err != nil {
			return err
		}
		if err = d.Set("commit_email", commit_email); err != nil {
			return err
		}
	}
	if err = d.Set("commit_message", commit.GetCommit().GetMessage()); err != nil {
		return err
	}

	return nil
}

func resourceGithubRepositoryFileUpdate(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.Background()

	repo := d.Get("repository").(string)
	file := d.Get("file").(string)

	if branch, ok := d.GetOk("branch"); ok {
		log.Printf("[DEBUG] Using explicitly set branch: %s", branch.(string))
		if err := checkRepositoryBranchExists(client, owner, repo, branch.(string)); err != nil {
			if d.Get("autocreate_branch").(bool) {
				branchRefName := "refs/heads/" + branch.(string)
				sourceBranchName := d.Get("autocreate_branch_source_branch").(string)
				sourceBranchRefName := "refs/heads/" + sourceBranchName

				if _, hasSourceSHA := d.GetOk("autocreate_branch_source_sha"); !hasSourceSHA {
					ref, _, err := client.Git.GetRef(ctx, owner, repo, sourceBranchRefName)
					if err != nil {
						return fmt.Errorf("error querying GitHub branch reference %s/%s (%s): %s",
							owner, repo, sourceBranchRefName, err)
					}
					d.Set("autocreate_branch_source_sha", *ref.Object.SHA)
				}
				sourceBranchSHA := d.Get("autocreate_branch_source_sha").(string)
				if _, _, err := client.Git.CreateRef(ctx, owner, repo, &github.Reference{
					Ref:    &branchRefName,
					Object: &github.GitObject{SHA: &sourceBranchSHA},
				}); err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	opts, err := resourceGithubRepositoryFileOptions(d)
	if err != nil {
		return err
	}

	if *opts.Message == fmt.Sprintf("Add %s", file) {
		m := fmt.Sprintf("Update %s", file)
		opts.Message = &m
	}

	create, _, err := client.Repositories.CreateFile(ctx, owner, repo, file, opts)
	if err != nil {
		return err
	}

	if err = d.Set("commit_sha", create.GetSHA()); err != nil {
		return err
	}

	return resourceGithubRepositoryFileRead(d, meta)
}

func resourceGithubRepositoryFileDelete(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.Background()

	repo := d.Get("repository").(string)
	file := d.Get("file").(string)

	var branch string

	message := fmt.Sprintf("Delete %s", file)

	if commitMessage, hasCommitMessage := d.GetOk("commit_message"); hasCommitMessage {
		message = commitMessage.(string)
	}

	sha := d.Get("sha").(string)
	opts := &github.RepositoryContentFileOptions{
		Message: &message,
		SHA:     &sha,
	}

	if b, ok := d.GetOk("branch"); ok {
		log.Printf("[DEBUG] Using explicitly set branch: %s", b.(string))
		if err := checkRepositoryBranchExists(client, owner, repo, b.(string)); err != nil {
			if d.Get("autocreate_branch").(bool) {
				branchRefName := "refs/heads/" + b.(string)
				sourceBranchName := d.Get("autocreate_branch_source_branch").(string)
				sourceBranchRefName := "refs/heads/" + sourceBranchName

				if _, hasSourceSHA := d.GetOk("autocreate_branch_source_sha"); !hasSourceSHA {
					ref, _, err := client.Git.GetRef(ctx, owner, repo, sourceBranchRefName)
					if err != nil {
						return fmt.Errorf("error querying GitHub branch reference %s/%s (%s): %s",
							owner, repo, sourceBranchRefName, err)
					}
					d.Set("autocreate_branch_source_sha", *ref.Object.SHA)
				}
				sourceBranchSHA := d.Get("autocreate_branch_source_sha").(string)
				if _, _, err := client.Git.CreateRef(ctx, owner, repo, &github.Reference{
					Ref:    &branchRefName,
					Object: &github.GitObject{SHA: &sourceBranchSHA},
				}); err != nil {
					return err
				}
			} else {
				return err
			}
		}
		branch = b.(string)
		opts.Branch = &branch
	}

	_, _, err := client.Repositories.DeleteFile(ctx, owner, repo, file, opts)
	return err
}

func autoBranchDiffSuppressFunc(k, _, _ string, d *schema.ResourceData) bool {
	if !d.Get("autocreate_branch").(bool) {
		switch k {
		case "autocreate_branch", "autocreate_branch_source_branch", "autocreate_branch_source_sha":
			return true
		}
	}
	return false
}
