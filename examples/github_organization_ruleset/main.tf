resource "github_organization_ruleset" "protect_branch" {
  enforcement = "active"
  name        = "Protect Master Branch (Beta)"
  target      = "branch"
  conditions {
    repository_property {
      exclude = []
      include = [
        {
          name  = "visibility"
          value = "public"
        }
      ]
    }
    ref_name {
      exclude = []
      include = [
        "~DEFAULT_BRANCH",
      ]
    }
  }
  rules {
    creation                = false
    deletion                = true
    non_fast_forward        = true
    required_linear_history = false
    required_signatures     = false
    update                  = false

    pull_request {
      dismiss_stale_reviews_on_push     = true
      require_code_owner_review         = true
      require_last_push_approval        = true
      required_approving_review_count   = 1
      required_review_thread_resolution = false
    }

    required_status_checks {
      strict_required_status_checks_policy = false

      required_check {
        context        = "SonarQube Code Analysis"
        integration_id = 0
      }
    }

    required_workflows {
      required_workflow {
        repository_id = "442515790"
        path          = ".github/workflows/trivy-scan-pr-commit.yml"
        ref           = "refs/heads/master"
      }
      required_workflow {
        repository_id = "442515790"
        path          = ".github/workflows/required_neodora.yml"
        ref           = "refs/heads/master"
      }
    }

  }
}
