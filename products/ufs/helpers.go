package ufs

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
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

const (
	defaultTag    = "Default"
	statusPending = "pending"
	statusDELETED = "DELETED"
)

func schemaSetToStringSlice(value interface{}) []string {
	set := value.(*schema.Set)
	values := make([]string, 0, set.Len())
	for _, item := range set.List() {
		values = append(values, item.(string))
	}
	return values
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

func writeToFile(filePath string, data interface{}) error {
	absPath, err := getAbsPath(filePath)
	if err != nil {
		return err
	}

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
	// Keep the legacy behavior: output-file write failures are intentionally
	// ignored after attempting to replace the target file.
	os.Remove(absPath)
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

func stateFuncTag(value interface{}) string {
	if len(value.(string)) == 0 {
		return defaultTag
	}
	return value.(string)
}

func chargeTypeToAPI(value string) string {
	if value == "" {
		return value
	}
	return strings.ToUpper(value[:1]) + value[1:]
}

var validateTag = validation.StringMatch(
	regexp.MustCompile(`^[A-Za-z0-9\p{Han}-_.]{0,63}$`),
	"expected value to be 0 - 63 characters and only support chinese, english, numbers, '-', '_', '.'",
)
