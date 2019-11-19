provider "aws" {
  region = "us-west-2"
  assume_role {
    role_arn = var.role_arn
  }
}

terraform {
  backend "s3" {
    workspace_key_prefix = "skycatch"
  }
}
