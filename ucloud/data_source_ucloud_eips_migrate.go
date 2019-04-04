package ucloud

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform/terraform"
)

var eipschargeTypePattern = regexp.MustCompile(`^eips\.\d+\.charge_type$`)
var eipschargeModePattern = regexp.MustCompile(`^eips\.\d+\.charge_mode$`)

func dataSourceUCloudEipsMigrateState(
	v int, is *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
	switch v {
	case 0:
		return migrateUCloudEipsV0toV1(is)
	default:
		return is, fmt.Errorf("unexpected schema version: %d", v)
	}
}

func migrateUCloudEipsV0toV1(is *terraform.InstanceState) (*terraform.InstanceState, error) {
	if is.Empty() || is.Attributes == nil {
		return is, nil
	}

	for k, v := range is.Attributes {

		if eipschargeTypePattern.MatchString(k) {
			is.Attributes[k] = upperCamelCvt.convert(v)
		}

		if eipschargeModePattern.MatchString(k) {
			is.Attributes[k] = upperCamelCvt.convert(v)
		}
	}

	return is, nil
}
