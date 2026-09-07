package udpn

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

const (
	notFoundCode      = "Notfound"
	statusPending     = "pending"
	statusInitialized = "initialized"
)

var validateDuration = validation.IntBetween(0, 9)

type providerError struct {
	errorCode string
	message   string
}

func (err *providerError) Error() string {
	return fmt.Sprintf("[ERROR] Terraform UCloud Provider Error: Code: %s Message: %s", err.errorCode, err.message)
}

func newNotFoundError(message string) error {
	return &providerError{errorCode: notFoundCode, message: message}
}

func getNotFoundMessage(product, id string) string {
	return fmt.Sprintf("the specified %s %s is not found", product, id)
}

func isNotFoundError(err error) bool {
	providerErr, ok := err.(*providerError)
	return ok && (providerErr.errorCode == notFoundCode ||
		strings.Contains(strings.ToLower(providerErr.message), notFoundCode))
}

func timestampToString(timestamp int) string {
	return time.Unix(int64(timestamp), 0).Format(time.RFC3339)
}

func upperCamelConvert(input string) string {
	if input == "" {
		return ""
	}
	return lowerCamelToLower(strings.ToLower(input[:1]) + input[1:])
}

func upperCamelUnconvert(input string) string {
	if input == "" {
		return ""
	}
	output := lowerToLowerCamel(input)
	return strings.ToUpper(output[:1]) + output[1:]
}

func lowerCamelToLower(input string) string {
	var state int
	var words []string
	var buffer strings.Builder

	for i := 0; i < len(input); i++ {
		current, next := input[i], lookAhead(input, i, 1)
		if next == 0 {
			buffer.WriteByte(toLowerASCII(current))
			words = append(words, buffer.String())
			buffer.Reset()
			break
		}

		if state == 0 {
			if isUpperASCII(next) {
				buffer.WriteByte(current)
				state = 1
				words = append(words, buffer.String())
				buffer.Reset()
			} else {
				buffer.WriteByte(current)
			}
			continue
		}

		if state == 1 {
			buffer.WriteByte(toLowerASCII(current))
			if isUpperASCII(next) {
				state = 3
			} else {
				state = 0
			}
			continue
		}

		if isUpperASCII(next) {
			buffer.WriteByte(toLowerASCII(current))
			continue
		}
		words = append(words, buffer.String())
		buffer.Reset()
		buffer.WriteByte(toLowerASCII(current))
		state = 0
	}

	return strings.Join(words, "_")
}

func lowerToLowerCamel(input string) string {
	parts := strings.Split(input, "_")
	if len(parts) == 0 {
		return ""
	}
	var output strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		output.WriteString(strings.ToUpper(part[:1]))
		output.WriteString(part[1:])
	}
	result := output.String()
	if result == "" {
		return ""
	}
	return strings.ToLower(result[:1]) + result[1:]
}

func lookAhead(input string, index, forward int) byte {
	if len(input) <= index+forward {
		return 0
	}
	return input[index+forward]
}

func isUpperASCII(value byte) bool {
	return value >= 'A' && value <= 'Z'
}

func toLowerASCII(value byte) byte {
	if value >= 'A' && value <= 'Z' {
		return value + ('a' - 'A')
	}
	return value
}
