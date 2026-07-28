terraform {
  required_providers {
    cvp = {
      source  = "ioplane/cvp"
      version = "~> 0.1"
    }
  }
}

# Bearer token is recommended for CI/CD (rotatable). Never commit the token —
# source it from the environment / a secret manager.
provider "cvp" {
  endpoint    = "cvp.example.com:443"
  auth_method = "bearer"
  token       = var.cvp_token
}

variable "cvp_token" {
  type      = string
  sensitive = true
}
