package ucloud

import (
	"github.com/ucloud/ucloud-sdk-go/services/uads"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"log"
)

func (client *UCloudClient) describeUADSById(id string) (*uads.ServiceInfo, error) {
	req := client.uadsconn.NewDescribeNapServiceInfoRequest()
	req.ResourceId = ucloud.String(id)
	req.ProjectId = nil
	resp, err := client.uadsconn.DescribeNapServiceInfo(req)
	if err != nil {
		return nil, err
	}

	if resp == nil || len(resp.ServiceInfo) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("uads", id))
	}

	return &resp.ServiceInfo[0], nil
}

func (client *UCloudClient) describeUADSAllowedDomain(id string, domain string) (*uads.BlockAllowDomainEntry, error) {
	req := client.uadsconn.NewGetNapAllowListDomainRequest()
	req.ResourceId = ucloud.String(id)
	req.Domain = ucloud.String(domain)
	resp, err := client.uadsconn.GetNapAllowListDomain(req)
	if err != nil {
		return nil, err
	}

	if resp == nil || len(resp.DomainList) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("uads", id))
	}

	return &resp.DomainList[0], nil
}

func (client *UCloudClient) describeUADSBGPServiceIP(id string, ip string) (*uads.GameIpInfoTotal, error) {
	req := client.uadsconn.NewGetBGPServiceIPRequest()
	req.ResourceId = ucloud.String(id)
	req.BgpIP = ucloud.String(ip)
	resp, err := client.uadsconn.GetBGPServiceIP(req)
	if err != nil {
		return nil, err
	}

	log.Printf("%v", resp)

	if resp == nil || len(resp.GameIPInfo) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("uads", id))
	}

	return &resp.GameIPInfo[0], nil
}

func (client *UCloudClient) describeUADSBGPServiceFwdRule(id string, ruleIndex int) (*uads.BGPFwdRule, error) {
	req := client.uadsconn.NewGetBGPServiceFwdRuleRequest()
	req.ResourceId = ucloud.String(id)
	req.RuleIndex = ucloud.Int(ruleIndex)
	resp, err := client.uadsconn.GetBGPServiceFwdRule(req)
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.RuleInfo) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("uads", id))
	}
	return &resp.RuleInfo[0], nil
}

func (client *UCloudClient) describeUADSBGPServiceFwdRuleByIpPort(id string, ip string, port int) (*uads.BGPFwdRule, error) {
	limit := 10
	for offset := 0; ; offset += limit {
		req := client.uadsconn.NewGetBGPServiceFwdRuleRequest()
		req.ResourceId = ucloud.String(id)
		req.BgpIP = ucloud.String(ip)
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)
		resp, err := client.uadsconn.GetBGPServiceFwdRule(req)
		if err != nil {
			return nil, err
		}
		if resp == nil || len(resp.RuleInfo) < 1 {
			return nil, newNotFoundError(getNotFoundMessage("uads", id))
		}
		for _, rule := range resp.RuleInfo {
			if port == 0 {
				if rule.FwdType == "IP" {
					return &rule, nil
				}
			} else {
				if rule.BgpIPPort == port {
					return &rule, nil
				}
			}
		}
	}
}
