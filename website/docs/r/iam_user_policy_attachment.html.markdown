---
subcategory: "IAM"
layout: "ucloud"
page_title: "UCloud: ucloud_iam_user_policy_attachment"
description: |-
  Provides an IAM user membership resource.
---

# ucloud_iam_user_policy_attachment

Provides an IAM group membership resource.

## Example Usage

```hcl
resource "ucloud_iam_user" "foo" {
	name  = "tf-acc-iam-user"
	login_enable = false
	is_frozen = false
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
resource "ucloud_iam_user_policy_attachment" "foo" {
	user_name  = ucloud_iam_user.foo.name
	policy_urn = ucloud_iam_policy.foo.urn
	project_id = ucloud_iam_project.foo.id
}
```

## Argument Reference

The following arguments are supported:

* `user_name` - (Required, ForceNew) Name of the IAM user.
* `policy_urn` - (Required, ForceNew) URN of the IAM policy, including user policy and system policy.
* `project_id` - (Optional, ForceNew) ID of the IAM project, which is the scope of the policy.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `create_time` - The time of creation of group policy attachment, formatted in RFC3339 time string.

## Import
IAM user policy attachment can be imported using `account/<user-name>/<policy-urn>` for account scope policy, or `project/<project-id>/<user-name>/<policy-urn>` for project scope policy, e.g.

```
$ terraform import ucloud_iam_group_policy_attachment.example project/org-xxx/test-user/ucs:iam::ucs:policy/AdministratorAccess
```