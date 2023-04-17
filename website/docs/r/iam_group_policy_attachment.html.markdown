---
subcategory: "IAM"
layout: "ucloud"
page_title: "UCloud: ucloud_iam_group_policy_attachment"
description: |-
  Provides an IAM group policy attachment resource.
---

# ucloud_iam_group

Provides an IAM group policy attachment resource.

## Example Usage

```hcl
resource "ucloud_iam_group" "foo" {
	name  = "tf-acc-iam-group"
	comment = "comment"
}
resource "ucloud_iam_project" "foo" {
	name  = "tf-acc-iam-project"
}
resource "ucloud_iam_policy" "foo" {
	name  = "tf-acc-iam-policy"
	comment = "comment"
    policy = jsonencode({
      Version = "1"
      Statement = [
      {
        Action = [
          "*",
        ]
        Effect   = "Allow"
        Resource = ["*"]
      },
      ]
    })
	scope_type = "Project"
}
resource "ucloud_iam_group_policy_attachment" "foo" {
	group_name  = ucloud_iam_group.foo.name
	policy_urn = ucloud_iam_policy.foo.urn
	project_id = ucloud_iam_project.foo.id
}
```

## Argument Reference

The following arguments are supported:

* `group_name` - (Required, ForceNew) Name of the IAM user group.
* `policy_urn` - (Required, ForceNew) URN of the IAM policy, including user policy and system policy.
* `project_id` - (Optional, ForceNew) ID of the IAM project, which is the scope of the policy.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `create_time` - The time of creation of group policy attachment, formatted in RFC3339 time string.

## Import
IAM group policy attachment resource can be imported using `account/<group-name>/<policy-urn>` for account scope policy, or `project/<project-id>/<group-name>/<policy-urn>` for project scope policy, e.g.

```
$ terraform import ucloud_iam_group_policy_attachment.example project/org-xxx/test-group/ucs:iam::ucs:policy/AdministratorAccess
```