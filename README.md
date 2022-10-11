Terraform Provider
==================

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

<a href="https://terraform.io">
    <img src=".github/terraform_logo.svg" width="600px">
</a>

Requirements
------------

- [Terraform](https://www.terraform.io/downloads.html) 0.13.x
- [Go](https://golang.org/doc/install) 1.18+ (to build the provider plugin)

Building The Provider
---------------------

Clone repository to: `/path/to/terraform-provider-ucloud`

```shell
git clone git@github.com:ucloud/terraform-provider-ucloud /path/to/terraform-provider-ucloud
```

Enter the provider directory and build the provider

```shell
cd /path/to/terraform-provider-ucloud
make build
```

Using the provider
----------------------

If you're building the provider, follow the instructions to [install it as a plugin.](https://www.terraform.io/docs/plugins/basics.html#installing-a-plugin) After placing it into your plugins directory,  run `terraform init` to initialize it.

Developing the Provider
---------------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.18+ is *required*).

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `bin` directory.

```sh
make build
./bin/terraform-provider-ucloud
```

In order to test the provider, you can simply run `make test`.

```sh
make test
```

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
make testacc
```

Replace the default binary
---------------------------

You can replace the default provider with the compiled dev binary, first run `make dev`:

```shell
make dev
```

Then, create a terraform config file in `~/.terraformrc`, add the following content:

```hcl
provider_installation {

  dev_overrides {
    "ucloud/ucloud" = "{your-home}/.terraform.d/plugins"
  }

  direct {}
}
```

Now, the `ucloud` provider will be replaced to your compiled binary, when you executing `terraform` command, you will see the following message:

```
| Warning: Provider development overrides are in effect
│
│ The following provider development overrides are set in the CLI configuration:
│  - ucloud/ucloud in {your-home}/.terraform.d/plugins
│
│ The behavior may therefore not match any released version of the provider and applying changes may cause the state to become
│ incompatible with published releases.
```

After development done, you should clean the `~/.terraformrc` file.

## Acceptance Testing

Before making a release, the resources and data sources are tested automatically with acceptance tests (the tests are located in the `ucloud/*_test.go` files).

You can run them by entering the following instructions in a terminal:

```
cd /path/to/terraform-provider-ucloud
export UCLOUD_PUBLIC_KEY=xxx
export UCLOUD_PRIVATE_KEY=xxx
export UCLOUD_REGION=xxx
export UCLOUD_PROJECT_ID=xxx
TF_ACC=1 TF_LOG=INFO go test ./ucloud -v -run="^TestAccUCloud" -timeout=1440m
```

## Reference

UCloud Provider [Official Docs](https://www.terraform.io/docs/providers/ucloud/index.html)
