package umem

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	pumem "github.com/ucloud/ucloud-sdk-go/private/services/umem"
	"github.com/ucloud/ucloud-sdk-go/services/umem"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func (c *productClient) describeActiveStandbyRedisById(id string) (*umem.URedisGroupSet, error) {
	if id == "" {
		return nil, newNotFoundError(getNotFoundMessage("redis", id))
	}
	conn := c.umemconn

	req := conn.NewDescribeURedisGroupRequest()
	req.GroupId = ucloud.String(id)

	resp, err := conn.DescribeURedisGroup(req)
	if err != nil {
		return nil, err
	}
	if resp != nil && resp.GetRetCode() != 0 {
		return nil, fmt.Errorf("error on reading redis %q, %s", id, resp.GetMessage())
	}
	if resp == nil || len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("redis", id))
	}

	return &resp.DataSet[0], nil
}

func (c *productClient) describeDistributedRedisById(id string) (*umem.UMemSpaceSet, error) {
	if id == "" {
		return nil, newNotFoundError(getNotFoundMessage("redis", id))
	}
	conn := c.umemconn

	req := conn.NewDescribeUMemSpaceRequest()
	req.SpaceId = ucloud.String(id)

	resp, err := conn.DescribeUMemSpace(req)
	if err != nil {
		return nil, err
	}
	if resp != nil && resp.GetRetCode() != 0 {
		return nil, fmt.Errorf("error on reading redis %q, %s", id, resp.GetMessage())
	}
	if resp == nil || len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("redis", id))
	}

	return &resp.DataSet[0], nil
}

// describeDistributedRedisBlockInfoById pulls the shard (block) information of a
// distributed redis instance. DescribeUMemBlockInfo requires Limit and Offset, so
// the shards are collected page by page. The returned read mode is the cluster level
// read/write splitting strategy reported alongside the shard data set.
func (c *productClient) describeDistributedRedisBlockInfoById(id, zone string) ([]umem.UMemBlockInfo, string, error) {
	if id == "" {
		return nil, "", newNotFoundError(getNotFoundMessage("redis", id))
	}
	conn := c.umemconn

	limit := 100
	offset := 0
	var blocks []umem.UMemBlockInfo
	var readMode string
	for {
		req := conn.NewDescribeUMemBlockInfoRequest()
		req.SpaceId = ucloud.String(id)
		req.Zone = ucloud.String(zone)
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)

		resp, err := conn.DescribeUMemBlockInfo(req)
		if err != nil {
			return nil, "", err
		}
		if resp != nil && resp.GetRetCode() != 0 {
			return nil, "", fmt.Errorf("error on reading redis block info %q, %s", id, resp.GetMessage())
		}
		if resp == nil {
			break
		}

		readMode = resp.ReadMode
		blocks = append(blocks, resp.DataSet...)

		if len(resp.DataSet) < limit {
			break
		}
		offset = offset + limit
	}

	return blocks, readMode, nil
}

// describeDistributedRedisProxyInfoById pulls all the proxy information of a
// distributed redis instance. DescribeUDRedisProxyInfo requires SpaceId and
// the instance zone (the backend rejects the call with "Missing Params
// [zone_id]" without it) and returns the full proxy data set in a single call.
func (c *productClient) describeDistributedRedisProxyInfoById(id, zone string) ([]umem.UDRedisProxyInfo, error) {
	if id == "" {
		return nil, newNotFoundError(getNotFoundMessage("redis", id))
	}
	conn := c.umemconn

	req := conn.NewDescribeUDRedisProxyInfoRequest()
	req.SpaceId = ucloud.String(id)
	req.Zone = ucloud.String(zone)

	resp, err := conn.DescribeUDRedisProxyInfo(req)
	if err != nil {
		return nil, err
	}
	if resp != nil && resp.GetRetCode() != 0 {
		return nil, fmt.Errorf("error on reading redis proxy info %q, %s", id, resp.GetMessage())
	}
	if resp == nil {
		return nil, nil
	}

	return resp.DataSet, nil
}

func (c *productClient) describeActiveStandbyMemcacheById(id string) (*pumem.UMemDataSet, error) {
	if id == "" {
		return nil, newNotFoundError(getNotFoundMessage("memcache", id))
	}

	req := c.pumemconn.NewDescribeUMemRequest()
	req.ResourceId = ucloud.String(id)
	req.Protocol = ucloud.String("memcache")

	resp, err := c.pumemconn.DescribeUMem(req)
	if err != nil {
		return nil, err
	}
	if resp != nil && resp.GetRetCode() != 0 {
		return nil, fmt.Errorf("error on reading memcache %q, %s", id, resp.GetMessage())
	}
	if resp == nil || len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("memcache", id))
	}

	return &resp.DataSet[0], nil
}

const (
	// memoryInstanceWaitTimeout is the wait budget for create/restart/rename
	// operations on UMem instances.
	memoryInstanceWaitTimeout = 10 * time.Minute
	// redisResizeWaitTimeout is the wait budget for capacity resizes. A resize is
	// an async data migration (distributed redis migrates shard slots) and has
	// been observed to exceed the former fixed 10m budget.
	redisResizeWaitTimeout = 60 * time.Minute
)

func waitForMemoryInstance(refresh func() (interface{}, string, error), timeout time.Duration) error {
	conf := resource.StateChangeConf{
		Timeout:    timeout,
		Delay:      3 * time.Second,
		MinTimeout: 2 * time.Second,
		Target:     []string{statusInitialized},
		Pending:    []string{statusPending},
		Refresh:    refresh,
	}

	_, err := conf.WaitForState()
	if err != nil {
		return err
	}

	return nil
}
