resource "cvp_workspace" "bgp_as_bump" {
  display_name = "[net] PR#4211 Update BGP AS on rack-3 leaves"
  description  = "Bump AS-65001 -> AS-65010 across all rack-3 leaves."
  auto_build   = true
  auto_submit  = false
  auto_approve = false
}
