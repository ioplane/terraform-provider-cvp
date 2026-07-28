resource "cvp_change_control" "deploy" {
  workspace_id        = cvp_workspace.bgp_as_bump.id
  description         = "Deploy leaf BGP AS bump"
  auto_approve        = false # human approves via UI or a separate CI step
  wait_for_execution  = true
  stage_execution     = "rolling"
  stop_on_failure     = true
  rollback_on_failure = true

  depends_on = [cvp_studio_inputs.rack3_bgp]
}
