---
page_title: "dbtcloud_service_token Resource - dbtcloud"
subcategory: ""
description: |-
  
---

# dbtcloud_service_token (Resource)




## Example Usage

```terraform
// use dbt_cloud_service_token instead of dbtcloud_service_token for the legacy resource names
// legacy names will be removed from 0.3 onwards

resource "dbtcloud_service_token" "test_service_token" {
  name = "Test Service Token"
  service_token_permissions {
    permission_set = "git_admin"
    all_projects   = true
  }
  service_token_permissions {
    permission_set = "job_admin"
    all_projects   = false
    project_id     = dbtcloud_project.test_project.id
  }
}

// permission_set accepts one of the following values:
// "account_admin","admin","database_admin","git_admin","team_admin","job_admin","job_viewer","analyst","developer","stakeholder","readonly","project_creator","account_viewer","metadata_only"
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Service token name

### Optional

- `service_token_permissions` (Block Set) Permissions set for the service token (see [below for nested schema](#nestedblock--service_token_permissions))
- `state` (Number) Service token state (1 is active, 2 is inactive)

### Read-Only

- `id` (String) The ID of this resource.
- `token_string` (String, Sensitive) Service token secret value (only accessible on creation))
- `uid` (String) Service token UID (part of the token)

<a id="nestedblock--service_token_permissions"></a>
### Nested Schema for `service_token_permissions`

Required:

- `all_projects` (Boolean) Whether or not to apply this permission to all projects for this service token
- `permission_set` (String) Set of permissions to apply

Optional:

- `project_id` (Number) Project ID to apply this permission to for this service token

## Import

Import is supported using the following syntax:

```shell
# Import using a group ID found in the URL or via the API.
terraform import dbtcloud_group.test_service_token "service_token_id"
terraform import dbtcloud_group.test_service_token 12345
```
