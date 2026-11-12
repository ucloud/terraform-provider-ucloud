package udb_mysql

import (
	"fmt"
	"time"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

// clientFromMeta returns the shared base SDK client used to issue generic
// invokes against the MySQL nvme control plane.
func clientFromMeta(meta interface{}) (*ucloud.Client, error) {
	runtime, ok := meta.(product.RuntimeV1)
	if !ok {
		return nil, fmt.Errorf("invalid provider runtime %T", meta)
	}

	client, err := runtime.ProductClient(Name, newClient)
	if err != nil {
		return nil, err
	}
	typed, ok := client.(*ucloud.Client)
	if !ok {
		return nil, fmt.Errorf("product client %q has unexpected type %T", Name, client)
	}
	return typed, nil
}

// newClient builds the generic-invoke capable SDK client. The MySQL nvme
// actions are not exposed by any typed service client, so we use the SDK base
// client and the generic request/response primitives.
func newClient(
	config *ucloud.Config,
	credential *auth.Credential,
	handlers []ucloud.HttpRequestHandler,
) interface{} {
	longTimeoutConfig := *config
	longTimeoutConfig.Timeout = 60 * time.Second
	client := ucloud.NewClient(&longTimeoutConfig, credential)
	for _, handler := range handlers {
		_ = client.AddHttpRequestHandler(handler)
	}
	return client
}

// basePayload returns the common request fields shared by every action. Zone
// is omitted when empty so list actions can span a region.
func basePayload(client *ucloud.Client, zone string) map[string]interface{} {
	cfg := client.GetConfig()
	payload := map[string]interface{}{
		"Region":    cfg.Region,
		"ProjectId": cfg.ProjectId,
	}
	if zone != "" {
		payload["Zone"] = zone
	}
	return payload
}

// invoke sends one Action via generic invoke and returns the decoded response
// payload, or an error when the request fails or RetCode is non-zero.
func invoke(client *ucloud.Client, action string, payload map[string]interface{}) (map[string]interface{}, error) {
	payload["Action"] = action
	req := client.NewGenericRequest()
	if err := req.SetPayload(payload); err != nil {
		return nil, fmt.Errorf("set payload: %w", err)
	}

	resp, err := client.GenericInvoke(req)
	if err != nil {
		return nil, err
	}
	if resp.GetRetCode() != 0 {
		return nil, fmt.Errorf("api %s ret code %d: %s", action, resp.GetRetCode(), resp.GetMessage())
	}
	return resp.GetPayload(), nil
}
