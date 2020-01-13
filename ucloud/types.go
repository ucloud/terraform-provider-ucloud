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

func parseStringToInt64(str string) int64 {
	// skip error, because has been validated by parseCidrBlock
	result, _ := strconv.Atoi(str)
	return int64(result)
}

/*
parseUCloudCidrBlock will parse cidr with specific range constraints
cidr must contained by subnet as followed
	- 192.168.*.[8, 16, 24 ...]
	- 172.[16-32].*.[8, 16, 24 ...]
	- 10.*.*.[8, 16, 24 ...]
*/
func parseUCloudCidrBlock(s string) (*cidrBlock, error) {
	cidr, err := parseCidrBlock(s)
	if err != nil {
		return nil, err
	}

	n := strings.Split(s, "/")
	network, _ := n[0], n[1]

	// if user input "192.168.1.1/24", it should be "192.168.1.0/24" with net mask
	if network != cidr.Network {
		return nil, fmt.Errorf("should use network ip matched with net mask")
	}

	n = strings.Split(network, ".")

	a := parseStringToInt64(n[0])
	b := parseStringToInt64(n[1])
	c := parseStringToInt64(n[2])
	d := parseStringToInt64(n[3])

	// check 192.168.--------.-----000
	if a == 192 && b == 168 && 16 <= cidr.Mask && cidr.Mask <= 29 && (((a<<24)+(b<<16)+(c<<8)+d)&(((1<<32)-1)>>uint(cidr.Mask))) == 0 {
		return cidr, nil
	}

	// check 172.0001----.--------.-----000
	if a == 172 && b&0xf0 == 16 && 12 <= cidr.Mask && cidr.Mask <= 29 && (((a<<24)+(b<<16)+(c<<8)+d)&(((1<<32)-1)>>uint(cidr.Mask))) == 0 {
		return cidr, nil
	}

	// check 10.--------.--------.-----000
	if a == 10 && 8 <= cidr.Mask && cidr.Mask <= 29 && (((a<<24)+(b<<16)+(c<<8)+d)&(((1<<32)-1)>>uint(cidr.Mask))) == 0 {
		return cidr, nil
	}

	return nil, fmt.Errorf("invalid cidr network")
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
		return nil, fmt.Errorf("instance type is invalid, expected like n-customized-1-2")
	}

	hostType := splited[0]
	err := checkStringIn(hostType, availableHostTypes)
	if err != nil {
		return nil, fmt.Errorf("instance type is invalid, the host type of customized %q can only be %q", "instance_type", "n")
	}

	hostScaleType := splited[1]

	cpu, err := strconv.Atoi(splited[2])
	if err != nil {
		return nil, fmt.Errorf("cpu count is invalid, please use a number")
	}

	memory, err := strconv.Atoi(splited[3])
	if err != nil {
		return nil, fmt.Errorf("memory count is invalid, please use a number")
	}

	if cpu != 1 && (cpu%2) != 0 {
		return nil, fmt.Errorf("expected the number of cores of cpu must be divisible by 2 without a remainder (except single core), got %d", cpu)
	}

	if memory != 1 && (memory%2) != 0 {
		return nil, fmt.Errorf("expected the number of memory must be divisible by 2 without a remainder (except single memory), got %d", memory)
	}

	if cpu/memory > 2 || memory/cpu > 12 || (cpu/memory == 2 && cpu%memory != 0) || (memory/cpu == 12 && memory%cpu != 0) {
		return nil, fmt.Errorf("the ratio of cpu to memory should be range of 2:1 ~ 1:12, got %d:%d", cpu, memory)
	}

	if (memory/cpu == 1 || memory/cpu == 2 || memory/cpu == 4 || memory/cpu == 8) && memory%cpu == 0 {
		return nil, fmt.Errorf("instance type is invalid, expected %q like %q,"+
			"the Type can be highcpu, basic, standard, highmem when the ratio of cpu to memory is 1:1, 1:2, 1:4, 1:8", "n-Type-CPU", "n-standard-1")
	}

	if cpu < 1 || 32 < cpu {
		return nil, fmt.Errorf("expected cpu to be in the range (1 - 32), got %d", cpu)
	}

	if memory < 1 || 128 < memory {
		return nil, fmt.Errorf("expected memory to be in the range (1 - 128),got %d", memory)
	}

	t := &instanceType{}
	t.HostType = hostType
	t.HostScaleType = hostScaleType
	t.CPU = cpu
	t.Memory = memory * 1024
	return t, nil
}

var availableOutstandingCpu = []int{4, 8, 16, 32, 64}

func parseInstanceTypeByNormal(splited ...string) (*instanceType, error) {
	if len(splited) != 3 {
		return nil, fmt.Errorf("instance type is invalid, expected like n-standard-1")
	}

	hostType := splited[0]
	err := checkStringIn(hostType, []string{"n", "o"})
	if err != nil {
		return nil, fmt.Errorf("instance type is invalid, the host type of %q must be one of %#v", "instance_type", []string{"n", "o"})
	}

	hostScaleType := splited[1]

	if scale, ok := instanceTypeScaleMap[hostScaleType]; !ok {
		return nil, fmt.Errorf("instance type is invalid, expected like %q,"+
			"the Type can be highcpu, basic, standard, highmem when the ratio of cpu to memory is 1:1, 1:2, 1:4, 1:8, got %q ", "n-standard-1", hostScaleType)
	} else {
		cpu, err := strconv.Atoi(splited[2])
		if err != nil {
			return nil, fmt.Errorf("cpu count is invalid, please use a number")
		}

		if cpu != 1 && (cpu%2) != 0 {
			return nil, fmt.Errorf("expected the number of cores of cpu must be divisible by 2 without a remainder (except single core), got %d", cpu)
		}

		if hostType == "o" {
			if err := checkIntIn(cpu, availableOutstandingCpu); err != nil {
				return nil, fmt.Errorf("expected cpu of outstanding instancetype %q", err)
			}

			if hostScaleType == "highmem" && cpu == 64 {
				return nil, fmt.Errorf("this instance type %q is not supported, please refer to instance type document", "o-highmem-64")
			}
		} else {
			if hostScaleType == "highmem" && cpu > 16 {
				return nil, fmt.Errorf("expected cpu to be in the range (1 - 16) for normal highmem instance type, got %d", cpu)
			}

			if cpu < 1 || 32 < cpu {
				return nil, fmt.Errorf("expected cpu to be in the range (1 - 32) for normal instance type, got %d", cpu)
			}
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

type dbInstanceType struct {
	Engine string
	Type   string
	Memory int
}

var availableDBEngine = []string{"mysql", "percona"}
var availableDBTypes = []string{"ha"}
var availableDBMemory = []int{1, 2, 4, 6, 8, 12, 16, 24, 32, 48, 64, 96, 128, 192, 256, 320}

func parseDBInstanceType(s string) (*dbInstanceType, error) {
	splited := strings.Split(s, "-")
	if len(splited) != 3 {
		return nil, fmt.Errorf("db instance type is invalid, should like xx-xx-1, got %q", s)
	}
	engine := splited[0]
	if err := checkStringIn(engine, availableDBEngine); err != nil {
		return nil, fmt.Errorf("db instance type is invalid, the engine %s", err)
	}

	dbType := splited[1]
	if err := checkStringIn(dbType, availableDBTypes); err != nil {
		return nil, fmt.Errorf("db instance type is invalid, the type %s", err)
	}

	memory, err := strconv.Atoi(splited[2])
	if err != nil {
		return nil, fmt.Errorf("db instance type is invalid, the memory %s", err)
	}

	if err := checkIntIn(memory, availableDBMemory); err != nil {
		return nil, fmt.Errorf("db instance type is invalid, memory is out of range, %s", err)
	}

	t := &dbInstanceType{}
	t.Engine = engine
	t.Type = dbType
	t.Memory = memory

	return t, nil
}

type redisInstanceType struct {
	Engine string
	Type   string
	Memory int
}

var availableRedisType = []string{"master", "distributed"}
var availableMasterRedisMemory = []int{1, 2, 4, 6, 8, 12, 16, 24, 32}

func parseRedisInstanceType(s string) (*redisInstanceType, error) {
	splited := strings.Split(s, "-")
	if len(splited) != 3 {
		return nil, fmt.Errorf("redis instance type is invalid, should like redis-xx-1, got %s", s)
	}

	engine := splited[0]
	if engine != "redis" {
		return nil, fmt.Errorf("redis instance type is invalid, the engine of instance type must be %q", "redis")
	}

	t := splited[1]
	if err := checkStringIn(t, availableRedisType); err != nil {
		return nil, fmt.Errorf("redis instance type is invalid, the type of instance type  %s", err)
	}

	memory, err := strconv.Atoi(splited[2])
	if err != nil {
		return nil, fmt.Errorf("redis instance type is invalid, the memory of instance type %s", err)
	}

	if engine == "redis" && t == "master" {
		if err := checkIntIn(memory, availableMasterRedisMemory); err != nil {
			return nil, fmt.Errorf("redis instance type is invalid, memory of instance type is out of range, %s", err)
		}
	} else if engine == "redis" && t == "distributed" {
		if memory < 16 || 1000 < memory {
			return nil, fmt.Errorf("redis instance type is invalid, memory of instance type should between 16-1000, got %v", memory)
		}

		if memory%4 != 0 {
			return nil, fmt.Errorf("redis instance type is invalid, memory of instance type should multiple of 4, got %v", memory)
		}
	} else {
		return nil, fmt.Errorf("redis instance type is invalid, excepted redis.master/redis.distributed, got %q", fmt.Sprintf("%s.%s", engine, t))
	}

	return &redisInstanceType{
		Engine: engine,
		Type:   t,
		Memory: memory,
	}, nil
}

type memcacheInstanceType struct {
	Engine string
	Type   string
	Memory int
}

var availableMasterMemcacheMemory = []int{1, 2, 4, 8, 16, 24, 32}

func parseMemcacheInstanceType(s string) (*memcacheInstanceType, error) {
	splited := strings.Split(s, "-")
	if len(splited) != 3 {
		return nil, fmt.Errorf("memcache instance type is invalid, should like memcache-xx-1, got %s", s)
	}

	engine := splited[0]
	if engine != "memcache" {
		return nil, fmt.Errorf("memcache instance type is invalid, the engine of instance type must be %q", "memcache")
	}

	t := splited[1]
	if t != "master" {
		return nil, fmt.Errorf("memcache instance type is invalid, the type of instance type must be %q", "master")
	}

	memory, err := strconv.Atoi(splited[2])
	if err != nil {
		return nil, fmt.Errorf("memcache instance type is invalid, the memory of instance type %s", err)
	}

	if err := checkIntIn(memory, availableMasterMemcacheMemory); err != nil {
		return nil, fmt.Errorf("memcache instance type is invalid, memory of instance type is out of range, %s", err)
	}

	return &memcacheInstanceType{
		Engine: engine,
		Type:   t,
		Memory: memory,
	}, nil
}
