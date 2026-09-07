package uphost

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

const (
	NotFound = "Notfound"

	defaultTag                = "Default"
	statusStopped             = "Stopped"
	instanceStatusInstallFail = "Install Fail"
	instanceStatusResizeFail  = "ResizeFail"
	uphostStatusStopping      = "Stopping"
	eipResourceTypeUPHost     = "upm"
)

type ProviderError struct {
	errorCode string
	message   string
}

func (e *ProviderError) Error() string {
	return fmt.Sprintf("[ERROR] Terraform UCloud Provider Error: Code: %s Message: %s", e.errorCode, e.message)
}

func (err *ProviderError) ErrorCode() string {
	return err.errorCode
}

func (err *ProviderError) Message() string {
	return err.message
}

func newNotFoundError(str string) error {
	return &ProviderError{
		errorCode: NotFound,
		message:   str,
	}
}

func getNotFoundMessage(product, id string) string {
	return fmt.Sprintf("the specified %s %s is not found", product, id)
}

func isNotFoundError(err error) bool {
	if e, ok := err.(*ProviderError); ok &&
		(e.ErrorCode() == NotFound || strings.Contains(strings.ToLower(e.Message()), NotFound)) {
		return true
	}

	return false
}

type stringConverter struct {
	values  map[string]string
	reverse map[string]string
}

func newStringConverter(values map[string]string) stringConverter {
	reverse := make(map[string]string)
	for key, value := range values {
		reverse[value] = key
	}
	return stringConverter{values: values, reverse: reverse}
}

func (converter stringConverter) convert(value string) string {
	if converted, ok := converter.values[value]; ok {
		return converted
	}
	return value
}

func (converter stringConverter) unconvert(value string) string {
	if converted, ok := converter.reverse[value]; ok {
		return converted
	}
	return value
}

var raidTypeCvt = newStringConverter(map[string]string{
	"Raid0":  "raid0",
	"Raid1":  "raid1",
	"Raid5":  "raid5",
	"Raid10": "raid10",
	"NoRaid": "no_raid",
})

var validateTag = validation.StringMatch(
	regexp.MustCompile(`^[A-Za-z0-9\p{Han}-_.]{0,63}$`),
	"expected value to be 0 - 63 characters and only support chinese, english, numbers, '-', '_', '.'",
)

type upperCaseConverter struct{}

func (upperCaseConverter) convert(value string) string {
	return strings.ToLower(value)
}

var upperCvt = upperCaseConverter{}

type upperCamelConverter struct{}

func (cvt upperCamelConverter) convertWithErr(input string) (string, error) {
	if len(input) == 0 {
		return "", nil
	}

	if 'a' <= input[0] && input[0] <= 'z' {
		return "", fmt.Errorf("excepted upper camel should not be leading by lowercase character, got %q", input)
	}

	return lowerCamelToLower(strings.ToLower(input[:1]) + input[1:]), nil
}

func (cvt upperCamelConverter) convert(input string) string {
	output, _ := cvt.convertWithErr(input)
	return output
}

func (cvt upperCamelConverter) unconvertWithErr(input string) (string, error) {
	if len(input) == 0 {
		return "", nil
	}

	if input != strings.ToLower(input) {
		return "", fmt.Errorf("excepted input string is lowercase with underscore, got %q", input)
	}

	output := lowerToLowerCamel(input)
	return strings.ToUpper(output[:1]) + output[1:], nil
}

func (cvt upperCamelConverter) unconvert(input string) string {
	output, _ := cvt.unconvertWithErr(input)
	return output
}

func lowerCamelToLower(input string) string {
	var state int
	var words []string
	buf := strings.Builder{}
	for i := 0; i < len(input); i++ {
		c, l1 := input[i], lookAhead(&input, i, 1)

		if l1 == 0 {
			buf.Write(bytes.ToLower([]byte{c}))
			words = append(words, buf.String())
			buf.Reset()
			break
		}

		if state == 0 {
			if 'A' <= l1 && l1 <= 'Z' {
				buf.WriteByte(c)
				state = 1

				words = append(words, buf.String())
				buf.Reset()
			} else {
				buf.WriteByte(c)
			}

			continue
		}

		if state == 1 {
			if 'A' <= l1 && l1 <= 'Z' {
				buf.WriteByte(c + ('a' - 'A'))
				state = 3
			} else {
				buf.WriteByte(c + ('a' - 'A'))
				state = 0
			}

			continue
		}

		if state == 3 {
			if 'A' <= l1 && l1 <= 'Z' {
				buf.WriteByte(c + ('a' - 'A'))
			} else {
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
	if len(*input) <= index+forward {
		return 0
	}
	return (*input)[index+forward]
}

var upperCamelCvt = upperCamelConverter{}

func stateFuncTag(v interface{}) string {
	if len(v.(string)) == 0 {
		return defaultTag
	}
	return v.(string)
}

func schemaSetToStringSlice(s interface{}) []string {
	vL := []string{}
	for _, v := range s.(*schema.Set).List() {
		vL = append(vL, v.(string))
	}
	return vL
}

func isStringIn(val string, availables []string) bool {
	for _, choice := range availables {
		if val == choice {
			return true
		}
	}
	return false
}

func hashStringArray(arr []string) string {
	var buf bytes.Buffer
	for _, s := range arr {
		buf.WriteString(fmt.Sprintf("%s-", s))
	}
	return fmt.Sprintf("%d", hashcode.String(buf.String()))
}

func getAbsPath(filePath string) (string, error) {
	if strings.HasPrefix(filePath, "~") {
		usr, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("get current user got an error: %#v", err)
		}

		if usr.HomeDir != "" {
			filePath = strings.Replace(filePath, "~", usr.HomeDir, 1)
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

	var bs []byte
	switch data.(type) {
	case string:
		bs = []byte(data.(string))
	default:
		bs, err = json.MarshalIndent(data, "", "\t")
		if err != nil {
			return fmt.Errorf("MarshalIndent data %#v and got an error: %#v", data, err)
		}
	}

	ioutil.WriteFile(absPath, bs, 422)
	return nil
}
