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

func TestAccDbtCloudJobResource(t *testing.T) {

	jobName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	jobName2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	jobName3 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	environmentName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDbtCloudJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudJobResourceBasicConfig(jobName, projectName, environmentName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "name", jobName),
				),
			},
			// RENAME
			{
				Config: testAccDbtCloudJobResourceBasicConfig(jobName2, projectName, environmentName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "name", jobName2),
				),
			},
			// MODIFY
			{
				Config: testAccDbtCloudJobResourceFullConfig(jobName2, projectName, environmentName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "name", jobName2),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "dbt_version", DBT_CLOUD_VERSION),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "target_name", "test"),
					resource.TestCheckResourceAttr("dbtcloud_job.test_job", "timeout_seconds", "180"),
					resource.TestCheckResourceAttrSet("dbtcloud_job.test_job", "project_id"),
					resource.TestCheckResourceAttrSet("dbtcloud_job.test_job", "environment_id"),
					resource.TestCheckResourceAttrSet("dbtcloud_job.test_job", "is_active"),
					resource.TestCheckResourceAttrSet("dbtcloud_job.test_job", "num_threads"),
					resource.TestCheckResourceAttrSet("dbtcloud_job.test_job", "run_generate_sources"),
					resource.TestCheckResourceAttrSet("dbtcloud_job.test_job", "generate_docs"),
				),
			},
			// DEFERRING JOBS
			{
				Config: testAccDbtCloudJobResourceDeferringJobConfig(jobName, jobName2, jobName3, projectName, environmentName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job"),
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job_2"),
					testAccCheckDbtCloudJobExists("dbtcloud_job.test_job_3"),
					resource.TestCheckResourceAttrSet("dbtcloud_job.test_job_2", "deferring_job_id"),
					resource.TestCheckResourceAttrSet("dbtcloud_job.test_job_3", "self_deferring"),
				),
			},
			// IMPORT
			{
				ResourceName:            "dbtcloud_job.test_job",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccDbtCloudJobResourceBasicConfig(jobName, projectName, environmentName string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_job_project" {
    name = "%s"
}

resource "dbtcloud_environment" "test_job_environment" {
    project_id = dbtcloud_project.test_job_project.id
    name = "%s"
    dbt_version = "%s"
    type = "development"
}

resource "dbtcloud_job" "test_job" {
  name        = "%s"
  project_id = dbtcloud_project.test_job_project.id
  environment_id = dbtcloud_environment.test_job_environment.environment_id
  execute_steps = [
    "dbt test"
  ]
  triggers = {
    "github_webhook": false,
    "git_provider_webhook": false,
    "schedule": false,
    "custom_branch_only": false,
  }
}
`, projectName, environmentName, DBT_CLOUD_VERSION, jobName)
}

func testAccDbtCloudJobResourceFullConfig(jobName, projectName, environmentName string) string {
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_job_project" {
    name = "%s"
}

resource "dbtcloud_environment" "test_job_environment" {
    project_id = dbtcloud_project.test_job_project.id
    name = "%s"
    dbt_version = "%s"
    type = "development"
}

resource "dbtcloud_environment" "test_job_environment_new" {
    project_id = dbtcloud_project.test_job_project.id
    name = "DEPL %s"
    dbt_version = "%s"
    type = "deployment"
}

resource "dbtcloud_job" "test_job" {
  name        = "%s"
  project_id = dbtcloud_project.test_job_project.id
  environment_id = dbtcloud_environment.test_job_environment_new.environment_id
  dbt_version = "%s"
  execute_steps = [
    "dbt test"
  ]
  triggers = {
    "github_webhook": false,
    "git_provider_webhook": false,
    "schedule": true,
    "custom_branch_only": false,
  }
  is_active = true
  num_threads = 37
  target_name = "test"
  run_generate_sources = true
  generate_docs = true
  schedule_type = "every_day"
  schedule_hours = [9, 17]
  timeout_seconds = 180
}
`, projectName, environmentName, DBT_CLOUD_VERSION, environmentName, DBT_CLOUD_VERSION, jobName, DBT_CLOUD_VERSION)
}

func testAccDbtCloudJobResourceDeferringJobConfig(jobName, jobName2, jobName3, projectName, environmentName string, selfDeferring bool) string {
	deferParam := "deferring_job_id = dbtcloud_job.test_job.id"
	if selfDeferring {
		deferParam = "self_deferring = true"
	}
	return fmt.Sprintf(`
resource "dbtcloud_project" "test_job_project" {
    name = "%s"
}

resource "dbtcloud_environment" "test_job_environment_new" {
    project_id = dbtcloud_project.test_job_project.id
    name = "DEPL %s"
    dbt_version = "%s"
    type = "deployment"
}

resource "dbtcloud_job" "test_job" {
  name        = "%s"
  project_id = dbtcloud_project.test_job_project.id
  environment_id = dbtcloud_environment.test_job_environment_new.environment_id
  dbt_version = "%s"
  execute_steps = [
    "dbt test"
  ]
  triggers = {
    "github_webhook": false,
    "git_provider_webhook": false,
    "schedule": true,
    "custom_branch_only": false,
  }
  is_active = true
  num_threads = 37
  target_name = "test"
  run_generate_sources = true
  generate_docs = true
  schedule_type = "every_day"
  schedule_hours = [9, 17]
}

resource "dbtcloud_job" "test_job_2" {
  name        = "%s"
  project_id = dbtcloud_project.test_job_project.id
  environment_id = dbtcloud_environment.test_job_environment_new.environment_id
  execute_steps = [
    "dbt test"
  ]
  triggers = {
    "github_webhook": false,
    "git_provider_webhook": false,
    "schedule": false,
    "custom_branch_only": false,
  }
  %s
}

resource "dbtcloud_job" "test_job_3" {
	name        = "%s"
	project_id = dbtcloud_project.test_job_project.id
	environment_id = dbtcloud_environment.test_job_environment_new.environment_id
	execute_steps = [
	  "dbt test"
	]
	triggers = {
	  "github_webhook": false,
	  "git_provider_webhook": false,
	  "schedule": false,
	  "custom_branch_only": false,
	}
	self_deferring = true
  }
`, projectName, environmentName, DBT_CLOUD_VERSION, jobName, DBT_CLOUD_VERSION, jobName2, deferParam, jobName3)
}

func testAccCheckDbtCloudJobExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}
		apiClient := testAccProvider.Meta().(*dbt_cloud.Client)
		_, err := apiClient.GetJob(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudJobDestroy(s *terraform.State) error {
	apiClient := testAccProvider.Meta().(*dbt_cloud.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_job" {
			continue
		}
		_, err := apiClient.GetJob(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Job still exists")
		}
		notFoundErr := "resource-not-found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
