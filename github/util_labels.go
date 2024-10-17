package github

import (
	"context"

	"github.com/google/go-github/v66/github"
)

func flattenLabels(labels []*github.Label) []interface{} {
	if labels == nil {
		return make([]interface{}, 0)
	}

	results := make([]interface{}, 0)

	for _, l := range labels {
		result := make(map[string]interface{})

		result["name"] = l.GetName()
		result["color"] = l.GetColor()
		result["description"] = l.GetDescription()
		result["url"] = l.GetURL()

		results = append(results, result)
	}

	return results
}

func listLabels(client *github.Client, ctx context.Context, owner, repository string) ([]*github.Label, error) {
	options := &github.ListOptions{
		PerPage: maxPerPage,
	}

	labels := make([]*github.Label, 0)

	for {
		ls, resp, err := client.Issues.ListLabels(ctx, owner, repository, options)
		if err != nil {
			return nil, err
		}

		labels = append(labels, ls...)

		if resp.NextPage == 0 {
			break
		}
		options.Page = resp.NextPage
	}

	return labels, nil
}
