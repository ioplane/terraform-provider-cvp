# A cvp_workspace holds declarative state (name, description). The workflow
# verbs — build, submit, abandon, rebase, rollback — are provider *actions*
# (ADR 0006), invoked via lifecycle.action_trigger. Requires Terraform >= 1.14.

resource "cvp_workspace" "bgp_as_bump" {
  display_name = "[net] PR#4211 Update BGP AS on rack-3 leaves"
  description  = "Bump AS-65001 -> AS-65010 across all rack-3 leaves."

  # Build the workspace automatically after it is created or updated.
  lifecycle {
    action_trigger {
      events  = [after_create, after_update]
      actions = [action.cvp_workspace_build.bgp_as_bump]
    }
  }
}

# Start a build (arista.workspace.v1 REQUEST_START_BUILD).
action "cvp_workspace_build" "bgp_as_bump" {
  config {
    workspace_id = cvp_workspace.bgp_as_bump.workspace_id
  }
}

# Submit is left as a deliberate, human-triggered step (separation of duties):
# run `terraform apply` against this action, or wire it to a trigger in non-prod.
action "cvp_workspace_submit" "bgp_as_bump" {
  config {
    workspace_id = cvp_workspace.bgp_as_bump.workspace_id
    force        = false
  }
}
