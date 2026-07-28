# Invoke the abandon verb on a workspace.
action "cvp_workspace_abandon" "example" {
  config {
    workspace_id = cvp_workspace.example.workspace_id
  }
}
