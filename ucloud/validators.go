package ucloud

import (
	"fmt"
	"net/url"
	"time"
)

func validateBaseUrl(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if _, err := url.Parse(value); err != nil {
		errors = append(errors, fmt.Errorf("%q is invalid, should be an valid ucloud base_url, got %q, parse error: %s", "base_url", value, err))
	}
	return
}

func validateAssumeRoleDuration(v interface{}, k string) (ws []string, errors []error) {
	_, err := time.ParseDuration(v.(string))
	if err != nil {
		errors = append(errors, fmt.Errorf("%q cannot be parsed as a duration: %w", k, err))
	}
	return
}
