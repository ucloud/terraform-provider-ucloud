package udisk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
)

const (
	defaultTag = "Default"

	statusPending = "pending"
	statusRunning = "Running"
	statusStopped = "Stopped"

	diskStatusAvailable = "Available"
	diskStatusInUse     = "InUse"
	diskStatusAttaching = "Attaching"
	diskStatusDetaching = "Detaching"

	backupModeSnapshotService = "SnapshotService"
	snapshotStatusNormal      = "Normal"
	snapshotStatusFailed      = "Failed"

	instanceStatusInitializing = "Initializing"
	instanceStatusStarting     = "Starting"
	instanceStatusStopping     = "Stopping"
	instanceStatusRebooting    = "Rebooting"
	instanceStatusInstallFail  = "Install Fail"
	instanceStatusResizeFail   = "ResizeFail"
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

type stringConverter struct {
	values  map[string]string
	reverse map[string]string
}

func newStringConverter(values map[string]string) stringConverter {
	reverse := make(map[string]string, len(values))
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

type intConverter map[int]string

func (converter intConverter) convert(value int) string {
	if converted, ok := converter[value]; ok {
		return converted
	}
	return "unknown"
}

type boolConverter map[bool]string

func (converter boolConverter) convert(value bool) string {
	return converter[value]
}

var boolCamelCvt = boolConverter{true: "Yes", false: "No"}

var diskTypeCvt = newStringConverter(map[string]string{
	"DataDisk":      "data_disk",
	"SSDDataDisk":   "ssd_data_disk",
	"SystemDisk":    "system_disk",
	"SSDSystemDisk": "ssd_system_disk",
	"RSSDDataDisk":  "rssd_data_disk",
})

var snapshotDiskTypeCvt = intConverter{
	0: "data_disk",
	1: "system_disk",
	2: "ssd_data_disk",
	3: "ssd_system_disk",
	4: "rssd_data_disk",
}

var upperCamelCvt = newStringConverter(map[string]string{
	"Year":    "year",
	"Month":   "month",
	"Dynamic": "dynamic",
})

var (
	validateDuration = validation.IntBetween(0, 9)
	validateDiskName = validation.StringMatch(
		regexp.MustCompile(`^[A-Za-z0-9\p{Han}-_]{6,63}$`),
		"expected value to be 6 - 63 characters and only support chinese, english, numbers, '-', '_'",
	)
	validateTag = validation.StringMatch(
		regexp.MustCompile(`^[A-Za-z0-9\p{Han}-_.]{0,63}$`),
		"expected value to be 0 - 63 characters and only support chinese, english, numbers, '-', '_', '.'",
	)
)

func validateMod(num int) schema.SchemaValidateFunc {
	return func(value interface{}, key string) (warnings []string, errors []error) {
		if value.(int)%num != 0 {
			errors = append(errors, fmt.Errorf("expected %q to be multiple of 10, got %d", key, value.(int)))
		}
		return warnings, errors
	}
}

func validateAll(validators ...schema.SchemaValidateFunc) schema.SchemaValidateFunc {
	return func(value interface{}, key string) (warnings []string, errors []error) {
		for _, validator := range validators {
			validatorWarnings, validatorErrors := validator(value, key)
			warnings = append(warnings, validatorWarnings...)
			errors = append(errors, validatorErrors...)
		}
		return warnings, errors
	}
}

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
	if err := os.WriteFile(absPath, content, 0o644); err != nil {
		return err
	}
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

func waitInstanceReady(client *productClient, instanceID string, timeout time.Duration) (*uhost.UHostInstanceSet, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{statusPending, instanceStatusInitializing, instanceStatusStarting, instanceStatusStopping, instanceStatusRebooting},
		Target:     []string{statusRunning, statusStopped},
		Refresh:    instanceStateRefreshFunc(client, instanceID, ""),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	output, err := stateConf.WaitForState()
	if instance, ok := output.(*uhost.UHostInstanceSet); ok {
		return instance, err
	}
	return nil, err
}

func instanceStateRefreshFunc(client *productClient, instanceID, target string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		instance, err := client.describeInstanceById(instanceID)
		if err != nil {
			if isNotFoundError(err) {
				return nil, statusPending, nil
			}
			return nil, "", err
		}
		state := instance.State
		if target != "" && state != target {
			if state == instanceStatusResizeFail {
				return nil, "", fmt.Errorf("resizing instance failed")
			}
			if state == instanceStatusInstallFail {
				return nil, "", fmt.Errorf("install failed")
			}
			state = statusPending
		}
		return instance, state, nil
	}
}

func waitAndUpdateInstanceState(client *productClient, instanceID, state string, timeout time.Duration) error {
	instance, err := waitInstanceReady(client, instanceID, timeout)
	if err != nil {
		return fmt.Errorf("error on waiting instance reach a ready status %v", err)
	}
	if instance.State == state {
		return nil
	}
	var updateErr error
	switch state {
	case statusStopped:
		updateErr = client.stopInstanceByID(instance.UHostId)
	case statusRunning:
		updateErr = client.startInstanceByID(instance.UHostId)
	default:
		return fmt.Errorf("unsupported instance state %q", state)
	}
	if updateErr != nil {
		return updateErr
	}
	if _, err = waitInstanceReady(client, instanceID, timeout); err != nil {
		return fmt.Errorf("error on waiting instance reach a ready status %v", err)
	}
	return nil
}
