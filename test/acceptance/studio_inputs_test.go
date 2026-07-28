//go:build acceptance

package acceptance

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccStudioInputs_basic drives cvp_studio_inputs through a workspace against
// a real studio. It targets the studio root path with a trivial value, so it is
// safe on any lab (the workspace is abandoned on destroy, never submitted).
// CVP_ACC_STUDIO_ID overrides the studio (default studio-date-time).
func TestAccStudioInputs_basic(t *testing.T) {
	preCheck(t)

	const rn = "cvp_studio_inputs.test"
	studioID := envOr("CVP_ACC_STUDIO_ID", "studio-date-time")
	rid := runID(t)
	wsName := "tf-acc-" + rid + "-si"

	create := `{}`
	update := `{"ntpServerResolver":null,"ntpSourceInterfaceResolver":null,"timezoneResolver":null}`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories,
		CheckDestroy:             testAccCheckStudioInputsDestroy(t),
		Steps: []resource.TestStep{
			{ // create at studio root
				Config: testAccStudioInputsConfig(wsName, studioID, "[]", create),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(rn, "studio_id", studioID),
					resource.TestCheckResourceAttr(rn, "inputs_json", create),
					resource.TestCheckResourceAttr(rn, "path.#", "0"),
					resource.TestCheckResourceAttrSet(rn, "workspace_id"),
					resource.TestCheckResourceAttrSet(rn, "id"),
				),
			},
			{ // idempotency
				Config: testAccStudioInputsConfig(wsName, studioID, "[]", create),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			{ // update inputs_json in place
				Config: testAccStudioInputsConfig(wsName, studioID, "[]", update),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(rn, "inputs_json", update),
				),
			},
			{ // import by composite id
				ResourceName:      rn,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccStudioInputsConfig(wsName, studioID, pathHCL, inputsJSON string) string {
	return providerConfig() + fmt.Sprintf(`
resource "cvp_workspace" "test" {
  display_name = %q
}

resource "cvp_studio_inputs" "test" {
  workspace_id = cvp_workspace.test.workspace_id
  studio_id    = %q
  path         = %s
  inputs_json  = %q
}
`, wsName, studioID, pathHCL, inputsJSON)
}

// testAccCheckStudioInputsDestroy asserts the enclosing workspace (and thus its
// inputs) is gone after destroy.
func testAccCheckStudioInputsDestroy(t *testing.T) resource.TestCheckFunc {
	t.Helper()
	return func(s *terraform.State) error {
		client := accClient(t)
		ctx := context.Background()
		for name, rs := range s.RootModule().Resources {
			if rs.Type != "cvp_workspace" {
				continue
			}
			id := rs.Primary.Attributes["workspace_id"]
			gone, err := workspaceGone(ctx, client, id)
			if err != nil {
				return fmt.Errorf("checking destroy of %s (%s): %w", name, id, err)
			}
			if !gone {
				return fmt.Errorf("workspace %s (%s) still exists after destroy", name, id)
			}
		}
		return nil
	}
}
