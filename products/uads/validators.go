package uads

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

var validateDuration = validation.IntBetween(0, 9)

var validateAntiDDoSInstanceName = validation.StringMatch(
	regexp.MustCompile(`^[A-Za-z0-9\p{Han}-_]{1,63}$`),
	"expected value to be 6 - 63 characters and only support chinese, english, numbers, '-', '_'",
)

var validateToaID = validation.IntBetween(0, 254)
