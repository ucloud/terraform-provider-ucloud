package ucloud

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform/terraform"
)

var autoRenewPattern = regexp.MustCompile(`^instances\.\d+\.auto_renew$`)
var isBootPattern = regexp.MustCompile(`^instances\.\d+\.disk_set\.\d+\.is_boot$`)
var diskTypePattern = regexp.MustCompile(`^instances\.\d+\.disk_set\.\d+\.type$`)
var memoryPattern = regexp.MustCompile(`^instances\.\d+\.memory$`)
var chargeTypePattern = regexp.MustCompile(`^instances\.\d+\.charge_type$`)

func dataSourceUCloudInstancesMigrateState(
	v int, is *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
	switch v {
	case 0:
		return migrateUCloudInstancesV0toV1(is)
	default:
		return is, fmt.Errorf("unexpected schema version: %d", v)
	}
}

func migrateUCloudInstancesV0toV1(is *terraform.InstanceState) (*terraform.InstanceState, error) {
	if is.Empty() || is.Attributes == nil {
		return is, nil
	}

	for k, v := range is.Attributes {
		if autoRenewPattern.MatchString(k) {
			if v == "Yes" {
				is.Attributes[k] = "true"
			}

			if v == "No" {
				is.Attributes[k] = "false"
			}
		}

		if isBootPattern.MatchString(k) {
			if v == "Yes" {
				is.Attributes[k] = "true"
			}

			if v == "No" {
				is.Attributes[k] = "false"
			}
		}

		if diskTypePattern.MatchString(k) {
			is.Attributes[k] = upperCvt.convert(v)
		}

		if memoryPattern.MatchString(k) {
			if m, err := strconv.Atoi(v); err == nil {
				is.Attributes[k] = strconv.Itoa(m / 1024)
			}
		}

		if chargeTypePattern.MatchString(k) {
			is.Attributes[k] = upperCamelCvt.convert(v)
		}
	}

	return is, nil
}
