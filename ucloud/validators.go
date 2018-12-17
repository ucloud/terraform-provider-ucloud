package ucloud

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/validation"
)

func validateInstanceType(v interface{}, k string) (ws []string, errors []error) {
	instanceType := v.(string)

	_, err := parseInstanceType(instanceType)
	if err != nil {
		errors = append(errors, err)
	}

	return
}

var instancePasswordUpperPattern = regexp.MustCompile(`[A-Z]`)
var instancePasswordLowerPattern = regexp.MustCompile(`[a-z]`)
var instancePasswordNumPattern = regexp.MustCompile(`[0-9]`)
var instancePasswordSpecialPattern = regexp.MustCompile(`[` + "`" + `()~!@#$%^&*-+=_|{}\[\]:;'<>,.?/]`)
var instancePasswordPattern = regexp.MustCompile(`^[A-Za-z0-9` + "`" + `()~!@#$%^&*-+=_|{}\[\]:;'<>,.?/]{8,30}$`)

func validateInstancePassword(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !instancePasswordPattern.MatchString(value) {
		errors = append(errors, fmt.Errorf("%q is invalid, should have between 8-30 characters and any characters must be legal, got %q", k, value))
	}

	categoryCount := 0
	if instancePasswordUpperPattern.MatchString(value) {
		categoryCount++
	}

	if instancePasswordLowerPattern.MatchString(value) {
		categoryCount++
	}

	if instancePasswordNumPattern.MatchString(value) {
		categoryCount++
	}

	if instancePasswordSpecialPattern.MatchString(value) {
		categoryCount++
	}

	if categoryCount < 3 {
		errors = append(errors, fmt.Errorf("%q is invalid, should have least 3 items of capital letters, lower case letters, numbers and special characters, got %q", k, value))
	}

	return
}

func validateSecurityGroupPort(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	splited := strings.Split(value, "-")
	if len(splited) > 2 {
		errors = append(errors, fmt.Errorf("%q is invalid, should like a number or number1-number2, got %q", k, value))
	}

	fromPort, err := strconv.Atoi(splited[0])
	if err != nil {
		errors = append(errors, fmt.Errorf("%q is invalid, should like a number or number1-number2, got %q", k, value))
	}

	if fromPort < 1 || fromPort > 65535 {
		errors = append(errors, fmt.Errorf("%q is invalid, should between 1-65535, got %q", k, value))
	}

	if len(splited) == 1 {
		return
	}

	toPort, err := strconv.Atoi(splited[1])
	if err != nil {
		errors = append(errors, fmt.Errorf("%q is invalid, should like a number or number1-number2, got %q", k, value))
	}

	if toPort < 1 || toPort > 65535 {
		errors = append(errors, fmt.Errorf("%q is invalid, should between 1-65535, got %q", k, value))
	}

	if toPort <= fromPort {
		errors = append(errors, fmt.Errorf("%q is invalid, for number1|number2, number2 must be greater than number1, got %q", k, value))
	}

	return
}

func validateUCloudCidrBlock(v interface{}, k string) (ws []string, errors []error) {
	cidr := v.(string)

	_, err := parseUCloudCidrBlock(cidr)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q is invalid, should like 0.0.0.0/0, got error %s", k, err))
	}

	return
}

var validateDuration = validation.IntBetween(1, 9)

var validateDiskName = validation.StringMatch(
	regexp.MustCompile(`^[A-Za-z0-9\p{Han}-_]{6,63}$`),
	"expected value to be 6 - 63 characters and only support chinese, english, numbers, '-', '_'",
)

var validateName = validation.StringMatch(
	regexp.MustCompile(`^[A-Za-z0-9\p{Han}-_.]{1,63}$`),
	"expected value to be 1 - 63 characters and only support chinese, english, numbers, '-', '_', '.'",
)

var validateTag = validation.StringMatch(
	regexp.MustCompile(`^[A-Za-z0-9\p{Han}-_.]{0,63}$`),
	"expected value to be 0 - 63 characters and only support chinese, english, numbers, '-', '_', '.'",
)
