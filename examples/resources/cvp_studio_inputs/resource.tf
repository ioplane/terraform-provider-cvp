resource "cvp_studio_inputs" "rack3_bgp" {
  workspace_id = cvp_workspace.bgp_as_bump.workspace_id
  studio_id    = "studio-l3ls" # studio id (slug); an empty path targets the studio root
  path         = ["tenants", "[name=default]", "vrfs", "[name=RED]", "lbBgp"]

  inputs_json = jsonencode({
    asn          = 65010
    routerId     = "10.0.0.1"
    updateSource = "Loopback0"
  })
}
