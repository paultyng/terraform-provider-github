package github

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/go-github/v57/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	codeQLWorkflowRunFailure  = "codeql setup workflow failed for repository"
	codeQLWorkflowRunInFlight = "codeql setup for repository still in progress"
)

type DefaultSetupConfigurationResponse struct {
	RunId  int64  `json:"run_id"`
	RunUrl string `json:"run_url"`
}

func resourceGithubRepositoryCodeScanning() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubRepositoryCodeScanningCreate,
		Read:   resourceGithubRepositoryCodeScanningRead,
		Update: resourceGithubRepositoryCodeScanningUpdate,
		Delete: resourceGithubRepositoryCodeScanningDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The GitHub repository",
			},
			"owner": {
				Type:     schema.TypeString,
				Required: true,
			},
			"languages": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"query_suite": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"default",
					"extended",
				}, false),
			},
			"state": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"configured",
					"not-configured",
				}, false),
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"wait": {
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
			},
		},
	}
}

func resourceGithubRepositoryCodeScanningCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	owner := meta.(*Owner).name
	repoName := d.Get("repository").(string)

	createUpdateOpts := createUpdateCodeScanning(d, meta)
	ctx := context.Background()

	_, _, err := client.CodeScanning.UpdateDefaultSetupConfiguration(ctx,
		owner,
		repoName,
		&createUpdateOpts,
	)

	if err != nil {
		_, ok := err.(*github.AcceptedError)
		if !ok {
			return err
		}
	}

	err = retry.RetryContext(ctx, 3*time.Second, func() *retry.RetryError {
		conf, _, err := client.CodeScanning.GetDefaultSetupConfiguration(ctx, owner, repoName)
		if err != nil {
			return retry.NonRetryableError(err)
		}

		if *conf.State == "not-configured" {
			return retry.RetryableError(errors.New("not configured yet"))
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("Error waiting for default setup configuration (%s) to be configured: %s", d.Id(), err)
	}

	d.SetId(buildTwoPartID(owner, repoName))

	return resourceGithubRepositoryCodeScanningRead(d, meta)
}

func resourceGithubRepositoryCodeScanningDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	owner, repoName, err := parseTwoPartID(d.Id(), "owner", "repository")
	if err != nil {
		return err
	}

	createUpdateOpts := createUpdateCodeScanning(d, meta)
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	_, _, err = client.CodeScanning.UpdateDefaultSetupConfiguration(ctx, owner, repoName, &createUpdateOpts)
	return err
}

func resourceGithubRepositoryCodeScanningRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	owner, repoName, err := parseTwoPartID(d.Id(), "owner", "repository")
	if err != nil {
		return err
	}

	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	config, _, err := client.CodeScanning.GetDefaultSetupConfiguration(ctx, owner, repoName)
	if err != nil {
		return err
	}

	timeString := ""

	if config.UpdatedAt != nil {
		timeString = config.UpdatedAt.String()
	}

	d.Set("repository", repoName)
	d.Set("owner", owner)
	d.Set("state", config.GetState())
	d.Set("query_suite", config.GetQuerySuite())
	d.Set("languages", config.Languages)
	d.Set("updated_at", timeString)

	return nil
}

func resourceGithubRepositoryCodeScanningUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	owner, repoName, err := parseTwoPartID(d.Id(), "owner", "repository")
	if err != nil {
		return err
	}

	createUpdateOpts := createUpdateCodeScanning(d, meta)
	ctx := context.Background()

	_, _, err = client.CodeScanning.UpdateDefaultSetupConfiguration(ctx,
		owner,
		repoName,
		&createUpdateOpts,
	)
	if err != nil {
		return err
	}

	d.SetId(buildTwoPartID(owner, repoName))

	return resourceGithubRepositoryCodeScanningRead(d, meta)
}

func createUpdateCodeScanning(d *schema.ResourceData, meta interface{}) github.UpdateDefaultSetupConfigurationOptions {
	data := github.UpdateDefaultSetupConfigurationOptions{}

	if v, ok := d.GetOk("query_suite"); ok {
		querySuite := v.(string)
		data.QuerySuite = &querySuite
	}

	data.State = d.Get("state").(string)

	return data
}

func waitForCodeQLActionCompleteFunc(ctx context.Context, client *github.Client, resourceId string, runId int64) resource.RetryFunc {
	return func() *resource.RetryError {
		owner, repoName, err := parseTwoPartID(resourceId, "owner", "repository")
		if err != nil {
			return resource.NonRetryableError(err)
		}

		workflowRun, _, err := client.Actions.GetWorkflowRunByID(ctx, owner, repoName, runId)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		switch *workflowRun.Status {
		case "success":
			return nil
		case "failure", "timed out", "cancelled":
			return resource.NonRetryableError(errors.New(codeQLWorkflowRunFailure))
		case "queued", "in progress", "waiting":
			return resource.RetryableError(errors.New(codeQLWorkflowRunInFlight))
		}

		return nil
	}
}
