variable "owner" {
  description = "GitHub owner used to configure the provider"
  type        = string
}

variable "github_token" {
  description = "GitHub access token used to configure the provider"
  type        = string
}

variable "individual_repo" {
  description = "Name of repo owned by an individual to create the issue in"
  type = string
}

variable "org_repo" {
  description = "Name of repo owned by an organization to create the issue in"
  type = string
}

variable "org" {
  description = "Name of an org that owns the repo specified in org_repo"
  type = string
}