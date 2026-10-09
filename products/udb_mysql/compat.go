package udb_mysql

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
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const (
	defaultTag             = "Default"
	storageClassNvme       = "CLOUD_RSSD"
	specificationClassNvme = "O"

	defaultPasswordNum = "012346789"
	defaultPasswordStr = "abcdefghijklmnopqrstuvwxyz"
	defaultPasswordSpe = "-_"

	// lifecycle states reported by DescribeUDBInstance
	mysqlStateInit         = "Init"
	mysqlStateStarting     = "Starting"
	mysqlStateRunning      = "Running"
	mysqlStateShutdown     = "Shutdown"
	mysqlStateShutoff      = "Shutoff"
	mysqlStateFail         = "Fail"
	mysqlStateDeleting     = "Deleting"
	mysqlStateDeleteFail   = "DeleteFail"
	mysqlStateRecoverFail  = "RecoverFail"
	mysqlStateShutdownFail = "ShutdownFail"
	mysqlStateStartFail    = "StartFail"
	mysqlStateUpgradeFail  = "UpgradeFail"

	statusPending = "pending"
)

// dbVersionList is the supported DBTypeId values for the MySQL product.
var dbVersionList = []string{"mysql-5.7", "mysql-8.0", "mysql-8.4"}

// dbMachineTypeList is the supported nvme machine types (o.*). Storage class
// and specification class are fixed to the nvme values by this product.
var dbMachineTypeList = []string{
	"o.mysql2m.small",    // 1C2G
	"o.mysql2m.medium",   // 2C4G
	"o.mysql2m.xlarge",   // 4C8G
	"o.mysql2m.2xlarge",  // 8C16G
	"o.mysql2m.4xlarge",  // 16C32G
	"o.mysql2m.8xlarge",  // 32C64G
	"o.mysql2m.12xlarge", // 48C96G
	"o.mysql2m.16xlarge", // 64C128G
	"o.mysql4m.medium",   // 2C8G
	"o.mysql4m.xlarge",   // 4C16G
	"o.mysql4m.2xlarge",  // 8C32G
	"o.mysql4m.4xlarge",  // 16C64G
	"o.mysql4m.8xlarge",  // 32C128G
	"o.mysql4m.16xlarge", // 64C256G
	"o.mysql8m.medium",   // 2C16G
	"o.mysql8m.xlarge",   // 4C32G
	"o.mysql8m.2xlarge",  // 8C64G
	"o.mysql8m.4xlarge",  // 16C128G
	"o.mysql8m.8xlarge",  // 32C256G
	"o.mysql8m.16xlarge", // 64C512G
}

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

func isStringIn(value string, choices []string) bool {
	for _, choice := range choices {
		if value == choice {
			return true
		}
	}
	return false
}

func payloadString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// payloadInt decodes a JSON number (float64) or integer into an int. The SDK
// generic response unmarshals every number into float64.
func payloadInt(m map[string]interface{}, key string) int {
	switch v := m[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	}
	return 0
}

func timestampToString(timestamp int) string {
	if timestamp <= 0 {
		return ""
	}
	return time.Unix(int64(timestamp), 0).Format(time.RFC3339)
}

// normalizeInstanceMode maps the backend instance mode values onto the stable
// schema values Normal / HA.
func normalizeInstanceMode(mode string) string {
	switch strings.ToLower(mode) {
	case "ha":
		return "HA"
	case "normal", "sp", "standalone", "cn", "":
		return "Normal"
	default:
		return mode
	}
}

func schemaSetToStringSlice(value interface{}) []string {
	set, ok := value.(*schema.Set)
	if !ok {
		return nil
	}
	result := make([]string, 0, set.Len())
	for _, item := range set.List() {
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
	_ = os.Remove(absPath)
	return ioutil.WriteFile(absPath, content, 0644)
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
