provider "github" {
}

terraform {
  required_providers {
    github = {
      source  = "integrations/github"
    }
  }
}

resource "github_actions_organization_secret" "plaintext_secret" {
  secret_name      = "test_plaintext_secret"
  plaintext_value  = "123"
  visibility       = "private"
}

resource "github_actions_organization_secret" "encrypted_secret" {
  secret_name      = "test_encrypted_secret"
  plaintext_value  = "123"
  visibility       = "private"
  destroy_on_drift = false
}
