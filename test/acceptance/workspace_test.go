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

func TestAccWorkspace_basic(t *testing.T) {
	preCheck(t)

	const rn = "cvp_workspace.test"
	rid := runID(t)
	name := "tf-acc-" + rid + "-ws-basic"
	updated := name + " (updated)"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories,
		CheckDestroy:             testAccCheckWorkspaceDestroy(t),
		Steps: []resource.TestStep{
			{ // create
				Config: testAccWorkspaceConfig(name, "created by acceptance"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(rn, "display_name", name),
					resource.TestCheckResourceAttr(rn, "description", "created by acceptance"),
					resource.TestCheckResourceAttr(rn, "state", "WORKSPACE_STATE_PENDING"),
					resource.TestCheckResourceAttr(rn, "needs_build", "false"),
					resource.TestCheckResourceAttrSet(rn, "workspace_id"),
					resource.TestCheckResourceAttrSet(rn, "created_by"),
				),
			},
			{ // idempotency — no diff on replan
				Config: testAccWorkspaceConfig(name, "created by acceptance"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			{ // update display_name + description in place
				Config: testAccWorkspaceConfig(updated, "edited by acceptance"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(rn, "display_name", updated),
					resource.TestCheckResourceAttr(rn, "description", "edited by acceptance"),
				),
			},
			{ // import by workspace_id
				ResourceName:                         rn,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "workspace_id",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[rn]
					if !ok {
						return "", fmt.Errorf("resource %s not found in state", rn)
					}
					return rs.Primary.Attributes["workspace_id"], nil
				},
			},
		},
	})
}

func testAccWorkspaceConfig(displayName, description string) string {
	return providerConfig() + fmt.Sprintf(`
resource "cvp_workspace" "test" {
  display_name = %q
  description  = %q
}
`, displayName, description)
}

// testAccCheckWorkspaceDestroy asserts every cvp_workspace in state is gone from
// CVP after destroy — a leaked workspace fails the suite (docs/testing.md).
func testAccCheckWorkspaceDestroy(t *testing.T) resource.TestCheckFunc {
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
