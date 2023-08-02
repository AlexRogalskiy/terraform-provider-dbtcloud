package resources_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDbtCloudNotificationResource(t *testing.T) {

	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDbtCloudNotificationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudNotificationResourceCreateNotifications(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudNotificationExists("dbtcloud_notification.test_notification_internal"),
					resource.TestCheckResourceAttr("dbtcloud_notification.test_notification_internal", "notification_type", "1"),
					resource.TestCheckResourceAttrSet("dbtcloud_notification.test_notification_internal", "on_success.0"),
					resource.TestCheckResourceAttrSet("dbtcloud_notification.test_notification_internal", "on_cancel.0"),
					resource.TestCheckResourceAttrSet("dbtcloud_notification.test_notification_internal", "on_cancel.1"),
					resource.TestCheckResourceAttrSet("dbtcloud_notification.test_notification_internal", "on_failure.0"),

					testAccCheckDbtCloudNotificationExists("dbtcloud_notification.test_notification_external"),
					resource.TestCheckResourceAttr("dbtcloud_notification.test_notification_external", "notification_type", "4"),
					resource.TestCheckResourceAttrSet("dbtcloud_notification.test_notification_external", "on_failure.0"),
					resource.TestCheckResourceAttr("dbtcloud_notification.test_notification_external", "external_email", "nomail@mail.com"),
				),
			},
			// MODIFY
			{
				Config: testAccDbtCloudNotificationResourceModifyNotifications(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudNotificationExists("dbtcloud_notification.test_notification_internal"),
					resource.TestCheckNoResourceAttr("dbtcloud_notification.test_notification_internal", "on_cancel.0"),

					testAccCheckDbtCloudNotificationExists("dbtcloud_notification.test_notification_external"),
					resource.TestCheckResourceAttrSet("dbtcloud_notification.test_notification_external", "on_cancel.0"),
				),
			},
			// IMPORT
			{
				ResourceName:            "dbtcloud_notification.test_notification_internal",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			{
				ResourceName:            "dbtcloud_notification.test_notification_external",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccDbtCloudNotificationResourceBasicConfig(projectName string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_notification_project" {
	name = "%s"
}
	
resource "dbtcloud_environment" "test_notification_environment" {
	project_id  = dbtcloud_project.test_notification_project.id
	name        = "Test Env Notification"
	dbt_version = "%s"
	type        = "development"
}
	
resource "dbtcloud_job" "test_notification_job_1" {
	name           = "Job 1 TF"
	project_id     = dbtcloud_project.test_notification_project.id
	environment_id = dbtcloud_environment.test_notification_environment.environment_id
	execute_steps = [
		"dbt compile"
	]
	triggers = {
		"github_webhook" : false,
		"git_provider_webhook" : false,
		"schedule" : false,
		"custom_branch_only" : false,
	}
}
	
resource "dbtcloud_job" "test_notification_job_2" {
	name           = "Job 2 TF"
	project_id     = dbtcloud_project.test_notification_project.id
	environment_id = dbtcloud_environment.test_notification_environment.environment_id
	execute_steps = [
		"dbt test"
	]
	triggers = {
		"github_webhook" : false,
		"git_provider_webhook" : false,
		"schedule" : false,
		"custom_branch_only" : false,
	}
}
`, projectName, DBT_CLOUD_VERSION)
}

func testAccDbtCloudNotificationResourceCreateNotifications(projectName string) string {

	notificationsConfig := `
resource "dbtcloud_notification" "test_notification_internal" {
	user_id           = 100
	on_success        = [dbtcloud_job.test_notification_job_1.id]
	on_failure        = [dbtcloud_job.test_notification_job_2.id]
	on_cancel         = [dbtcloud_job.test_notification_job_1.id, dbtcloud_job.test_notification_job_2.id]
	notification_type = 1
}
	
resource "dbtcloud_notification" "test_notification_external" {
	user_id           = 100
	on_failure        = [dbtcloud_job.test_notification_job_2.id]
	notification_type = 4
	external_email    = "nomail@mail.com"
}
`
	return testAccDbtCloudNotificationResourceBasicConfig(projectName) + "\n" + notificationsConfig
}

func testAccDbtCloudNotificationResourceModifyNotifications(projectName string) string {

	notificationsConfig := `
resource "dbtcloud_notification" "test_notification_internal" {
	user_id           = 100
	on_success        = [dbtcloud_job.test_notification_job_1.id]
	on_failure        = [dbtcloud_job.test_notification_job_2.id]
	on_cancel         = []
	notification_type = 1
}
	
resource "dbtcloud_notification" "test_notification_external" {
	user_id           = 100
	on_failure        = [dbtcloud_job.test_notification_job_2.id]
	on_cancel         = [dbtcloud_job.test_notification_job_1.id]
	notification_type = 4
	external_email    = "nomail@mail.com"
}
`
	return testAccDbtCloudNotificationResourceBasicConfig(projectName) + "\n" + notificationsConfig
}

func testAccCheckDbtCloudNotificationExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		apiClient := testAccProvider.Meta().(*dbt_cloud.Client)
		_, err := apiClient.GetNotification(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudNotificationDestroy(s *terraform.State) error {
	apiClient := testAccProvider.Meta().(*dbt_cloud.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_notification" {
			continue
		}
		_, err := apiClient.GetNotification(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Notification still exists")
		}
		notFoundErr := "resource-not-found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
