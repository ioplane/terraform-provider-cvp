# Invoke the cancel_build verb on a workspace.
action "cvp_workspace_cancel_build" "example" {
  config {
    workspace_id = cvp_workspace.example.workspace_id
  }
}
