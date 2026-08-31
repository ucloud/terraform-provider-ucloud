package vpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

const (
	defaultTag = "Default"

	statusPending     = "pending"
	statusInitialized = "initialized"

	eipResourceTypeNatGateway = "natgw"
)

var (
	validateName = validation.StringMatch(
		regexp.MustCompile(`^[A-Za-z0-9\p{Han}-_.]{1,63}$`),
		"expected value to be 1 - 63 characters and only support chinese, english, numbers, '-', '_', '.'",
	)
	validateNatGatewayName = validation.StringMatch(
		regexp.MustCompile(`^[A-Za-z0-9\p{Han}-_.]{6,63}$`),
		"expected value to be 1 - 63 characters and only support chinese, english, numbers, '-', '_', '.'",
	)
	validateTag = validation.StringMatch(
		regexp.MustCompile(`^[A-Za-z0-9\p{Han}-_.]{0,63}$`),
		"expected value to be 0 - 63 characters and only support chinese, english, numbers, '-', '_', '.'",
	)
)

type providerError struct {
	errorCode string
	message   string
}

func (err *providerError) Error() string {
	return fmt.Sprintf("[ERROR] Terraform UCloud Provider Error: Code: %s Message: %s", err.errorCode, err.message)
}

func newNotFoundError(message string) error {
	return &providerError{errorCode: "Notfound", message: message}
}

func getNotFoundMessage(product, id string) string {
	return fmt.Sprintf("the specified %s %s is not found", product, id)
}

func isNotFoundError(err error) bool {
	providerErr, ok := err.(*providerError)
	return ok && (providerErr.errorCode == "Notfound" || strings.Contains(strings.ToLower(providerErr.message), "notfound"))
}

type upperCaseConverter struct{}

func (upperCaseConverter) convert(value string) string {
	return strings.ToLower(value)
}

func (upperCaseConverter) unconvert(value string) string {
	return strings.ToUpper(value)
}

var upperCvt = upperCaseConverter{}

func stateFuncTag(value interface{}) string {
	if value.(string) == "" {
		return defaultTag
	}
	return value.(string)
}

func schemaSetToStringSlice(value interface{}) []string {
	result := make([]string, 0)
	for _, item := range value.(*schema.Set).List() {
		result = append(result, item.(string))
	}
	return result
}

func interfaceSliceToStringSlice(value []interface{}) []string {
	result := make([]string, 0)
	for _, item := range value {
		result = append(result, item.(string))
	}
	return result
}

func isStringIn(value string, choices []string) bool {
	for _, choice := range choices {
		if value == choice {
			return true
		}
	}
	return false
}

func hashStringArray(values []string) string {
	var buffer bytes.Buffer
	for _, value := range values {
		buffer.WriteString(fmt.Sprintf("%s-", value))
	}
	return fmt.Sprintf("%d", hashcode.String(buffer.String()))
}

func timestampToString(timestamp int) string {
	return time.Unix(int64(timestamp), 0).Format(time.RFC3339)
}

func getAbsPath(filePath string) (string, error) {
	if strings.HasPrefix(filePath, "~") {
		currentUser, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("get current user got an error: %#v", err)
		}
		if currentUser.HomeDir != "" {
			filePath = strings.Replace(filePath, "~", currentUser.HomeDir, 1)
		}
	}
	return filepath.Abs(filePath)
}

func writeToFile(filePath string, data interface{}) error {
	absPath, err := getAbsPath(filePath)
	if err != nil {
		return err
	}

	os.Remove(absPath)
	var content []byte
	switch value := data.(type) {
	case string:
		content = []byte(value)
	default:
		content, err = json.MarshalIndent(value, "", "\t")
		if err != nil {
			return fmt.Errorf("MarshalIndent data %#v and got an error: %#v", data, err)
		}
	}
	_ = ioutil.WriteFile(absPath, content, 422)
	return nil
}

type cidrBlock struct {
	Network string
	Mask    int
}

func parseCidrBlock(value string) (*cidrBlock, error) {
	if strings.Contains(value, ":") {
		return nil, fmt.Errorf("ipv6 is not supported now")
	}

	_, ipNet, err := net.ParseCIDR(value)
	if err != nil {
		return nil, fmt.Errorf("cidr block %q cannot be parsed, %s", value, err)
	}

	intMask, _ := ipNet.Mask.Size()
	return &cidrBlock{Network: ipNet.IP.String(), Mask: intMask}, nil
}

func parseStringToInt64(value string) int64 {
	result, _ := strconv.Atoi(value)
	return int64(result)
}

// parseUCloudCidrBlock applies the legacy VPC private-network constraints.
func parseUCloudCidrBlock(value string) (*cidrBlock, error) {
	cidr, err := parseCidrBlock(value)
	if err != nil {
		return nil, err
	}

	networkParts := strings.Split(value, "/")
	network := networkParts[0]
	if network != cidr.Network {
		return nil, fmt.Errorf("should use network ip matched with net mask")
	}

	ipParts := strings.Split(network, ".")
	a := parseStringToInt64(ipParts[0])
	b := parseStringToInt64(ipParts[1])
	c := parseStringToInt64(ipParts[2])
	d := parseStringToInt64(ipParts[3])

	if a == 192 && b == 168 && 16 <= cidr.Mask && cidr.Mask <= 29 && (((a<<24)+(b<<16)+(c<<8)+d)&(((1<<32)-1)>>uint(cidr.Mask))) == 0 {
		return cidr, nil
	}

	if a == 172 && b&0xf0 == 16 && 12 <= cidr.Mask && cidr.Mask <= 29 && (((a<<24)+(b<<16)+(c<<8)+d)&(((1<<32)-1)>>uint(cidr.Mask))) == 0 {
		return cidr, nil
	}

	if a == 10 && 8 <= cidr.Mask && cidr.Mask <= 29 && (((a<<24)+(b<<16)+(c<<8)+d)&(((1<<32)-1)>>uint(cidr.Mask))) == 0 {
		return cidr, nil
	}

	return nil, fmt.Errorf("invalid cidr network")
}

func validateCIDRBlock(value interface{}, key string) (warnings []string, errors []error) {
	cidr := value.(string)
	if _, err := parseUCloudCidrBlock(cidr); err != nil {
		errors = append(errors, fmt.Errorf("%q is invalid, excepted cidr network in one of 192.168.x.x/x, 172.x.x.x/x, 10.x.x.x/x, got %s, %s", key, cidr, err))
	}
	return warnings, errors
}

func validatePortRange(value interface{}, key string) (warnings []string, errors []error) {
	portRange := value.(string)
	split := strings.Split(portRange, "-")
	if len(split) > 2 {
		errors = append(errors, fmt.Errorf("%q is invalid, should like a number or number1-number2, got %q", key, portRange))
	}

	fromPort, err := strconv.Atoi(split[0])
	if err != nil {
		errors = append(errors, fmt.Errorf("%q is invalid, should like a number or number1-number2, got %q", key, portRange))
	}
	if fromPort < 1 || fromPort > 65535 {
		errors = append(errors, fmt.Errorf("%q is invalid, should between 1-65535, got %q", key, portRange))
	}

	if len(split) == 1 {
		return warnings, errors
	}

	toPort, err := strconv.Atoi(split[1])
	if err != nil {
		errors = append(errors, fmt.Errorf("%q is invalid, should like a number or number1-number2, got %q", key, portRange))
	}
	if toPort < 1 || toPort > 65535 {
		errors = append(errors, fmt.Errorf("%q is invalid, should between 1-65535, got %q", key, portRange))
	}
	if toPort <= fromPort {
		errors = append(errors, fmt.Errorf("%q is invalid, for number1|number2, number2 must be greater than number1, got %q", key, portRange))
	}

	return warnings, errors
}

func hashCIDR(value interface{}) int {
	cidr := value.(string)
	if _, _, err := net.ParseCIDR(cidr); err != nil {
		return 0
	}
	return hashcode.String(cidr)
}

type associationInfo struct {
	PrimaryType  string
	PrimaryId    string
	ResourceType string
	ResourceId   string
}

var associaPattern = regexp.MustCompile(`^([^$]+)#([^:]+):([^$]+)#(.+)$`)

func parseAssociationInfo(assocID string) (*associationInfo, error) {
	matched := associaPattern.FindStringSubmatch(assocID)
	if len(matched) < 5 {
		return nil, fmt.Errorf("invalid identity of association")
	}
	return &associationInfo{
		PrimaryType:  matched[1],
		PrimaryId:    matched[2],
		ResourceType: matched[3],
		ResourceId:   matched[4],
	}, nil
}
