package github

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/go-github/v38/github"
)

// checkRepositoryBranchExists tests if a branch exists in a repository.
func checkRepositoryBranchExists(client *github.Client, owner, repo, branch string) error {
	ctx := context.WithValue(context.Background(), ctxId, buildTwoPartID(repo, branch))
	_, _, err := client.Repositories.GetBranch(ctx, owner, repo, branch, true)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotFound {
				return fmt.Errorf("Branch %s not found in repository %s/%s or repository is not readable", branch, owner, repo)
			}
		}
		return err
	}

	return nil
}

// checkRepositoryFileExists tests if a file exists in a repository.
func checkRepositoryFileExists(client *github.Client, owner, repo, file, branch string) error {
	ctx := context.WithValue(context.Background(), ctxId, fmt.Sprintf("%s/%s", repo, file))
	fc, _, _, err := client.Repositories.GetContents(ctx, owner, repo, file, &github.RepositoryContentGetOptions{Ref: branch})
	if err != nil {
		return nil
	}
	if fc == nil {
		return fmt.Errorf("File %s not a file in in repository %s/%s or repository is not readable", file, owner, repo)
	}

	return nil
}

func getFileCommit(client *github.Client, owner, repo, file, branch string) (*github.RepositoryCommit, error) {
	ctx := context.WithValue(context.Background(), ctxId, fmt.Sprintf("%s/%s", repo, file))
	opts := &github.CommitsListOptions{
		SHA:  branch,
		Path: file,
	}
	allCommits := []*github.RepositoryCommit{}
	for {
		commits, resp, err := client.Repositories.ListCommits(ctx, owner, repo, opts)
		if err != nil {
			return nil, err
		}

		allCommits = append(allCommits, commits...)

		if resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage
	}

	for _, c := range allCommits {
		sha := c.GetSHA()

		// Skip merge commits
		if strings.Contains(c.Commit.GetMessage(), "Merge branch") {
			continue
		}

		rc, _, err := client.Repositories.GetCommit(ctx, owner, repo, sha, nil)
		if err != nil {
			return nil, err
		}

		for _, f := range rc.Files {
			if f.GetFilename() == file && f.GetStatus() != "removed" {
				log.Printf("[DEBUG] Found file: %s in commit: %s", file, sha)
				return rc, nil
			}
		}
	}

	return nil, fmt.Errorf("Cannot find file %s in repo %s/%s", file, owner, repo)
}
