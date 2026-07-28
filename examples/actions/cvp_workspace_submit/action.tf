# Submit a workspace, creating change controls. Set force = true to submit
# despite build warnings (REQUEST_SUBMIT_FORCE).
action "cvp_workspace_submit" "example" {
  config {
    workspace_id = cvp_workspace.example.workspace_id
    force        = false
  }
}
