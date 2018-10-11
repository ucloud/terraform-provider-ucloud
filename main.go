package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/ucloud/terraform-provider-ucloud/ucloud"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: ucloud.Provider,
	})
}
