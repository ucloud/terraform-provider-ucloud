package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	providerucloud "github.com/terraform-providers/terraform-provider-ucloud/ucloud"
)

func main() {
	resourceName := flag.String("resource", "", "Terraform resource name")
	inputPath := flag.String("input", "", "input InstanceState JSON")
	outputPath := flag.String("output", "", "output normalized InstanceState JSON")
	flag.Parse()

	if *resourceName == "" || *inputPath == "" || *outputPath == "" {
		fmt.Fprintln(os.Stderr, "-resource, -input, and -output are required")
		os.Exit(2)
	}

	input, err := os.ReadFile(*inputPath)
	if err != nil {
		fail("read input state: %v", err)
	}
	var state terraform.InstanceState
	if err := json.Unmarshal(input, &state); err != nil {
		fail("decode input state: %v", err)
	}

	provider := providerucloud.Provider().(*schema.Provider)
	resource, ok := provider.ResourcesMap[*resourceName]
	if !ok {
		fail("resource %q is not registered", *resourceName)
	}
	normalized := resource.Data(&state).State()
	if normalized == nil {
		fail("resource %q dropped state %q", *resourceName, state.ID)
	}
	encoded, err := json.MarshalIndent(normalized, "", "  ")
	if err != nil {
		fail("encode normalized state: %v", err)
	}
	encoded = append(encoded, byte(10))
	if err := writePrivateFile(*outputPath, encoded); err != nil {
		fail("write normalized state: %v", err)
	}
}

func writePrivateFile(path string, content []byte) (err error) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
	}()

	// OpenFile does not change the mode of an existing file.
	if err := file.Chmod(0600); err != nil {
		return err
	}
	_, err = file.Write(content)
	return err
}

func fail(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
