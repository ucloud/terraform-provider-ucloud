package ufs

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

var validateDuration = validation.IntBetween(0, 9)

var validateUFSVolumeName = validation.StringMatch(
	regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9-_]{7,63}$`),
	"expected value to be 6 - 63 characters and only support english, numbers, '-', '_', and can not prefix with '-'",
)

func validateMod(num int) schema.SchemaValidateFunc {
	return func(value interface{}, key string) (warnings []string, errors []error) {
		if value.(int)%num != 0 {
			errors = append(errors, fmt.Errorf("expected %q to be multiple of 10, got %d", key, value))
		}
		return warnings, errors
	}
}

func validateAll(validators ...schema.SchemaValidateFunc) schema.SchemaValidateFunc {
	return func(value interface{}, key string) (warnings []string, errors []error) {
		for _, validator := range validators {
			validatorWarnings, validatorErrors := validator(value, key)
			warnings = append(warnings, validatorWarnings...)
			errors = append(errors, validatorErrors...)
		}
		return warnings, errors
	}
}
