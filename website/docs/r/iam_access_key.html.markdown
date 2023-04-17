---
subcategory: "IAM"
layout: "ucloud"
page_title: "UCloud: ucloud_iam_access_key"
description: |-
  Provides an IAM access key resource.
---

# ucloud_iam_access_key

Provides an IAM access key resource.

## Example Usage

```hcl
resource "ucloud_iam_user" "foo" {
	name = "tf-acc-iam-user"
	login_enable = false
	is_frozen = false
}
resource "ucloud_iam_access_key" "foo" {
	user_name = "${ucloud_iam_user.foo.name}"
}
```

## Argument Reference

The following arguments are supported:

* `user_name` - (Required, ForceNew) Name of the IAM user.
* `secret_file` - (Optional, ForceNew) The name of file that can save access key id and access key secret.
* `status` - (Optional) Status of access key. It must be `Active` or `Inactive`. Default value is `Active`.
* `pgp_key` - (Optional) Either a base-64 encoded PGP public key, or a keybase username in the form `keybase:some_person_that_exists`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The access key ID.
* `status` - The access key status.
* `secret` - The secret access key. Note that this will be written to the state file. Alternatively, you may supply a `pgp_key` instead, which will prevent the secret from being stored in plaintext.
* `key_fingerprint` - The fingerprint of the PGP key used to encrypt the secret
* `encrypted_secret` - The encrypted secret, base64 encoded. ~> NOTE: The encrypted secret may be decrypted using the command line, for example: `terraform output encrypted_secret | base64 --decode | keybase pgp decrypt`.


