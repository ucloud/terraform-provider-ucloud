package ipsecvpn

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

// Converter is use for converting string to another string with specifical style
type styleConverter interface {
	convertWithErr(string) (string, error)
	unconvertWithErr(string) (string, error)
	convert(string) string
	unconvert(string) string
}

type upperConverter struct{}

func newUpperConverter(specials map[string]string) styleConverter {
	return &upperConverter{}
}

// convert is an utils used for converting upper case name with underscore into lower case with underscore.
func (cvt *upperConverter) convertWithErr(input string) (string, error) {
	if input != strings.ToUpper(input) {
		return "", fmt.Errorf("excepted input string is uppercase with underscore, got %q", input)
	}
	return cvt.convert(input), nil
}

func (cvt *upperConverter) convert(input string) string {
	return strings.ToLower(input)
}

// unconvert is an utils used for converting lower case with underscore into upper case name with underscore.
func (cvt *upperConverter) unconvertWithErr(input string) (string, error) {
	if input != strings.ToLower(input) {
		return "", fmt.Errorf("excepted input string is lowercase with underscore, got %q", input)
	}
	return strings.ToUpper(input), nil
}

func (cvt *upperConverter) unconvert(input string) string {
	return strings.ToUpper(input)
}

type lowerCamelConverter struct{}

func newLowerCamelConverter(specials map[string]string) styleConverter {
	return &lowerCamelConverter{}
}

func (cvt *lowerCamelConverter) convertWithErr(input string) (string, error) {
	if len(input) == 0 {
		return "", nil
	}

	if 'A' <= input[0] && input[0] <= 'Z' {
		return "", fmt.Errorf("excepted lower camel should not be leading by uppercase character, got %q", input)
	}

	return lowerCamelToLower(input), nil
}

func (cvt *lowerCamelConverter) convert(input string) string {
	output, _ := cvt.convertWithErr(input)
	return output
}

func (cvt *lowerCamelConverter) unconvertWithErr(input string) (string, error) {
	if len(input) == 0 {
		return "", nil
	}

	if input != strings.ToLower(input) {
		return "", fmt.Errorf("excepted input string is lowercase with underscore, got %q", input)
	}

	return cvt.unconvert(input), nil
}

func (cvt *lowerCamelConverter) unconvert(input string) string {
	return lowerToLowerCamel(input)
}

type upperCamelConverter struct{}

func newUpperCamelConverter(specials map[string]string) styleConverter {
	return &upperCamelConverter{}
}

func (cvt *upperCamelConverter) convertWithErr(input string) (string, error) {
	if len(input) == 0 {
		return "", nil
	}

	if 'a' <= input[0] && input[0] <= 'z' {
		return "", fmt.Errorf("excepted upper camel should not be leading by lowercase character, got %q", input)
	}

	return lowerCamelToLower(strings.ToLower(input[:1]) + input[1:]), nil
}

func (cvt *upperCamelConverter) convert(input string) string {
	output, _ := cvt.convertWithErr(input)
	return output
}

func (cvt *upperCamelConverter) unconvertWithErr(input string) (string, error) {
	if len(input) == 0 {
		return "", nil
	}

	if input != strings.ToLower(input) {
		return "", fmt.Errorf("excepted input string is lowercase with underscore, got %q", input)
	}

	output := lowerToLowerCamel(input)
	return strings.ToUpper(output[:1]) + output[1:], nil
}

func (cvt *upperCamelConverter) unconvert(input string) string {
	output, _ := cvt.unconvertWithErr(input)
	return output
}

func lowerCamelToLower(input string) string {
	// eg. createFail -> create_fail; createUDBFAIL -> create_udb_fail -> createUdbFail
	var state int
	var words []string
	buf := strings.Builder{}
	for i := 0; i < len(input); i++ {
		c, l1 := input[i], lookAhead(&input, i, 1)

		// last character
		if l1 == 0 {
			buf.Write(bytes.ToLower([]byte{c}))
			words = append(words, buf.String())
			buf.Reset()
			break
		}

		if state == 0 {
			if 'A' <= l1 && l1 <= 'Z' {
				// createing UDBInstance
				//         ^ ^
				//         | |
				//         c l1
				buf.WriteByte(c)
				state = 1

				words = append(words, buf.String())
				buf.Reset()
			} else {
				// createi ngUDBInstance
				//       ^ ^
				//       | |
				//       c l1
				buf.WriteByte(c)
			}

			continue
		}

		if state == 1 {
			if 'A' <= l1 && l1 <= 'Z' {
				// createingU DBInstance
				//          ^ ^
				//          | |
				//          c l1
				buf.WriteByte(c + ('a' - 'A'))
				state = 3
			} else {
				// createingI nstance
				//          ^ ^
				//          | |
				//          c l1
				buf.WriteByte(c + ('a' - 'A'))
				state = 0
			}

			continue
		}

		if state == 3 {
			if 'A' <= l1 && l1 <= 'Z' {
				// createingUD BInstance
				//           ^ ^
				//           | |
				//           c l1
				buf.WriteByte(c + ('a' - 'A'))
			} else {
				// createingUDBI nstance
				//             ^ ^
				//             | |
				//             c l1
				words = append(words, buf.String())
				buf.Reset()

				buf.WriteByte(c + ('a' - 'A'))
				state = 0
			}

			continue
		}
	}

	return strings.Join(words, "_")
}

func lowerToLowerCamel(input string) string {
	iL := strings.Split(input, "_")
	oL := make([]string, len(iL))
	for i, s := range iL {
		oL[i] = strings.Title(s)
	}
	output := strings.Join(oL, "")
	return strings.ToLower(output[:1]) + output[1:]
}

func lookAhead(input *string, index, forward int) byte {
	if len((*input)) <= index+forward {
		return 0
	}
	return (*input)[index+forward]
}

const defaultTag = "Default"

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

type boolConverter struct {
	c map[bool]string
	r map[string]bool
}

func newBoolConverter(input map[bool]string) boolConverter {
	reversed := make(map[string]bool)
	for key, value := range input {
		reversed[value] = key
	}
	return boolConverter{c: input, r: reversed}
}

func (converter boolConverter) convert(value bool) string {
	if converted, ok := converter.c[value]; ok {
		return converted
	}
	return "unknown"
}

func (converter boolConverter) unconvert(value string) bool {
	if converted, ok := converter.r[value]; ok {
		return converted
	}
	return false
}

type stringConverter struct {
	c map[string]string
	r map[string]string
}

func newStringConverter(input map[string]string) stringConverter {
	reversed := make(map[string]string)
	for key, value := range input {
		reversed[value] = key
	}
	return stringConverter{c: input, r: reversed}
}

func (converter stringConverter) convert(value string) string {
	if converted, ok := converter.c[value]; ok {
		return converted
	}
	return value
}

func (converter stringConverter) unconvert(value string) string {
	if converted, ok := converter.r[value]; ok {
		return converted
	}
	return value
}

var (
	upperCamelCvt = newUpperCamelConverter(nil)
	boolCamelCvt  = newBoolConverter(map[bool]string{
		true:  "Yes",
		false: "No",
	})
	vpnAutoCvt = newStringConverter(map[string]string{
		"Auto": "auto",
	})
	vpnDisableCvt = newStringConverter(map[string]string{
		"Disable": "disable",
	})
	vpnIkeVersionCvt = newStringConverter(map[string]string{
		"IKE V1": "ikev1",
		"IKE V2": "ikev2",
	})
	validateDuration = validation.IntBetween(0, 9)
	validateName     = validation.StringMatch(
		regexp.MustCompile(`^[A-Za-z0-9\p{Han}-_.]{1,63}$`),
		"expected value to be 1 - 63 characters and only support chinese, english, numbers, '-', '_', '.'",
	)
	validateTag = validation.StringMatch(
		regexp.MustCompile(`^[A-Za-z0-9\p{Han}-_.]{0,63}$`),
		"expected value to be 0 - 63 characters and only support chinese, english, numbers, '-', '_', '.'",
	)
	validateVPNPreSharedKey = validation.StringMatch(
		regexp.MustCompile(`^[A-Za-z0-9!@#$%^&*()_+-={}\[\]:,./'~]{1,128}$`),
		"expected value to be 1 - 128 characters and only support english, numbers, !@#$%^&*()_+-=[]:,./'~",
	)
)

func stateFuncTag(value interface{}) string {
	if len(value.(string)) == 0 {
		return defaultTag
	}
	return value.(string)
}

func notEmptyStringInSet(value string) bool {
	return value != ""
}

func schemaSetToStringSlice(value interface{}) []string {
	result := []string{}
	for _, item := range value.(*schema.Set).List() {
		result = append(result, item.(string))
	}
	return result
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

	ioutil.WriteFile(absPath, content, 422)
	return nil
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

func validateVpnAuto(value interface{}, key string) (warnings []string, errors []error) {
	input := value.(string)
	if strings.EqualFold(input, "auto") && input != "auto" {
		errors = append(errors, fmt.Errorf("%q is invalid, should set it as %q if you want to automatic identification, got %q", key, "auto", input))
	}
	if input == "" {
		errors = append(errors, fmt.Errorf("%q is invalid, can not be set as null string, got %q", key, input))
	}
	return warnings, errors
}

func validateCIDRBlock(value interface{}, key string) (warnings []string, errors []error) {
	cidr := value.(string)
	if _, err := parseUCloudCidrBlock(cidr); err != nil {
		errors = append(errors, fmt.Errorf("%q is invalid, excepted cidr network in one of 192.168.x.x/x, 172.x.x.x/x, 10.x.x.x/x, got %s, %s", key, cidr, err))
	}
	return warnings, errors
}

type cidrBlock struct {
	Network string
	Mask    int
}

func parseCidrBlock(value string) (*cidrBlock, error) {
	if strings.Contains(value, ":") {
		return nil, fmt.Errorf("ipv6 is not supported now")
	}

	_, network, err := net.ParseCIDR(value)
	if err != nil {
		return nil, fmt.Errorf("cidr block %q cannot be parsed, %s", value, err)
	}

	mask, _ := network.Mask.Size()
	return &cidrBlock{Network: network.IP.String(), Mask: mask}, nil
}

func parseStringToInt64(value string) int64 {
	result, _ := strconv.Atoi(value)
	return int64(result)
}

func parseUCloudCidrBlock(value string) (*cidrBlock, error) {
	cidr, err := parseCidrBlock(value)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(value, "/")
	network := parts[0]
	if network != cidr.Network {
		return nil, fmt.Errorf("should use network ip matched with net mask")
	}

	address := strings.Split(network, ".")
	a := parseStringToInt64(address[0])
	b := parseStringToInt64(address[1])
	c := parseStringToInt64(address[2])
	d := parseStringToInt64(address[3])

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
