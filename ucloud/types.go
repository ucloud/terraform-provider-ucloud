package ucloud

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

type cidrBlock struct {
	Network string
	Mask    int
}

func parseCidrBlock(s string) (*cidrBlock, error) {
	if strings.Contains(s, ":") {
		return nil, fmt.Errorf("ipv6 is not supported now")
	}

	_, ipNet, err := net.ParseCIDR(s)
	if err != nil {
		return nil, fmt.Errorf("cidr block %q cannot be parsed, %s", s, err)
	}

	intMask, _ := ipNet.Mask.Size()
	cidr := cidrBlock{
		Network: ipNet.IP.String(),
		Mask:    intMask,
	}

	return &cidr, nil
}

func (c *cidrBlock) String() string {
	return fmt.Sprintf("%s/%v", c.Network, c.Mask)
}

type instanceType struct {
	CPU           int
	Memory        int
	HostType      string
	HostScaleType string
}

func parseInstanceType(s string) (*instanceType, error) {
	splited := strings.Split(s, "-")
	if len(splited) < 3 {
		return nil, fmt.Errorf("instance type is invalid, got %q", s)
	}

	if splited[1] == "customized" {
		return parseInstanceTypeByCustomize(splited...)
	}

	return parseInstanceTypeByNormal(splited...)
}

func (i *instanceType) String() string {
	if i.Iscustomized() {
		return fmt.Sprintf("%s-%s-%v-%v", i.HostType, i.HostScaleType, i.CPU, i.Memory)
	} else {
		return fmt.Sprintf("%s-%s-%v", i.HostType, i.HostScaleType, i.CPU)
	}
}

func (i *instanceType) Iscustomized() bool {
	return i.HostScaleType == "customized"
}

var instanceTypeScaleMap = map[string]int{
	"highcpu":  1 * 1024,
	"basic":    2 * 1024,
	"standard": 4 * 1024,
	"highmem":  8 * 1024,
}

var availableHostTypes = []string{"n"}

func parseInstanceTypeByCustomize(splited ...string) (*instanceType, error) {
	if len(splited) != 4 {
		return nil, fmt.Errorf("instance type is invalid, expected like n-customize-1-2")
	}

	hostType := splited[0]
	err := checkStringIn(hostType, availableHostTypes)
	if err != nil {
		return nil, err
	}

	hostScaleType := splited[1]

	cpu, err := strconv.Atoi(splited[2])
	if err != nil {
		return nil, fmt.Errorf("cpu count is invalid, please use a number")
	}

	if cpu < 1 || 32 < cpu {
		return nil, fmt.Errorf("cpu count is invalid, it must between 1 ~ 32")
	}

	memory, err := strconv.Atoi(splited[3])
	if err != nil {
		return nil, fmt.Errorf("memory count is invalid, please use a number")
	}

	if memory < 1 || 256 < memory {
		return nil, fmt.Errorf("memory count is invalid, it must between 1 ~ 128")
	}

	if memory/cpu == 1 || memory/cpu == 2 || memory/cpu == 4 || memory/cpu == 8 {
		return nil, fmt.Errorf("instance type is invalid, expected %q like %q,"+
			"the Type can be highcpu, basic, standard, highmem when the ratio of cpu to memory is 1:1, 1:2, 1:4, 1:8", "n-Type-CPU", "n-standard-1")
	}

	t := &instanceType{}
	t.HostType = hostType
	t.HostScaleType = hostScaleType
	t.CPU = cpu
	t.Memory = memory * 1024
	return t, nil
}

func parseInstanceTypeByNormal(splited ...string) (*instanceType, error) {
	if len(splited) != 3 {
		return nil, fmt.Errorf("instance type is invalid, expected like n-standard-1")
	}

	hostType := splited[0]
	err := checkStringIn(hostType, availableHostTypes)
	if err != nil {
		return nil, err
	}

	hostScaleType := splited[1]
	if scale, ok := instanceTypeScaleMap[hostScaleType]; !ok {
		return nil, fmt.Errorf("instance type is invalid, expected like n-standard-1")
	} else {
		cpu, err := strconv.Atoi(splited[2])
		if err != nil {
			return nil, fmt.Errorf("cpu count is invalid, please use a number")
		}

		if cpu < 1 || 32 < cpu {
			return nil, fmt.Errorf("cpu count is invalid, it must between 1 ~ 32")
		}

		memory := cpu * scale

		t := &instanceType{}
		t.HostType = hostType
		t.HostScaleType = hostScaleType
		t.CPU = cpu
		t.Memory = memory
		return t, nil
	}
}

type associationInfo struct {
	PrimaryType  string
	PrimaryId    string
	ResourceType string
	ResourceId   string
}

var associaPattern = regexp.MustCompile("^([^$]+)#([^:]+):([^$]+)#(.+)$")

// parseAssociationInfo to decode association identify as some useful information,
// such as "eip#eip-xxx:uhost#uhost-xxx" is owned by two resource in this association,
// other representation is invalid.
func parseAssociationInfo(assocId string) (*associationInfo, error) {
	matched := associaPattern.FindStringSubmatch(assocId)

	if len(matched) < 5 {
		return nil, fmt.Errorf("invalid identity of association")
	}

	return &associationInfo{
		PrimaryType:  matched[1],
		PrimaryId:    matched[2],
		ResourceType: matched[3],
		ResourceId:   matched[4],
	}, nil
}

type attachmentInfo struct {
	PrimaryId string
	SecondId  string
	ThirdId   string
}

var attachmentPattern = regexp.MustCompile("^([^:]+):(.+):(.+)$")

// parseAttachmentInfo to decode attachment identify as some useful information,
// such as "ssl#xxx:lb#xxx:listener#xxx" is owned by three related resource in this attachment,
// other representation is invalid.
func parseAttachmentInfo(attachId string) (*attachmentInfo, error) {
	matched := attachmentPattern.FindStringSubmatch(attachId)

	if len(matched) < 4 {
		return nil, fmt.Errorf("invalid identity of attachment")
	}

	return &attachmentInfo{
		PrimaryId: matched[1],
		SecondId:  matched[2],
		ThirdId:   matched[3],
	}, nil
}

type dbInstanceType struct {
	Engine string
	Type   string
	Memory int
}

var availableDBEngine = []string{"mysql", "percona"}
var availableDBTypes = []string{"ha"}
var availableDBMemory = []int{1, 2, 4, 6, 8, 12, 16, 24, 32, 48, 64, 96, 128}

func parseDBInstanceType(s string) (*dbInstanceType, error) {
	splited := strings.Split(s, "-")
	if len(splited) != 3 {
		return nil, fmt.Errorf("db instance type is invalid, got %q", s)
	}
	engine := splited[0]
	if err := checkStringIn(engine, availableDBEngine); err != nil {
		return nil, err
	}

	dbType := splited[1]
	if err := checkStringIn(dbType, availableDBTypes); err != nil {
		return nil, err
	}

	memory, err := strconv.Atoi(splited[2])
	if err != nil {
		return nil, err
	}

	if err := checkIntIn(memory, availableDBMemory); err != nil {
		return nil, err
	}

	t := &dbInstanceType{}
	t.Engine = engine
	t.Type = dbType
	t.Memory = memory

	return t, nil
}
