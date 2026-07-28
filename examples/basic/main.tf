terraform {
  # Actions (lifecycle.action_trigger) require Terraform >= 1.14 / OpenTofu with
  # action support. The cvp_workspace resource alone works on older core.
  required_version = ">= 1.14.0"
  required_providers {
    cvp = {
      source  = "ioplane/cvp"
      version = "~> 0.2"
    }
  }
}

provider "cvp" {
  endpoint    = "um-cvp01.infra4.dev:443"
  auth_method = "cert"
  cert_pem    = file("${path.module}/certs/user.crt")
  key_pem     = file("${path.module}/certs/user.key")
  ca_pem      = file("${path.module}/certs/ca.crt")
}

resource "cvp_workspace" "net_leaf_bgp_s1" {
  display_name = "[net] PR#4211 Update BGP AS on rack-3 leaves"
  description  = "Bump AS-65001 → AS-65010 across all rack-3 leaves."

  # Build the workspace after it is created or updated. Submit/approve stay
  # deliberate, out-of-band steps (separation of duties, design.md non-goals).
  lifecycle {
    action_trigger {
      events  = [after_create, after_update]
      actions = [action.cvp_workspace_build.net_leaf_bgp_s1]
    }
  }
}

action "cvp_workspace_build" "net_leaf_bgp_s1" {
  config {
    workspace_id = cvp_workspace.net_leaf_bgp_s1.workspace_id
  }
}

resource "cvp_studio_inputs" "net_leaf_bgp_s1_rack3" {
  workspace_id = cvp_workspace.net_leaf_bgp_s1.workspace_id
  studio_id    = "1dd135ac-e1f3-4dd2-9b57-2bf1dbc3fa86" # EVPN Services with ESI Support
  path         = ["tenants", "[name=default]", "vrfs", "[name=RED]", "lbBgp"]
  inputs_json = jsonencode({
    asn          = 65010
    routerId     = "10.0.0.1"
    updateSource = "Loopback0"
  })
}

resource "cvp_change_control" "net_leaf_bgp_s1_deploy" {
  workspace_id        = cvp_workspace.net_leaf_bgp_s1.workspace_id
  description         = "Deploy leaf BGP AS bump"
  auto_approve        = false # human approves via UI or separate step
  wait_for_execution  = true
  stage_execution     = "rolling"
  stop_on_failure     = true
  rollback_on_failure = true

  depends_on = [cvp_studio_inputs.net_leaf_bgp_s1_rack3]
}

output "workspace_id" {
  value = cvp_workspace.net_leaf_bgp_s1.workspace_id
}

output "cc_status" {
  value = cvp_change_control.net_leaf_bgp_s1_deploy.status
}
