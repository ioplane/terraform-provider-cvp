# Invoke the rollback verb on a workspace.
action "cvp_workspace_rollback" "example" {
  config {
    workspace_id = cvp_workspace.example.workspace_id
  }
}
