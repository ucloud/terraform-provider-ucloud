package ucloud

import (
	"time"
)

const (
	// defaultMaxRetries is default max retry attempts number
	defaultMaxRetries = 3

	// defaultInSecure is a default value to enable https
	defaultInSecure = false

	// defaultWaitInterval is the inteval to wait for state changed after resource is created
	defaultWaitInterval = 10 * time.Second

	// defaultWaitMaxAttempts is the max attempts number to wait for state changed after resource is created
	defaultWaitMaxAttempts = 10

	// defaultWaitIgnoreError is if it will ignore error during wait for state changed after resource is created
	defaultWaitIgnoreError = false

	// defaultBaseURL is the api endpoint for advanced usage
	defaultBaseURL = "https://api.ucloud.cn"

	// defaultTag is the default tag for all of resources
	defaultTag = "Default"
)

const (
	// statusPending is the general status when remote resource is not completed
	statusPending = "pending"

	// statusInitialized is the general status when remote resource is completed
	statusInitialized = "initialized"

	// statusRunning is the general status when remote resource is running
	statusRunning = "running"

	// statusStopped is the general status when remote resource is stopped
	statusStopped = "stopped"
)

// listenerStatusCvt is used to covert int to string for status after read lb listener
var listenerStatusCvt = newIntConverter(map[int]string{
	0: "allNormal",
	1: "partNormal",
	2: "allException",
})

// lbAttachmentStatusCvt is used to covert int to string for status after read lb attachment
var lbAttachmentStatusCvt = newIntConverter(map[int]string{
	0: "normalRunning",
	1: "exceptionRunning",
})

// lowerCaseProdCvt is used to covert one lower string to another lower string
var lowerCaseProdCvt = newStringConverter(map[string]string{
	"instance": "uhost",
	"lb":       "ulb",
})

// titleCaseProdCvt is used to covert one lower string to another string begin with uppercase letters
var titleCaseProdCvt = newStringConverter(map[string]string{
	"instance": "UHost",
	"lb":       "ULB",
})

// dbModeCvt is used to covert basic to Normal and convert ha to HA
var dbModeCvt = newStringConverter(map[string]string{
	"basic": "Normal",
	"ha":    "HA",
})

// backupTypeCvt is used to transform string to int for backup type when read db backups
var backupTypeCvt = newIntConverter(map[int]string{
	0: "automatic",
	1: "manual",
})

// pgValueTypeCvt is used to transform int to string for value type after read parameter groups
var pgValueTypeCvt = newIntConverter(map[int]string{
	0:  "unknown",
	10: "int",
	20: "string",
	30: "bool",
})

// boolCamelCvt is used to transform bool value to Yes/No
var boolCamelCvt = newBoolConverter(map[bool]string{
	true:  "Yes",
	false: "No",
})

// boolLowerCvt is used to transform bool value to yes/no
var boolLowerCvt = newBoolConverter(map[bool]string{
	true:  "yes",
	false: "no",
})

// boolValueCvt is used to transform bool value to True/False
var boolValueCvt = newBoolConverter(map[bool]string{
	true:  "True",
	false: "False",
})

var diskTypeCvt = newStringConverter(map[string]string{
	"DataDisk":    "data_disk",
	"SSDDataDisk": "ssd_data_disk",
})

// upperCvt is used to transform uppercase with underscore to lowercase with underscore. eg. LOCAL_SSD -> local_ssd
var upperCvt = newUpperConverter(nil)

// lowerCamelCvt is used to transform lower camel case to lowercase with underscore. eg. localSSD -> local_ssd
var lowerCamelCvt = newLowerCamelConverter(nil)

// upperCamelCvt is used to transform uppercamel case to lowercase with underscore. eg. LocalSSD -> local_ssd
var upperCamelCvt = newUpperCamelConverter(nil)
