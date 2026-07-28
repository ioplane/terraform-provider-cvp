# Build a workspace automatically after it is created or updated.
resource "cvp_workspace" "example" {
  display_name = "example change"

  lifecycle {
    action_trigger {
      events  = [after_create, after_update]
      actions = [action.cvp_workspace_build.example]
    }
  }
}

action "cvp_workspace_build" "example" {
  config {
    workspace_id = cvp_workspace.example.workspace_id
  }
}
