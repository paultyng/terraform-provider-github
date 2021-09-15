package github

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/google/go-github/v39/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceGithubRepositoryPullRequest() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubRepositoryPullRequestCreate,
		Read:   resourceGithubRepositoryPullRequestRead,
		Update: resourceGithubRepositoryPullRequestUpdate,
		Delete: resourceGithubRepositoryPullRequestDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				_, baseRepository, _, err := parsePullRequestID(d)
				if err != nil {
					return nil, err
				}
				d.Set("base_repository", baseRepository)

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"owner": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"base_repository": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"base_ref": {
				Type:     schema.TypeString,
				Required: true,
			},
			"head_ref": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"title": {
				// Even though the documentation does not explicitly mark the
				// title field as required, attempts to create a PR with an
				// empty title result in a "missing_field" validation error
				// (HTTP 422).
				Type:     schema.TypeString,
				Required: true,
			},
			"body": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"maintainer_can_modify": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"base_sha": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"draft": {
				// The "draft" field is an interesting corner case because while
				// you can create a draft PR through the API, the documentation
				// does not indicate that you can change this field during
				// update:
				//
				// https://docs.github.com/en/rest/reference/pulls#update-a-pull-request
				//
				// And since you cannot manage the lifecycle of this field to
				// reconcile the actual state with the desired one, this field
				// cannot be managed by Terraform.
				Type:     schema.TypeBool,
				Computed: true,
			},
			"head_sha": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"labels": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				Description: "List of names of labels on the PR",
			},
			"number": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"opened_at": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"opened_by": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Username of the PR creator",
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceGithubRepositoryPullRequestCreate(d *schema.ResourceData, meta interface{}) error {
	ctx := context.TODO()
	client := meta.(*Owner).v3client

	// For convenience, by default we expect that the base repository and head
	// repository owners are the same, and both belong to the caller, indicating
	// a "PR within the same repo" scenario. The head will *always* belong to
	// the current caller, the base - not necessarily. The base will belong to
	// another namespace in case of forks, and this resource supports them.
	headOwner := meta.(*Owner).name

	baseOwner := headOwner
	if explicitBaseOwner, ok := d.GetOk("owner"); ok {
		baseOwner = explicitBaseOwner.(string)
	}

	baseRepository := d.Get("base_repository").(string)

	head := d.Get("head_ref").(string)
	if headOwner != baseOwner {
		head = strings.Join([]string{headOwner, head}, ":")
	}

	pullRequest, _, err := client.PullRequests.Create(ctx, baseOwner, baseRepository, &github.NewPullRequest{
		Title:               github.String(d.Get("title").(string)),
		Head:                github.String(head),
		Base:                github.String(d.Get("base_ref").(string)),
		Body:                github.String(d.Get("body").(string)),
		MaintainerCanModify: github.Bool(d.Get("maintainer_can_modify").(bool)),
	})

	if err != nil {
		return err
	}

	d.SetId(buildThreePartID(baseOwner, baseRepository, strconv.Itoa(pullRequest.GetNumber())))

	return resourceGithubRepositoryPullRequestRead(d, meta)
}

func resourceGithubRepositoryPullRequestRead(d *schema.ResourceData, meta interface{}) error {
	ctx := context.TODO()
	client := meta.(*Owner).v3client

	owner, repository, number, err := parsePullRequestID(d)
	if err != nil {
		return err
	}

	pullRequest, _, err := client.PullRequests.Get(ctx, owner, repository, number)
	if err != nil {
		return err
	}

	d.Set("number", pullRequest.GetNumber())

	if head := pullRequest.GetHead(); head != nil {
		d.Set("head_ref", head.GetRef())
		d.Set("head_sha", head.GetSHA())
	} else {
		// Totally unexpected condition. Better do that than segfault, I guess?
		log.Printf("[WARN] Head branch missing, expected %s", d.Get("head_ref"))
		d.SetId("")
		return nil
	}

	if base := pullRequest.GetBase(); base != nil {
		d.Set("base_ref", base.GetRef())
		d.Set("base_sha", base.GetSHA())
	} else {
		// Seme logic as with the missing head branch.
		log.Printf("[WARN] Base branch missing, expected %s", d.Get("base_ref"))
		d.SetId("")
		return nil
	}

	d.Set("body", pullRequest.GetBody())
	d.Set("title", pullRequest.GetTitle())
	d.Set("draft", pullRequest.GetDraft())
	d.Set("maintainer_can_modify", pullRequest.GetMaintainerCanModify())
	d.Set("number", pullRequest.GetNumber())
	d.Set("state", pullRequest.GetState())
	d.Set("opened_at", pullRequest.GetCreatedAt().Unix())
	d.Set("updated_at", pullRequest.GetUpdatedAt().Unix())

	if user := pullRequest.GetUser(); user != nil {
		d.Set("opened_by", user.GetLogin())
	}

	labels := []string{}
	for _, label := range pullRequest.Labels {
		labels = append(labels, label.GetName())
	}
	d.Set("labels", labels)

	return nil
}

func resourceGithubRepositoryPullRequestUpdate(d *schema.ResourceData, meta interface{}) error {
	ctx := context.TODO()
	client := meta.(*Owner).v3client

	owner, repository, number, err := parsePullRequestID(d)
	if err != nil {
		return err
	}

	update := &github.PullRequest{
		Title:               github.String(d.Get("title").(string)),
		Body:                github.String(d.Get("body").(string)),
		MaintainerCanModify: github.Bool(d.Get("maintainer_can_modify").(bool)),
	}

	if d.HasChange("base_ref") {
		update.Base = &github.PullRequestBranch{
			Ref: github.String(d.Get("base_ref").(string)),
		}
	}

	_, _, err = client.PullRequests.Edit(ctx, owner, repository, number, update)
	if err == nil {
		return resourceGithubRepositoryPullRequestRead(d, meta)
	}

	errors := []string{fmt.Sprintf("could not update the Pull Request: %v", err)}

	if err := resourceGithubRepositoryPullRequestRead(d, meta); err != nil {
		errors = append(errors, fmt.Sprintf("could not read the Pull Request after the failed update: %v", err))
	}

	return fmt.Errorf(strings.Join(errors, ", "))
}

func resourceGithubRepositoryPullRequestDelete(d *schema.ResourceData, meta interface{}) error {
	// It's not entirely clear how to treat PR deletion according to Terraform's
	// CRUD semantics. The approach we're taking here is to close the PR unless
	// it's already closed or merged. Merging it feels intuitively wrong in what
	// effectively is a destructor.
	if d.Get("state").(string) != "open" {
		d.SetId("")
		return nil
	}

	ctx := context.TODO()
	client := meta.(*Owner).v3client

	owner, repository, number, err := parsePullRequestID(d)
	if err != nil {
		return err
	}

	update := &github.PullRequest{State: github.String("closed")}
	if _, _, err = client.PullRequests.Edit(ctx, owner, repository, number, update); err != nil {
		return err
	}

	d.SetId("")
	return nil
}

func parsePullRequestID(d *schema.ResourceData) (owner, repository string, number int, err error) {
	var strNumber string

	if owner, repository, strNumber, err = parseThreePartID(d.Id(), "owner", "base_repository", "number"); err != nil {
		return
	}

	if number, err = strconv.Atoi(strNumber); err != nil {
		err = fmt.Errorf("invalid PR number %s: %w", strNumber, err)
	}

	return
}
