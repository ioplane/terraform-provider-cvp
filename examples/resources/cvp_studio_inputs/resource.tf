resource "cvp_studio_inputs" "rack3_bgp" {
  workspace_id = cvp_workspace.bgp_as_bump.id
  studio_id    = "1dd135ac-e1f3-4dd2-9b57-2bf1dbc3fa86" # EVPN Services with ESI Support
  path         = ["tenants", "[name=default]", "vrfs", "[name=RED]", "lbBgp"]

  inputs_json = jsonencode({
    asn          = 65010
    routerId     = "10.0.0.1"
    updateSource = "Loopback0"
  })
}
