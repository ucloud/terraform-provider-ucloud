package us3

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
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

const (
	defaultTag   = "Default"
	notFoundCode = "Notfound"
)

var (
	validateTag = validation.StringMatch(
		regexp.MustCompile(`^[A-Za-z0-9\p{Han}-_.]{0,63}$`),
		"expected value to be 0 - 63 characters and only support chinese, english, numbers, '-', '_', '.'",
	)
	validateUS3BucketName1 = validation.StringMatch(
		regexp.MustCompile(`^[a-z0-9][a-z0-9-]{4,62}[a-z0-9]$`),
		"expected value to be 3 - 63 characters and only support lowercase-letters, numbers, '-', and can not prefix with '-' or suffix with '-'",
	)
	validateUS3BucketName2 = validation.StringDoesNotMatch(
		regexp.MustCompile(`^(www).*`),
		"expected value not prefix with 'www'",
	)
	validateUS3BucketName3 = validation.StringDoesNotMatch(
		regexp.MustCompile(`^(cn-bj).*`),
		"expected value not prefix with 'cn-bj'",
	)
	validateUS3BucketName4 = validation.StringDoesNotMatch(
		regexp.MustCompile(`^(hk).*`),
		"expected value not prefix with 'hk'",
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

func stateFuncTag(value interface{}) string {
	if len(value.(string)) == 0 {
		return defaultTag
	}
	return value.(string)
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
	switch data.(type) {
	case string:
		content = []byte(data.(string))
	default:
		content, err = json.MarshalIndent(data, "", "\t")
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
