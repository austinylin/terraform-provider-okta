package okta

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/okta/okta-sdk-golang/okta/query"
)

func sweepGroupRules(client *testClient) error {
	var errorList []error
	// Should never need to deal with pagination
	rules, _, err := client.oktaClient.Group.ListRules(&query.Params{Limit: 300})
	if err != nil {
		return err
	}

	for _, s := range rules {
		if s.Status == "ACTIVE" {
			if _, err := client.oktaClient.Group.DeactivateRule(s.Id); err != nil {
				errorList = append(errorList, err)
				continue
			}
		}
		if _, err := client.oktaClient.Group.DeleteRule(s.Id, nil); err != nil {
			errorList = append(errorList, err)
		}

	}
	return condenseError(errorList)
}

func TestAccOktaGroupRuleCrud(t *testing.T) {
	ri := acctest.RandInt()
	resourceName := fmt.Sprintf("%s.test", groupRule)
	mgr := newFixtureManager("okta_group_rule")
	config := mgr.GetFixtures("basic.tf", ri, t)
	updatedConfig := mgr.GetFixtures("basic_updated.tf", ri, t)
	groupUpdate := mgr.GetFixtures("basic_group_update.tf", ri, t)
	deactivated := mgr.GetFixtures("basic_deactivated.tf", ri, t)
	name := buildResourceName(ri)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "status", "ACTIVE"),

					resource.TestCheckResourceAttr(resourceName, "expression_type", "urn:okta:expression:1.0"),
					resource.TestCheckResourceAttr(resourceName, "expression_value", "String.startsWith(user.articulateId,\"auth0|\")"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "status", "ACTIVE"),
				),
			},
			{
				Config: groupUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "status", "ACTIVE"),
				),
			},
			{
				Config: deactivated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "status", "INACTIVE"),
				),
			},
		},
	})
}

// Issue 162
func TestAccOktaGroupRuleIssue162(t *testing.T) {
	ri := acctest.RandInt()
	resourceName := fmt.Sprintf("%s.test", groupRule)
	mgr := newFixtureManager("okta_group_rule")
	config := mgr.GetFixtures("test_issue_162.tf", ri, t)
	name := buildResourceName(ri)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					ensureResourceExists(resourceName, groupRuleExist),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "status", "ACTIVE"),
				),
			},
		},
	})
}

func groupRuleExist(id string) (bool, error) {
	client := getOktaClientFromMetadata(testAccProvider.Meta())
	_, response, err := client.Group.GetRule(id)

	// We don't want to consider a 404 an error in some cases and thus the delineation
	if response.StatusCode == 404 {
		return false, nil
	}

	if err != nil {
		return false, responseErr(response, err)
	}

	return true, err
}
