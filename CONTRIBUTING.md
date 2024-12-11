# Contributing

Hi there! We're thrilled that you'd like to contribute to this project. Your help is essential to keep it great.

Contributions to this project are [released](https://help.github.com/articles/github-terms-of-service/#6-contributions-under-repository-license) to the public under the [project's open source license](LICENSE).

Please note that this project is released with a [Contributor Code of Conduct](CODE_OF_CONDUCT.md). By participating in this project you agree to abide by its terms.

Before submitting an issue or a pull request, please search the repository for existing content. Issues and PRs with multiple comments, reactions, and experience reports increase the likelihood of a merged change.

## Submitting a pull request

0. Fork and clone the repository.
0. Create a new branch: `git switch -c my-branch-name`.
0. Make your change, add tests, and make sure the tests still pass.
0. Push to your fork and submit a pull request.
0. Pat yourself on the back and wait for your pull request to be reviewed and merged.

Here are a few things you can do that will increase the likelihood of your pull request being accepted:

- Allow your pull request to receive edits by maintainers.
- Discuss your changes with the community in an issue.
- Write tests.
- Keep your change as focused as possible. If there are multiple changes you would like to make that are not dependent upon each other, please submit them as separate pull requests.
- Write a [good commit message](http://tbaggery.com/2008/04/19/a-note-about-git-commit-messages.html).

## Quick End-To-End Example

This section describes a typical sequence performed when developing locally. Full details of available tooling are available in the section on [Manual Testing](#manual-testing).

### Local Development Setup

Once you have the repository cloned, there's a couple of additional steps you'll need to take. Since most of the testing is acceptance or integration testing, we need to manipulate real GitHub resources in order to run it. Useful setup steps are listed below:

- If you haven't already, [create a GitHub organization you can use for testing](#github-organization).
  - Optional: you may find it beneficial to create a test user as well in order to avoid potential rate-limiting issues on your main account.
  - Your organization _must_ have a repository called `terraform-template-module`. The [terraformtesting/terraform-template-module](https://github.com/terraformtesting/terraform-template-module) repo is a good, re-usable example.
    - You _must_ make sure that the "Template Repository" item in Settings is checked for this repo.
- If you haven't already, generate a Personal Access Token (PAT) for authenticating your test runs.
- Export the necessary configuration for authenticating your provider with GitHub
  ```sh
  export GITHUB_TEST_AUTH_MODE="organization"
  export GITHUB_OWNER="<name of an organization>"
  export GITHUB_USERNAME="<username of the user who created the token>"
  export GITHUB_TOKEN="<token of a user with an organization account>"
  ```
- Build the project with `make build`
- Try an example test run from the default (`main`) branch, like `TF_LOG=DEBUG TF_ACC=1 go test -v ./... -run ^TestAccGithubRepositories`. All those tests should pass.

### Local Development Iteration

1. Write a test describing what you will fix. See [`github_label`](./github/resource_github_issue_label_test.go) for an example format.
1. Run your test and observe it fail. Enabling debug output allows for observing the underlying requests and responses made as well as viewing state (search `STATE:`) generated during the acceptance test run.
```sh
TF_LOG=DEBUG TF_ACC=1 go test -v ./... -run ^TestAccGithubIssueLabel
```
1. Align the resource's implementation to your test case and observe it pass:
```sh
TF_ACC=1 go test -v ./... -run ^TestAccGithubIssueLabel
```

Note that some resources still use a previous format that is incompatible with automated test runs, which depend on using the `skipUnlessMode` helper. When encountering these resources, tests should be rewritten to the latest format.

Also note that there is no build / `terraform init` / `terraform plan` sequence here.  It is uncommon to run into a bug or feature that requires iteration without using tests. When these cases arise, the `examples/` directory is used to approach the problem, which is detailed in the next section.

### Debugging the terraform provider

Println debugging can easily be used to obtain information about how code changes perform. If the `TF_LOG=DEBUG` level is set, calls to `log.Printf("[DEBUG] your message here")` will be printed in the program's output.

If a full debugger is desired, VSCode may be used. In order to do so,

0. Create a launch.json file with this configuration:
```json
{
	"name": "Attach to Process",
	"type": "go",
	"request": "attach",
	"mode": "local",
	"processId": 0,
}
```
Setting a `processId` of 0 allows a dropdown to select the process of the provider.

0. Add a sleep call (e.g. `time.Sleep(10 * time.Second)`) in the [`func providerConfigure(p *schema.Provider) schema.ConfigureFunc`](https://github.com/integrations/terraform-provider-github/blob/cec7e175c50bb091feecdc96ba117067c35ee351/github/provider.go#L274C1-L274C64) before the immediate `return` call. This will allow time to connect the debugger while the provider is initializing, before any critical logic happens.

0. Build the terraform provider with debug flags enabled and copy it to the appropriate bin folder with a command like `go build -gcflags="all=-N -l" -o ~/go/bin`.

0. Create or edit a `dev.tfrc` that points toward the newly-built binary, and export the `TF_CLI_CONFIG_FILE` variable to point to it. Further instructions on this process may be found in the [Building the provider](#using-a-local-version-of-the-provider) section.

0. Run a terraform command (e.g. `terraform apply`). While the provider pauses on initialization, go to VSCode and click "Attach to Process". In the search box that appears, type `terraform-provi` and select the terraform provider process.

0. The debugger is now connected! During a typical terraform command, the plugin will be invoked multiple times. If the debugger disconnects and the plugin is invoked again later in the run, the developer will have to re-attach each time as the process ID changes.


## Manual Testing

Manual testing should be performed on each PR opened in order to validate the provider's correct behavior and discover any regressions. Our automated testing is in an unhealthy spot at this point unfortunately, so extra care is required with manual testing. See [issue #1414](https://github.com/integrations/terraform-provider-github/issues/1414) for more details.

### Using a local version of the provider

Build the provider and specify the output directory:

```sh
$ go build -gcflags="all=-N -l" -o ~/go/bin
```

This enables verifying your locally built provider using examples available in the `examples/` directory.
Note that you will first need to configure your shell to map our provider to the local build:

```sh
export TF_CLI_CONFIG_FILE=path/to/project/examples/dev.tfrc
```

An example file is available in our `examples` directory and resembles:

```hcl
provider_installation {
  dev_overrides {
    "integrations/github" = "~/go/bin/"
  }

  direct {}
}
```

See https://www.terraform.io/docs/cli/config/config-file.html for more details.

When running examples, you should spot the following warning to confirm you are using a local build:

```console
Warning: Provider development overrides are in effect

The following provider development overrides are set in the CLI configuration:
 - integrations/github in /Users/jcudit/go/bin
```

### Environment variable reference

Commonly required environment variables are listed below:

```sh
# Enable debug logging
export TF_LOG=DEBUG

# Enables acceptance tests
export TF_ACC="1"

# Configure the URL override for GHES.
export GITHUB_BASE_URL=

# Configure acceptance testing mode; one of anonymous, individual, organization, team or enterprise. If not set will default to anonymous
export GITHUB_TEST_AUTH_MODE=

# Configure authentication for testing
export GITHUB_OWNER=
export GITHUB_USERNAME=
export GITHUB_TOKEN=

# Configure user level values
export GITHUB_TEST_USER_REPOSITORY=

# Configure for the org under test
export GITHUB_TEST_ORG_USER=
export GITHUB_TEST_ORG_REPOSITORY=
export GITHUB_TEST_ORG_TEMPLATE_REPOSITORY=
export GITHUB_TEST_ORG_APP_INSTALLATION_ID=

# Configure external (non-org) users
export GITHUB_TEST_EXTERNAL_USER=
export GITHUB_TEST_EXTERNAL_USER_TOKEN=
export GITHUB_TEST_EXTERNAL_USER2=

# Configure test options
export GITHUB_TEST_ADVANCED_SECURITY=
```

There are also a small amount of unit tests in the provider. Due to the nature of the provider, such tests are currently only recommended for exercising functionality completely internal to the provider. These may be executed by running `make test`.

### GitHub Organization

If you do not have an organization already that you are comfortable running tests against, you will need to [create one](https://help.github.com/en/articles/creating-a-new-organization-from-scratch). The free "Team for Open Source" org type is fine for these tests. The name of the organization must then be exported in your environment as `GITHUB_OWNER`.

Make sure that your organization has a `terraform-template-module` repository ([terraformtesting/terraform-template-module](https://github.com/terraformtesting/terraform-template-module) is an example you can clone) and that its "Template repository" item in Settings is checked.

If you are interested in using and/or testing GitHub's [Team synchronization](https://help.github.com/en/github/setting-up-and-managing-organizations-and-teams/synchronizing-teams-between-your-identity-provider-and-github) feature, please contact a maintainer as special arrangements can be made for your convenience.

### Example .vscode/launch.json file

This may come in handy when debugging both acceptance and manual testing.

```json
{
	// for information on how to debug the provider, see the CONTRIBUTING.md file
	"version": "0.2.0",
	"configurations": [
		{
			"name": "Launch test function",
			"type": "go",
			"request": "launch",
			"mode": "test",
			// note that the program file must be in the same package as the test to run,
			// though it does not necessarily have to be the file that contains the test.
			"program": "/home/kfcampbell/github/dev/terraform-provider-github/github/resource_github_team_members_test.go",
			"args": [
				"-test.v",
				"-test.run",
				"^TestAccGithubRepositoryTopics$" // ^ExactMatch$
			],
			"env": {
				"GITHUB_TEST_COLLABORATOR": "kfcampbell-terraform-test-user",
				"GITHUB_TEST_COLLABORATOR_TOKEN": "ghp_xxx",
				"GITHUB_TEST_USER": "kfcampbell",
				"GITHUB_TOKEN": "ghp_xxx",
				"GITHUB_TEMPLATE_REPOSITORY": "terraform-template-module",
				"GITHUB_TEMPLATE_REPOSITORY_RELEASE_ID": "12345678",
				// "GITHUB_OWNER": "kfcampbell-terraform-provider",
				// "GITHUB_OWNER": "kfcampbell",
				"GITHUB_ORGANIZATION": "kfcampbell-terraform-provider", // GITHUB_ORGANIZATION is required for organization integration tests
				"TF_CLI_CONFIG_FILE": "/home/kfcampbell/github/dev/terraform-provider-github/examples/dev.tfrc",
				"TF_ACC": "1",
				"TF_LOG": "DEBUG",
				"APP_INSTALLATION_ID": "12345678"
			}
		},
		{
			"name": "Attach to Process",
			"type": "go",
			"request": "attach",
			"mode": "local",
			"processId": 0
		}
	]
}
```
