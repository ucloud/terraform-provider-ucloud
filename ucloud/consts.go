package ucloud

import (
	"time"
)

// DefaultMaxRetries is default max retry attempts number
const DefaultMaxRetries = 3

// DefaultInSecure is a default value to enable https
const DefaultInSecure = false

// DefaultWaitInterval is the inteval to wait for state changed after resource is created
const DefaultWaitInterval = 10 * time.Second

// DefaultWaitMaxAttempts is the max attempts number to wait for state changed after resource is created
const DefaultWaitMaxAttempts = 10

// DefaultWaitIgnoreError is if it will ignore error during wait for state changed after resource is created
const DefaultWaitIgnoreError = false

//listenerStatus is used to tranform int to string for status after read lb listener
var listenerStatus transformer = map[int]string{
	0: "AllNormal",
	1: "PartNormal",
	2: "AllException",
}

//attachmentStatus is used to tranform int to string for status after read lb attachment
var attachmentStatus transformer = map[int]string{
	0: "NormalRunning",
	1: "ExceptionRunning",
}

//uhostMap is used to covert uhost to instance
var uhostMap converter = map[string]string{
	"instance": "uhost",
}

//uHostMap is used to covert UHost to instance
var uHostMap converter = map[string]string{
	"instance": "UHost",
}

//uDiskMap is used to covert UDisk to Disk
var uDiskMap converter = map[string]string{
	"Disk": "UDisk",
}

//uDiskMap is used to covert Udisk to Disk
var udiskMap converter = map[string]string{
	"Disk": "Udisk",
}

//ulbMap is used to covert ulb to lb
var ulbMap converter = map[string]string{
	"lb": "ulb",
}
