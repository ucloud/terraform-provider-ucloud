package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/providercompat"
	"github.com/terraform-providers/terraform-provider-ucloud/ucloud"
)

func main() {
	output := flag.String("output", "", "write the provider contract to this file instead of stdout")
	flag.Parse()

	provider, ok := ucloud.Provider().(*schema.Provider)
	if !ok {
		fmt.Fprintln(os.Stderr, "provider does not implement *schema.Provider")
		os.Exit(1)
	}
	contract, err := providercompat.Marshal(provider)
	if err != nil {
		fmt.Fprintf(os.Stderr, "build provider contract: %v\n", err)
		os.Exit(1)
	}

	if *output == "" {
		if _, err := os.Stdout.Write(contract); err != nil {
			fmt.Fprintf(os.Stderr, "write provider contract: %v\n", err)
			os.Exit(1)
		}
		return
	}
	if err := os.WriteFile(*output, contract, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write provider contract to %q: %v\n", *output, err)
		os.Exit(1)
	}
}
