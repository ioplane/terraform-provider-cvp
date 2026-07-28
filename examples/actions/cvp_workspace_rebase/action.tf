# Invoke the rebase verb on a workspace.
action "cvp_workspace_rebase" "example" {
  config {
    workspace_id = cvp_workspace.example.workspace_id
  }
}
