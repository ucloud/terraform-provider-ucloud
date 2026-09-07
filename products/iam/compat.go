package iam

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
)

const (
	iamStatusActive   = "Active"
	iamStatusInactive = "Inactive"
	iamStatusFrozen   = "Frozen"
	NotFound          = "Notfound"
)

type ProviderError struct {
	errorCode string
	message   string
}

func (err *ProviderError) Error() string {
	return fmt.Sprintf("[ERROR] Terraform UCloud Provider Error: Code: %s Message: %s", err.errorCode, err.message)
}

func (err *ProviderError) ErrorCode() string {
	return err.errorCode
}

func (err *ProviderError) Message() string {
	return err.message
}

func newNotFoundError(message string) error {
	return &ProviderError{errorCode: NotFound, message: message}
}

func getNotFoundMessage(product, id string) string {
	return fmt.Sprintf("the specified %s %s is not found", product, id)
}

func isNotFoundError(err error) bool {
	providerErr, ok := err.(*ProviderError)
	return ok && (providerErr.ErrorCode() == NotFound ||
		strings.Contains(strings.ToLower(providerErr.Message()), NotFound))
}

func schemaListToStringSlice(iface interface{}) []string {
	values := make([]string, 0)
	for _, value := range iface.([]interface{}) {
		values = append(values, value.(string))
	}
	return values
}

func interfaceSliceToStringSlice(iface []interface{}) []string {
	values := make([]string, 0, len(iface))
	for _, value := range iface {
		values = append(values, value.(string))
	}
	return values
}

func hashStringArray(values []string) string {
	var buffer bytes.Buffer
	for _, value := range values {
		buffer.WriteString(fmt.Sprintf("%s-", value))
	}
	return fmt.Sprintf("%d", hashcode.String(buffer.String()))
}

func hashString(value string) string {
	return fmt.Sprintf("%d", hashcode.String(value))
}

func timestampToString(timestamp int) string {
	return time.Unix(int64(timestamp), 0).Format(time.RFC3339)
}

func writeToFile(filePath string, data interface{}) error {
	absPath, err := getAbsPath(filePath)
	if err != nil {
		return err
	}

	_ = os.Remove(absPath)

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

	_ = ioutil.WriteFile(absPath, content, 422)
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
