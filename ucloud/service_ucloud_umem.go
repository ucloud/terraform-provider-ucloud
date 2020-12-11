package ucloud

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	pumem "github.com/ucloud/ucloud-sdk-go/private/services/umem"
	"github.com/ucloud/ucloud-sdk-go/services/umem"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func (c *UCloudClient) describeActiveStandbyRedisById(id string) (*umem.URedisGroupSet, error) {
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

	if resp == nil || len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("redis", id))
	}

	return &resp.DataSet[0], nil
}

func (c *UCloudClient) describeDistributedRedisById(id string) (*umem.UMemSpaceSet, error) {
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

	if resp == nil || len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("redis", id))
	}

	return &resp.DataSet[0], nil
}

func (c *UCloudClient) describeActiveStandbyMemcacheById(id string) (*pumem.UMemDataSet, error) {
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

	if resp == nil || len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("memcache", id))
	}

	return &resp.DataSet[0], nil
}

func waitForMemoryInstance(refresh func() (interface{}, string, error)) error {
	conf := resource.StateChangeConf{
		Timeout:    10 * time.Minute,
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
