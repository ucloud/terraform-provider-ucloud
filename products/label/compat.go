package label

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
)

const notFoundCode = "Notfound"

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

func interfaceSliceToStringSlice(values []interface{}) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value.(string))
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
		content, err = json.MarshalIndent(data, "", "\t")
		if err != nil {
			return fmt.Errorf("MarshalIndent data %#v and got an error: %#v", data, err)
		}
	}

	// Keep the legacy behavior: output-file write failures are intentionally ignored.
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

const UCloudLabelIDSeperator = "#"

func buildUCloudLabelID(key, value string) string {
	return key + UCloudLabelIDSeperator + value
}

func parseUCloudLabelID(id string) (string, string, error) {
	items := strings.Split(id, UCloudLabelIDSeperator)
	if len(items) != 2 {
		return "", "", fmt.Errorf("invalid label id: %s", id)
	}
	return items[0], items[1], nil
}

const UCloudLabelAttachmentIDSeperator = "#"

func buildUCloudLabelAttachmentID(key, value, resourceID string) string {
	return strings.Join([]string{key, value, resourceID}, UCloudLabelAttachmentIDSeperator)
}

func parseUCloudLabelAttachmentID(id string) (key string, value string, resourceID string, err error) {
	items := strings.Split(id, UCloudLabelAttachmentIDSeperator)
	if len(items) != 3 {
		return "", "", "", fmt.Errorf("invalid label attachment id: %s", id)
	}
	return items[0], items[1], items[2], nil
}
