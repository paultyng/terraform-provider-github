resource "github_issue_label" "individual_repo_label" {
  repository  = var.individual_repo
  name        = "foo"
  color       = "000000"
  description = "Test issue"
}

  resource "github_issue_label" "org_repo_label" {
  repository  = var.org_repo
  name        = "bar"
  color       = "000000"
  description = "Test issue"
  org         = var.org
}