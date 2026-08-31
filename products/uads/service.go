package uads

import (
	"fmt"
	"log"

	sdkuads "github.com/ucloud/ucloud-sdk-go/services/uads"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func describeUADSById(client *sdkuads.UADSClient, id string) (*sdkuads.ServiceInfo, error) {
	req := client.NewDescribeNapServiceInfoRequest()
	req.ResourceId = ucloud.String(id)
	req.NapType = ucloud.Int(1)
	req.ProjectId = nil
	resp, err := client.DescribeNapServiceInfo(req)
	if err != nil {
		return nil, err
	}
	if resp != nil && resp.GetRetCode() != 0 {
		return nil, fmt.Errorf("error on reading uads %q, %s", id, resp.GetMessage())
	}
	if resp == nil || len(resp.ServiceInfo) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("uads", id))
	}

	return &resp.ServiceInfo[0], nil
}

func describeUADSAllowedDomain(client *sdkuads.UADSClient, id string, domain string) (*sdkuads.BlockAllowDomainEntry, error) {
	req := client.NewGetNapAllowListDomainRequest()
	req.ResourceId = ucloud.String(id)
	req.Domain = ucloud.String(domain)
	resp, err := client.GetNapAllowListDomain(req)
	if err != nil {
		return nil, err
	}
	if resp != nil && resp.GetRetCode() != 0 {
		return nil, fmt.Errorf("error on reading uads allowed domain %q, %s", id, resp.GetMessage())
	}
	if resp == nil || len(resp.DomainList) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("uads", id))
	}

	return &resp.DomainList[0], nil
}

func describeUADSBGPServiceIP(client *sdkuads.UADSClient, id string, ip string) (*sdkuads.GameIpInfoTotal, error) {
	req := client.NewGetBGPServiceIPRequest()
	req.ResourceId = ucloud.String(id)
	req.BgpIP = ucloud.String(ip)
	resp, err := client.GetBGPServiceIP(req)
	if err != nil {
		return nil, err
	}
	if resp != nil && resp.GetRetCode() != 0 {
		return nil, fmt.Errorf("error on reading uads bgp service ip %q, %s", id, resp.GetMessage())
	}
	log.Printf("%v", resp)

	if resp == nil || len(resp.GameIPInfo) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("uads", id))
	}

	return &resp.GameIPInfo[0], nil
}

func describeUADSBGPServiceFwdRule(client *sdkuads.UADSClient, id string, ruleIndex int) (*sdkuads.BGPFwdRule, error) {
	req := client.NewGetBGPServiceFwdRuleRequest()
	req.ResourceId = ucloud.String(id)
	req.RuleIndex = ucloud.Int(ruleIndex)
	resp, err := client.GetBGPServiceFwdRule(req)
	if err != nil {
		return nil, err
	}
	if resp != nil && resp.GetRetCode() != 0 {
		return nil, fmt.Errorf("error on reading uads bgp service fwd rule %q, %s", id, resp.GetMessage())
	}
	if resp == nil || len(resp.RuleInfo) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("uads", id))
	}
	return &resp.RuleInfo[0], nil
}

func describeUADSBGPServiceFwdRuleByIpPort(client *sdkuads.UADSClient, id string, ip string, port int) (*sdkuads.BGPFwdRule, error) {
	limit := 10
	for offset := 0; ; offset += limit {
		req := client.NewGetBGPServiceFwdRuleRequest()
		req.ResourceId = ucloud.String(id)
		req.BgpIP = ucloud.String(ip)
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)
		resp, err := client.GetBGPServiceFwdRule(req)
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
