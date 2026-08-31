package udisk

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func resourceUCloudDiskAttachmentMigrateState(
	v int, is *terraform.InstanceState, meta interface{},
) (*terraform.InstanceState, error) {
	switch v {
	case 0:
		return migrateUCloudDiskAttachmentV0toV1(is)
	default:
		return is, fmt.Errorf("unexpected schema version: %d", v)
	}
}

var associationPattern = regexp.MustCompile(`^([^$]+)#([^:]+):([^$]+)#(.+)$`)

func migrateUCloudDiskAttachmentV0toV1(is *terraform.InstanceState) (*terraform.InstanceState, error) {
	if is.Empty() || is.Attributes == nil {
		return is, nil
	}

	matched := associationPattern.FindStringSubmatch(is.ID)
	if len(matched) < 5 {
		return is, fmt.Errorf("invalid identity of association")
	}
	is.ID = fmt.Sprintf("%s:%s", matched[2], matched[4])
	is.Attributes["id"] = is.ID
	return is, nil
}
