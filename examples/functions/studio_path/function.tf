# Build a cvp_studio_inputs path from ordered segments. String segments are
# group members / resolver ids; single-key objects become bracket key-notation.
# Requires Terraform >= 1.8.

output "vrf_description_path" {
  value = provider::cvp::studio_path(
    "tenants", { tenantName = "RED" },
    "vrfs", { vrfName = "RED-VRF" },
    "vrfDescription",
  )
  # => ["tenants","[tenantName=RED]","vrfs","[vrfName=RED-VRF]","vrfDescription"]
}

output "vlan_path" {
  value = provider::cvp::studio_path("vlans", { vlanId = 100 }, "name")
  # => ["vlans","[vlanId=100]","name"]
}

output "studio_root" {
  value = provider::cvp::studio_path()
  # => []
}
